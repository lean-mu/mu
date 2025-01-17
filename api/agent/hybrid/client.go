package hybrid

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lean-mu/mu/api/agent"
	"github.com/lean-mu/mu/api/common"
	"github.com/lean-mu/mu/api/models"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

// TODO(reed): this should use the fn_go bindings now, most likely, but get trigger by source
// needs to be dealt with before that can happen.

// client implements agent.DataAccess
type client struct {
	base string
	http *http.Client
}

func NewClient(u string) (agent.DataAccess, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if uri.Host == "" {
		return nil, errors.New("no host specified for client")
	}
	if uri.Scheme == "" {
		uri.Scheme = "http"
	}
	host := uri.Scheme + "://" + uri.Host + "/v2/"
	logrus.WithField("host", host).Debugf("NewClient")

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 128,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				ClientSessionCache: tls.NewLRUClientSessionCache(8096),
			},
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &client{
		base: host,
		http: httpClient,
	}, nil
}

var noQuery = map[string]string{}

func (cl *client) GetAppID(ctx context.Context, appName string) (string, error) {
	ctx, span := trace.StartSpan(ctx, "hybrid_client_get_app_id")
	defer span.End()

	var a struct {
		Items []*models.App `json:"items"`
	}

	err := cl.do(ctx, nil, &a, "GET", map[string]string{"name": appName}, "apps")

	if len(a.Items) == 0 {
		return "", errors.New("app not found")
	}

	return a.Items[0].ID, err
}

func (cl *client) GetAppByID(ctx context.Context, appID string) (*models.App, error) {
	ctx, span := trace.StartSpan(ctx, "hybrid_client_get_app_by_id")
	defer span.End()

	var a models.App
	err := cl.do(ctx, nil, &a, "GET", noQuery, "apps", appID)
	return &a, err
}

func (cl *client) GetTriggerBySource(ctx context.Context, appID string, triggerType, source string) (*models.Trigger, error) {
	ctx, span := trace.StartSpan(ctx, "hybrid_client_get_trigger_by_source")
	defer span.End()

	var trigger models.Trigger
	err := cl.do(ctx, nil, &trigger, "GET", noQuery, "runner", "apps", appID, "triggerBySource", triggerType, source)
	return &trigger, err
}

func (cl *client) GetFnByID(ctx context.Context, fnID string) (*models.Fn, error) {
	ctx, span := trace.StartSpan(ctx, "hybrid_client_get_fn_by_id")
	defer span.End()

	var fn models.Fn
	err := cl.do(ctx, nil, &fn, "GET", noQuery, "fns", fnID)
	if err != nil {
		return nil, err
	}
	return &fn, nil
}

type httpErr struct {
	code int
	error
}

func (cl *client) do(ctx context.Context, request, result interface{}, method string, query map[string]string, url ...string) error {

	// Sequence of 25..75  25..175  25..200  25..725  25..1575
	backoff := common.NewBackOff(common.BackOffConfig{
		MaxRetries: 5,
		Interval:   50,
		MinDelay:   25,
	})

	timer := common.NewTimer(25 * time.Millisecond)
	defer timer.Stop()

	for {
		// TODO this isn't re-using buffers very efficiently, but retries should be rare...
		err := cl.once(ctx, request, result, method, query, url...)
		switch err := err.(type) {
		case nil:
			return nil
		case *httpErr:
			if err.code < 500 {
				return err
			}
			// retry 500s...
		default:
			// this error wasn't from us [most likely], probably a conn refused/timeout, just retry it out
		}

		common.Logger(ctx).WithError(err).Error("error from API server, retrying")
		delay, ok := backoff.NextBackOff()
		if !ok {
			return err
		}

		timer.Reset(delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (cl *client) once(ctx context.Context, request, result interface{}, method string, query map[string]string, path ...string) error {
	ctx, span := trace.StartSpan(ctx, "hybrid_client_http_do")
	defer span.End()

	var b bytes.Buffer // TODO pool
	if request != nil {
		err := json.NewEncoder(&b).Encode(request)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, cl.url(query, path...), &b)
	if err != nil {
		return err
	}
	// shove the span headers in so that the server will continue this span
	var xxx b3.HTTPFormat
	xxx.SpanContextToRequest(span.SpanContext(), req)

	resp, err := cl.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { io.Copy(ioutil.Discard, resp.Body); resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		// one of our errors
		var msg struct {
			Msg string `json:"message"`
		}
		// copy into a buffer in case it wasn't from us
		var b bytes.Buffer
		io.Copy(&b, resp.Body)
		json.Unmarshal(b.Bytes(), &msg)
		if msg.Msg != "" {
			return &httpErr{code: resp.StatusCode, error: errors.New(msg.Msg)}
		}
		return &httpErr{code: resp.StatusCode, error: errors.New(b.String())}
	}

	if result != nil {
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cl *client) url(query map[string]string, args ...string) string {

	var queryValues = make(url.Values)
	for k, v := range query {
		queryValues.Add(k, v)
	}
	queryString := queryValues.Encode()

	baseUrl := cl.base + strings.Join(args, "/")

	if queryString != "" {
		baseUrl = baseUrl + "?" + queryString
	}
	return baseUrl
}

func (cl *client) Close() error {
	return nil
}
