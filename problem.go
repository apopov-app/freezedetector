package freezedetector

import (
	"time"
)

type baseProblem struct {
	time time.Time
	where WhereamiType
	request *request
}

func (problem *baseProblem) Where() string {
	return string(problem.where)
}

func (problem *baseProblem) Request() RequestI {
	return problem.request
}

func (problem *baseProblem) When() time.Time {
	return problem.time
}

type LossOfControlProblem struct {
	baseProblem
	funcName string
}

func (problem *LossOfControlProblem) Body() string {
	return "Func " + problem.funcName + " continued execution after the request finished."
}

type RequestTimeoutProblem struct {
	baseProblem
	timeout time.Duration
	requestId string
}

func (problem *RequestTimeoutProblem) Body() string {
	return "Request "+ problem.requestId +" not executed in the specified time ("+problem.timeout.String()+")"
}

type RequestNotGracefulClose struct {
	baseProblem
}

func (problem *RequestNotGracefulClose) Body() string {
	return "Request "+ problem.request.requestId +" not Graceful Close. Panic maybe?"
}