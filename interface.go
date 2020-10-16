package freezedetector

import "time"

type DetectorI interface {
	NewRequest(id string, timeout time.Duration, where whereamiType) RequestI
}

type RequestI interface {
	Callstack() (callstack []string)
	NewFunc(name string, where whereamiType) *requestFunc
	GracefulClose()
	Close()
}

type ProblemI interface {
	When() time.Time
	Where() string
	Body() string
	Request() *request
}
