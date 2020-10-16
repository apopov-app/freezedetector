package freezedetector

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestLossOfControlDetect(t *testing.T) {
	var problems []ProblemI
	detector := NewDetector(func(problem ProblemI) {
		problems = append(problems, problem)
	})
	func() {
		myRequest := detector.NewRequest("TESTREQUEST", -1, Whereami())
		defer myRequest.Close()
		// We start the gouroutine but wait only 1 second, instead of 30. Abnormal termination
		// fd.Close will be executed after the request is closed
		go func() {
			fd := myRequest.NewFunc("AnonymousFuncOne", Whereami()) // Detect "loss of control"
			time.Sleep(time.Second * 30)
			fd.Close()
		}()
		time.Sleep(time.Second * 1)
		myRequest.GracefulClose()
	}()

	CheckProblems(t, problems, reflect.TypeOf(&LossOfControlProblem{}))
}

func TestTimeoutDetect(t *testing.T) {
	var problems []ProblemI
	detector := NewDetector(func(problem ProblemI) {
		problems = append(problems, problem)
	})
	func() {
		myRequest := detector.NewRequest("TESTREQUEST", time.Second * 1, Whereami()) // Request timeout 1 second
		defer myRequest.Close()
		func() {
			fd := myRequest.NewFunc("AnonymousFuncOne", Whereami())
			time.Sleep(time.Second * 3) // Sleep 3 second, make error
			defer fd.Close()
		}()

		myRequest.GracefulClose()
	}()

	CheckProblems(t, problems, reflect.TypeOf(&RequestTimeoutProblem{}))
}

func TestGracefulCloseDetect(t *testing.T) {
	var problems []ProblemI
	detector := NewDetector(func(problem ProblemI) {
		problems = append(problems, problem)
	})
	func() {
		myRequest := detector.NewRequest("TESTREQUEST", -1, Whereami()) // Request timeout 1 second
		defer myRequest.Close()

		fd := myRequest.NewFunc("AnonymousFuncOne", Whereami())
		fd.Close()
	}()
	// I not call myRequest.GracefulClose()
	// Emulation of the problem that the function was not completed to the end and only exited by defer

	CheckProblems(t, problems, reflect.TypeOf(&RequestNotGracefulClose{}))
}

func CheckProblems(t *testing.T, problems []ProblemI, refType reflect.Type) {
	if len(problems) != 1 {
		t.Fatal("Test return " + strconv.Itoa(len(problems)) + " problems, need only one")
	}
	problem := problems[0]

	if !reflect.TypeOf(problem).AssignableTo(refType) {
		t.Fail()
	}
}