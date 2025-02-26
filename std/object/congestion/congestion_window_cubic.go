package congestion

import (
	"math"
	"sync"
	"time"

	"github.com/named-data/ndnd/std/log"
)

// CUBICCongestionWindow is an implementation of CongestionWindow using CUBIC algorithm
// ref: https://tools.ietf.org/html/rfc8312
type CUBICCongestionWindow struct {
	mutex		sync.RWMutex

	window 		float64					// window size
	eventCh		chan WindowEvent		// channel for emitting window change event

	rttEstimator 	RTTEstimator		// RTT estimator

	ssthresh		float64				// slow start threshold
	minSsthresh 	float64				// minimum slow start threshold
	aiStep			float64				// additive increase step
	mdCoef 			float64				// multiplicative decrease coefficient
	c 				float64				// aggressiveness factor
	windowMax 		float64				// maximum window size
	lastWindowMax 	float64				// last maximum window size
	fastConvergence bool				// whether to use fast convergence
	lastDecrease 	time.Time			// time of last window-decrease event
}

func NewCUBICCongestionWindow(cwnd int) *CUBICCongestionWindow {
	return &CUBICCongestionWindow{
		window: float64(cwnd),
		eventCh: make(chan WindowEvent),

		rttEstimator: NewKarnRTTEstimator(0.1, 0.1),

		ssthresh: math.MaxFloat64,
		minSsthresh: 2.0,
		aiStep: 1.0,
		mdCoef: 0.7,
		c: 0.4,
		windowMax: float64(cwnd),
		lastWindowMax: float64(cwnd),
		fastConvergence: true,
		lastDecrease: time.Now(),
	}
}

func (cw *CUBICCongestionWindow) String() string {
	return "cubic-congestion-window"
}

func (cw *CUBICCongestionWindow) Size() int {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()

	return int(cw.window)
}

// Cubic update algorithm
func (cw *CUBICCongestionWindow) CubicUpdate() {
	rtt := cw.rttEstimator.EstimatedRTT().Seconds()			// estimated RTT
	t := time.Since(cw.lastDecrease).Abs().Seconds()		// time since last decrease
	k := math.Cbrt((cw.windowMax * (1 - cw.mdCoef)) / cw.c)	// time takes to increase window to windowMax

	// estimated cubic window size
	wCubic := cw.c * math.Pow(t + rtt - k, 3) + cw.windowMax

	// TCP-friendly window size
	wEst := cw.windowMax * cw.mdCoef + 3 * (1 - cw.mdCoef) / (1 + cw.mdCoef) * t / rtt

	// update window size
	if cw.window < wEst {
		cw.window = wEst
	} else {
		cw.window += (wCubic - cw.window) / cw.window
	}
}

func (cw *CUBICCongestionWindow) IncreaseWindow() {
	cw.mutex.Lock()

	// slow start
	if cw.window <= cw.ssthresh {
		cw.window += 1
	} else {
		// congestion avoidance
		cw.CubicUpdate()
	}

	cw.mutex.Unlock()

	cw.EmitWindowEvent(time.Now(), cw.Size())
}

func (cw *CUBICCongestionWindow) DecreaseWindow() {
	cw.mutex.Lock()

	// update windowMax
	cw.windowMax = cw.window

	// faster convergence
	if cw.windowMax < cw.lastWindowMax && cw.fastConvergence {
		cw.lastWindowMax = cw.windowMax
		cw.windowMax *= (1 + cw.mdCoef) / 2	// further decrease windowMax
	} else {
		cw.lastWindowMax = cw.windowMax
	}

	// decrease window size
	cw.window *= cw.mdCoef
	cw.ssthresh = math.Max(cw.window, cw.minSsthresh)
	cw.lastDecrease = time.Now()

	cw.mutex.Unlock()

	cw.EmitWindowEvent(time.Now(), cw.Size())
}

func (cw *CUBICCongestionWindow) EventChannel() <-chan WindowEvent {
	return cw.eventCh
}

func (cw *CUBICCongestionWindow) HandleSignal(signal CongestionSignal) {
	switch signal {
	case SigData:
		cw.IncreaseWindow()
	case SigLoss, SigCongest:
		cw.DecreaseWindow()
	default:
		// no-op
	}
}

func (cw *CUBICCongestionWindow) EmitWindowEvent(age time.Time, cwnd int) {
	// non-blocking send to the channel
	select {
	case cw.eventCh <- WindowEvent{age: age, cwnd: cwnd}:
	default:
		// if the channel is full, we log the change event
		log.Debug(cw, "Window size changes", "window", cw.window)
	}
}