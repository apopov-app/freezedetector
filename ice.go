package freezedetector

import (
	"github.com/jimlawless/whereami"
	"strconv"
	"time"
)

type requestFunc struct {
	request *request
	time time.Time
	where whereamiType
	funcName string
	close bool
}

func (r *requestFunc) Close()  {
	r.close = true
}

type request struct {
	*freezedetector
	requestId string
	close bool
	gracefulClose bool
	where whereamiType
	callstack []*requestFunc
}

func (freezRequest *request) Callstack() (callstack []string) {
	for _, fn := range freezRequest.callstack {
		callstack = append(callstack, fn.time.String() + " - " + fn.funcName + "("+string(fn.where)+") close: " + strconv.FormatBool(fn.close))
	}
	return callstack
}

func (freezRequest *request) NewFunc(name string, where whereamiType) *requestFunc  {
	rFunc := &requestFunc{
		request: freezRequest,
		funcName: name,
		time: time.Now(),
		where: where,
	}
	freezRequest.callstack = append(freezRequest.callstack, rFunc)

	return rFunc
}

func (freezRequest *request) GracefulClose() {
	freezRequest.gracefulClose = true
}

func (freezRequest *request) Close() {
	// Detect "Loss Of Control"
	for _, fn := range freezRequest.callstack {
		if !fn.close {
			freezRequest.problemCallback(&LossOfControlProblem{
				baseProblem: baseProblem{
					time:      time.Now(),
					where: fn.where,
					request: freezRequest,
				},
				funcName: fn.funcName,
			})
		}
	}

	if !freezRequest.gracefulClose {
		freezRequest.problemCallback(&RequestNotGracefulClose{
			baseProblem: baseProblem{
				time:      time.Now(),
				where: freezRequest.where,
				request: freezRequest,
			},
		})
	}
	freezRequest.close = true
}

type freezedetector struct {
	problemCallback func(problem ProblemI)
}

func (freez *freezedetector) NewRequest(id string, timeout time.Duration, where whereamiType) RequestI {
	fd := &request{
		freezedetector: freez,
		requestId: id,
		where: where,
	}
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		go func() {
			<-timer.C
			if !fd.close {
				freez.problemCallback(&RequestTimeoutProblem{
					baseProblem: baseProblem{
						where: fd.where,
						time:      time.Now(),
						request: fd,
					},
					timeout: timeout,
					requestId: fd.requestId,
				})
			}
		}()
	}

	return fd
}

func NewDetector(callback func(problem ProblemI)) *freezedetector {
	return &freezedetector{
		problemCallback: callback,
	}
}

type whereamiType string
func Whereami() whereamiType {
	return whereamiType(whereami.WhereAmI(2))
}