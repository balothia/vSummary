package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gbolo/vsummary/common"
	"gopkg.in/go-playground/validator.v9"
	//"github.com/thoas/stats"
	//"github.com/codegangsta/negroni"
)

func handlerPoller(w http.ResponseWriter, req *http.Request) {

	// log time on debug
	defer common.ExecutionTime(time.Now(), "handle")

	// read in body
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("error reading request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Body.Close()

	// decode json body
	var poller common.Poller
	err = json.Unmarshal(reqBody, &poller)
	if err != nil {
		log.Errorf("failed to decode body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate
	validate := validator.New()

	err = validate.Struct(poller)
	if err != nil {
		log.Errorf("failed to validate body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert to backend
	err = backend.InsertPoller(poller)
	if err != nil {
		log.Errorf("failed to insert poller: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
