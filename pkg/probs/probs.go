package probs

import (
	"errors"
	"net/http"
	"sync"
)

var (
	readiness, liveness error
	isReady             bool
)
var mu = sync.RWMutex{}

func SetReadinessError(e error) {
	mu.Lock()
	defer mu.Unlock()
	readiness = e
}

func Ready() {
	mu.Lock()
	defer mu.Unlock()
	isReady = true
}

func IsReady() error {
	mu.RLock()
	defer mu.RUnlock()

	if readiness != nil {
		return readiness
	}

	if !isReady {
		return errors.New(`not ready`)
	}

	return nil
}

func SetLivenessError(e error) {
	mu.Lock()
	defer mu.Unlock()
	liveness = e
}

func GetLivenessError() error {
	mu.RLock()
	defer mu.RUnlock()
	return liveness
}

func Liveness(w http.ResponseWriter, r *http.Request) {
	err := GetLivenessError()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Write([]byte(err.Error()))
	w.WriteHeader(http.StatusBadRequest)
}

func Readiness(w http.ResponseWriter, r *http.Request) {
	err := IsReady()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Write([]byte(err.Error()))
	w.WriteHeader(http.StatusBadRequest)
}
