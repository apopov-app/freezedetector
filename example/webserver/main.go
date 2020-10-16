package main

import (
	"fmt"
	"github.com/ruelephant/freezedetector"
	"log"
	"net/http"
	"time"
)
type HelloHandler struct {
	detector freezedetector.DetectorI
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	myRequest := h.detector.NewRequest("HANDLER NAME", time.Second * 1, freezedetector.Whereami())
	defer myRequest.Close()

	time.Sleep(time.Second * 3) // Over limit
	fmt.Fprintf(w, "hello\n")

	myRequest.GracefulClose()
}

func main() {
	detector := freezedetector.NewDetector(func(problem freezedetector.ProblemI) {
		log.Println("problem! ", problem.When(), problem.Where(), problem.Body())
	})

	helloHandler := &HelloHandler{
		detector: detector,
	}
	http.Handle("/hello", helloHandler)

	http.ListenAndServe(":8090", nil)
}