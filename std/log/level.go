package log

import "fmt"

type Level int

const LevelTrace Level = -8
const LevelDebug Level = -4
const LevelInfo Level = 0
const LevelWarn Level = 4
const LevelError Level = 8
const LevelFatal Level = 12

func ParseLevel(s string) (Level, error) {
	switch s {
	case "TRACE":
		return LevelTrace, nil
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	case "FATAL":
		return LevelFatal, nil
	}
	return LevelInfo, fmt.Errorf("invalid log level: %s", s)
}

func (level Level) String() string {
	switch level {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
