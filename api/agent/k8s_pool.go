package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	pool "github.com/lean-mu/mu/api/runnerpool"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type k8sRunnerPool struct {
	client *kubernetes.Clientset
	// Handler to cancel the infinite loop
	cancel chan struct{}
	// mutex
	nodeLock sync.RWMutex
	// The actual runners, use runners[0].Address() to get the IP
	runners map[string]pool.Runner
}

func K8sDynamicRunnerPool(headlessServiceName string) pool.RunnerPool {
	return NewK8sDynamicRunnerPool(headlessServiceName, nil)
}

func NewK8sDynamicRunnerPool(headlessServiceName string, tlsConf *tls.Config, dialOpts ...grpc.DialOption) pool.RunnerPool {
	logrus.WithField("headlessServiceName", headlessServiceName).Info("Starting k8s dynamic runner pool")

	client, err := k8sClient()
	if err != nil {
		logrus.WithError(err).Error("Can't connect to k8s")
		return nil
	}

	ns, err := namespace()
	if err != nil {
		logrus.WithError(err).Error("Can't get namespace from k8s")
		return nil
	}

	logrus.WithField("namespace", ns).Info("current NS")

	dialOpts = append(dialOpts, grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))

	thePool := &k8sRunnerPool{
		client:  client,
		cancel:  make(chan struct{}),
		runners: map[string]pool.Runner{},
	}

	// Start watching the control plane for changes on endpoint
	go thePool.watch(ns, headlessServiceName, tlsConf, dialOpts)

	return thePool
}


// converts from map[string]pool.Runner to []pool.Runner
func (rp *k8sRunnerPool) Runners(ctx context.Context, call pool.RunnerCall) ([]pool.Runner, error) {

	rp.nodeLock.RLock()
	defer rp.nodeLock.RUnlock()

	logrus.Debugf("k8s_pool.len(rp.runners)=%d", len(rp.runners))
	nodes := make([]pool.Runner, 0, len(rp.runners))
	logrus.Debugf("len %d nodes", len(nodes))
	for key, runner := range rp.runners {
		logrus.Debugf("adding index %s nodes", key)
		nodes = append(nodes, runner)
	}

	logrus.Debugf("k8s_pool.Runners called - return %d nodes", len(nodes))
	return nodes, nil
}

func (rp *k8sRunnerPool) Shutdown(ctx context.Context) error {

	// Stop the process that monitors endpoints
	close(rp.cancel)

	// Now close each runner
	var retErr error
	for addr, r := range rp.runners {
		err := r.Close(ctx)
		if err != nil {
			logrus.WithError(err).WithField("runner_addr", addr).Error("Error closing runner")
			// Grab the first error only for now.
			if retErr == nil {
				retErr = err
			}
		}
	}

	return retErr
}

// Returns the list of runner's IP
func (rp *k8sRunnerPool) List() ([]string, error) {
	var nodes []string

	rp.nodeLock.RLock()
	defer rp.nodeLock.RUnlock()

	logrus.WithField("len(rp.runners)", len(rp.runners)).Warn("List - new runners size")
	for key, _ := range rp.runners {
		logrus.WithField("key", key).Warn("append ")
		nodes = append(nodes, key)
	}

	return nodes, nil
}

// Watch the headless service
func (rp *k8sRunnerPool) watch(ns string, headlessServiceName string, tlsConf *tls.Config, dialOpts []grpc.DialOption) {

	ctx := context.Background()

	// Try to reconnect 4 times
	attempts := 4
	// Wait 2s before retrying
	sleep, _ := time.ParseDuration("2s")

	var epWatch watch.Interface
	var err error

	for i := 0; i < attempts; i++ {
		if i > 0 {
			logrus.
				WithFields(logrus.Fields{
					"namespace": ns,
					"endpoint":  headlessServiceName,
					"error":     err,
				}).Warn("Can't create watch() on endpoint.")
			logrus.Infof("Retrying in %.0f seconds...", sleep.Seconds())
			time.Sleep(sleep)
			sleep *= 2
		}

		epWatch, err = rp.client.CoreV1().Endpoints(ns).Watch(ctx, metav1.ListOptions{FieldSelector: fields.SelectorFromSet(fields.Set{"metadata.name": headlessServiceName}).String()})
		if err == nil {
			break
		}
	}

	if err != nil {
		logrus.Fatalf("Failed to access k8s api to list fn runner endpoints, is rbac enabled? (check rbac.enabled in your helm chart) : %s", err)
		panic(err)
	}

	logrus.WithField("Namespace", ns).WithField("HeadlessService", headlessServiceName).Info("Watching for endpoints changes")

	// This runs forever.
	for {
		select {
		case <-rp.cancel:
			logrus.Info("Stopped watching for endpoints changes")
			epWatch.Stop()
			return
		case event := <-epWatch.ResultChan():

			if endpoint, ok := event.Object.(*v1.Endpoints); ok {
				logrus.WithField("Event", event.Type).Debug("Endpoint change detected")
				switch event.Type {
				case watch.Added:
					fallthrough

				case watch.Modified:
					list := extractAddress(endpoint)
					logrus.WithField("Endpoints", list).Info("New Endpoint detected")
					rp.registerNode(list, tlsConf, dialOpts)

				case watch.Deleted:
					list := extractAddress(endpoint)
					logrus.WithField("Endpoints", list).Info("Endpoint removed")
					rp.unregisterNode(list)
				}
			}
		}
	}
}

// Manage node update and deletion under the hood
func (rp *k8sRunnerPool) registerNode(newAddrs []string, tlsConf *tls.Config, dialOpts []grpc.DialOption) {
	logrus.Debug("registerNode")
	rp.nodeLock.Lock()
	defer rp.nodeLock.Unlock()

	for _, newadr := range newAddrs {
		if rp.runners[newadr] == nil {
			logrus.WithField("address", newadr).Debug("Adding node")
			newRunner, err := NewgRPCRunner(newadr, tlsConf, dialOpts...)
			if err == nil {
				rp.runners[newadr] = newRunner
				logrus.WithField("len(rp.runners)", len(rp.runners)).Debug("new runners size")
			} else {
				logrus.WithField("err", err).Warn("Error instantiating runner")
			}
		} else {
			logrus.WithField("address", newadr).Debug("Runner already registered, skipping")
		}
	}

}

func (rp *k8sRunnerPool) unregisterNode(newAddrs []string) {
	rp.nodeLock.Lock()
	defer rp.nodeLock.Unlock()

	for _, deladr := range newAddrs {
		if rp.runners[deladr] == nil {
			logrus.WithField("address", deladr).Debug("Runner to delete not found, skipping")
		} else {
			logrus.WithField("address", deladr).Debug("Removing node")
			delete(rp.runners, deladr)
			logrus.WithField("len(rp.runners)", len(rp.runners)).Warn("new runners size")
		}
	}
}

//////////////
// UTILITY FUNCTIONS
//////////////

// Initialized the K8s Client
func k8sClient() (*kubernetes.Clientset, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(k8sConfig)
}

// Return the current namespace
func namespace() (string, error) {
	ns, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + v1.ServiceAccountNamespaceKey)
	if err != nil {
		// If we're not running inside k8s, bail out
		return "", err
	}

	return strings.TrimSpace(string(ns)), nil
}

// Extract the addresses from k8s endpoint
func extractAddress(endpoints *v1.Endpoints) []string {
	var result []string

	for _, s := range endpoints.Subsets {
		port := s.Ports[0].Port
		for _, a := range s.Addresses {
			address := fmt.Sprintf("%s:%d", a.IP, port)
			result = append(result, address)
		}
	}

	return result
}
