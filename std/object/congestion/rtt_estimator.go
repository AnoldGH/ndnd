package congestion

import (
	"sync"
	"time"

	"github.com/named-data/ndnd/std/log"
)

// RTTEstimator provides an interface for estimating round-trip time.
type RTTEstimator interface {
	String() string

	EstimatedRTT() time.Duration	// get the estimated RTT
	DeviationRTT() time.Duration	// get the deviation of RTT

	AddMeasurement(sample time.Duration, retransmitted bool)	// add a new RTT measurement
}

// KarnRTTEstimator is an implementation of RTTEstimator using Karn's algorithm.
type KarnRTTEstimator struct {
	mutex	sync.RWMutex

	estimatedRTT time.Duration		// estimated RTT
	deviationRTT time.Duration		// deviation of RTT

	alpha float64					// alpha value for exponential smoothing
	beta  float64					// beta value for exponential smoothing
}

// NewKarnRTTEstimator creates a new KarnRTTEstimator.
func NewKarnRTTEstimator(alpha float64, beta float64) *KarnRTTEstimator {
	return &KarnRTTEstimator{
		estimatedRTT: 0.0,
		deviationRTT: 0.0,

		alpha: alpha,
		beta: beta,
	}
}

func (rtt *KarnRTTEstimator) String() string {
	return "karn-rtt-estimator"
}

// EstimatedRTT returns the estimated RTT.
func (rtt *KarnRTTEstimator) EstimatedRTT() time.Duration {
	rtt.mutex.RLock()
	defer rtt.mutex.RUnlock()

	return rtt.estimatedRTT
}

// DeviationRTT returns the deviation of RTT.
func (rtt *KarnRTTEstimator) DeviationRTT() time.Duration {
	rtt.mutex.RLock()
	rtt.mutex.RUnlock()

	return rtt.deviationRTT
}

func (rtt *KarnRTTEstimator) AddMeasurement(sample time.Duration, retransmitted bool) {
	if retransmitted {
		return	// ignore retransmitted packets
	}

	rtt.mutex.Lock()
	defer rtt.mutex.Unlock()

	// calculate new RTT using Karn's algorithm
	newEstimatedRTT := rtt.estimatedRTT.Seconds() + rtt.alpha * (sample - rtt.estimatedRTT).Seconds()
	newDeviationRTT := rtt.deviationRTT.Seconds() + rtt.beta * (sample - rtt.estimatedRTT).Seconds()

	// update RTT
	rtt.estimatedRTT = time.Duration(newEstimatedRTT * float64(time.Second))
	rtt.deviationRTT = time.Duration(newDeviationRTT * float64(time.Second))

	log.Debug(rtt, "new RTT", rtt.estimatedRTT.Seconds(), "new deviation", rtt.deviationRTT.Seconds())
}