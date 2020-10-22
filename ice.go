package freezedetector

import (
	"github.com/jimlawless/whereami"
	"strconv"
	"time"
	"sync"
)

type requestFunc struct {
	mu *sync.Mutex

	request *request
	time time.Time
	where WhereamiType
	funcName string
	close bool
}

func (r *requestFunc) Close()  {
	r.mu.Lock()
	r.close = true
	r.mu.Unlock()
}

type request struct {
	mu *sync.RWMutex

	*freezedetector
	requestId string
	close bool
	gracefulClose bool
	where WhereamiType
	callstack []*requestFunc
}

func (freezRequest *request) Callstack() (callstack []string) {
	freezRequest.mu.RLock()
	defer freezRequest.mu.RUnlock()

	for _, fn := range freezRequest.callstack {
		callstack = append(callstack, fn.time.String() + " - " + fn.funcName + "("+string(fn.where)+") close: " + strconv.FormatBool(fn.close))
	}
	return callstack
}

func (freezRequest *request) NewFunc(name string, where WhereamiType) RequestFuncI  {
	freezRequest.mu.RLock()
	isClose := freezRequest.close
	freezRequest.mu.RUnlock()

	if isClose {
		freezRequest.problemCallback(&LossOfControlProblem{
			baseProblem: baseProblem{
				time:      time.Now(),
				where: where,
				request: freezRequest,
			},
			funcName: name,
		})
	}

	rFunc := &requestFunc{
		mu: &sync.Mutex{},
		request: freezRequest,
		funcName: name,
		time: time.Now(),
		where: where,
	}
	freezRequest.mu.Lock()
	freezRequest.callstack = append(freezRequest.callstack, rFunc)
	freezRequest.mu.Unlock()
	return rFunc
}

func (freezRequest *request) GracefulClose() {
	freezRequest.mu.Lock()
	freezRequest.gracefulClose = true
	freezRequest.mu.Unlock()
}

func (freezRequest *request) Close() {
	// Detect "Loss Of Control"
	freezRequest.mu.RLock()
	callstack := freezRequest.callstack
	isGracefulClose := freezRequest.gracefulClose
	freezRequest.mu.RUnlock()

	for _, fn := range callstack {
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

	if !isGracefulClose {
		freezRequest.problemCallback(&RequestNotGracefulClose{
			baseProblem: baseProblem{
				time:      time.Now(),
				where: freezRequest.where,
				request: freezRequest,
			},
		})
	}

	freezRequest.mu.Lock()
	freezRequest.close = true
	freezRequest.mu.Unlock()
}

type freezedetector struct {
	problemCallback func(problem ProblemI)
}

func (freez *freezedetector) NewRequest(id string, timeout time.Duration, where WhereamiType) RequestI {
	fd := &request{
		mu: &sync.RWMutex{},
		freezedetector: freez,
		requestId: id,
		where: where,
	}
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		go func() {
			<-timer.C
			fd.mu.RLock()
			defer fd.mu.RUnlock()

			if !fd.close {
				var where WhereamiType
				calltraceLen := len(fd.callstack)
				if calltraceLen > 0 {
					where = fd.callstack[len(fd.callstack)-1].where
				} else {
					where = fd.where
				}

				freez.problemCallback(&RequestTimeoutProblem{
					baseProblem: baseProblem{
						where:     where,
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

type WhereamiType string
func Whereami() WhereamiType {
	return WhereamiType(whereami.WhereAmI(2))
}