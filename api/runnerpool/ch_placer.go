/* The consistent hash ring from the original fnlb.
   The behaviour of this depends on changes to the runner list leaving it relatively stable.
   The algorythm is similar to a URL Hash algorithm but usues the FNid as an input
   https://kemptechnologies.com/load-balancer/load-balancing-algorithms-techniques/
*/
package runnerpool

import (
	"context"

	"github.com/lean-mu/mu/api/models"

	"github.com/dchest/siphash"
	"github.com/sirupsen/logrus"
)

type chPlacer struct {
	cfg PlacerConfig
}

func NewCHPlacer(cfg *PlacerConfig) Placer {
	logrus.Infof("Creating new CH runnerpool placer with config=%+v", cfg)
	return &chPlacer{
		cfg: *cfg,
	}
}

func (p *chPlacer) GetPlacerConfig() PlacerConfig {
	return p.cfg
}

// This borrows the CH placement algorithm from the original FNLB.
// Because we ask a runner to accept load (queuing on the LB rather than on the nodes), we don't use
// the LB_WAIT to drive placement decisions: runners only accept work if they have the capacity for it.
func (p *chPlacer) PlaceCall(ctx context.Context, rp RunnerPool, call RunnerCall) error {
	logrus.Debugf("CH Placer - place call for fnId: %s ", call.Model().FnID)
	state := NewPlacerTracker(ctx, &p.cfg, call)
	logrus.Debugf("1")
	defer state.HandleDone()
	logrus.Debugf("2")

	key := call.Model().FnID
	logrus.Debugf("3")

	sum64 := siphash.Hash(0, 0x4c617279426f6174, []byte(key))

	logrus.Debugf("sum64 = %d", sum64 )

	var runnerPoolErr error
	for {
		var runners []Runner
		runners, runnerPoolErr = rp.Runners(ctx, call)

		logrus.Debugf("len(runners)=%d", len(runners) )

		i := int(jumpConsistentHash(sum64, int32(len(runners))))

		logrus.Debugf("jumpConsistentHash = %d", i )

		for j := 0; j < len(runners) && !state.IsDone(); j++ {

			r := runners[i]
			logrus.Debugf("r.Address() %s", r.Address() )


			logrus.Debugf("4")

			placed, err := state.TryRunner(r, call)
			logrus.Debugf("5")

			if placed {
				return err
			}

			logrus.Debugf("6")

			i = (i + 1) % len(runners)
		}

		if !state.RetryAllBackoff(len(runners), runnerPoolErr) {
			break
		}
	}

	if runnerPoolErr != nil {
		// If we haven't been able to place the function and we got an error
		// from the runner pool, return that error (since we don't have
		// enough runners to handle the current load and the runner pool is
		// having trouble).
		state.HandleFindRunnersFailure(runnerPoolErr)
		return runnerPoolErr
	}

	logrus.Debugf("Returning ErrCallTimeoutServerBusy")
	return models.ErrCallTimeoutServerBusy
}

// A Fast, Minimal Memory, Consistent Hash Algorithm:
// https://arxiv.org/ftp/arxiv/papers/1406/1406.2294.pdf
func jumpConsistentHash(key uint64, num_buckets int32) int32 {
	var b, j int64 = -1, 0
	for j < int64(num_buckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = (b + 1) * int64((1<<31)/(key>>33)+1)
	}
	return int32(b)
}
