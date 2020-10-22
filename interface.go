package freezedetector

import "time"

type DetectorI interface {
	NewRequest(id string, timeout time.Duration, where WhereamiType) RequestI
}

type RequestI interface {
	Callstack() (callstack []string)
	NewFunc(name string, where WhereamiType) RequestFuncI
	GracefulClose()
	Close()
}

type ProblemI interface {
	When() time.Time
	Where() string
	Body() string
	Request() RequestI
}

type RequestFuncI interface {
	Close()
}