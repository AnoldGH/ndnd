package congestion

import (
	"errors"
	"math"
)

type WindowType string
const (
	Fixed WindowType 	= "fixed"
	AIMD WindowType 	= "aimd"
)

// Set sets the window type from a string
func (t *WindowType) Set(s string) error {
	switch s {
	case "fixed":
		*t = Fixed
	case "aimd":
		*t = AIMD
	default:
		return errors.New(`must be one of "fixed", "aimd"`)
	}
	return nil
}

// String returns the string representation of the window type
func (t *WindowType) String() string {
	return string(*t)
}

// Type returns the type of the window type
func (t *WindowType) Type() string {
	return "windowType"
}

// CongestionOptions is a struct that holds the options for congestion control
type CongestionOptions struct {
	// shared options
	Type WindowType		// window type (fixed, aimd)
	InitCwnd int		// initial window size

	// Fixed options
	// none

	// AIMD options
	ssthresh int		// slow start threshold
	minSsthresh int		// minimum slow start threshold
	aiStep int			// additive increase step
	mdCoef float64		// multiplicative decrease coefficient
	resetCwnd bool		// whether to reset cwnd after decrease
}

// NewCongestionOptions creates a CongestionOptions struct with default values
func NewCongestionOptions() *CongestionOptions {
	return &CongestionOptions{
		Type: Fixed,
		InitCwnd: 100,

		ssthresh: math.MaxInt,
		minSsthresh: 16,
		aiStep: 1,
		mdCoef: 0.5,
		resetCwnd: false,
	}
}