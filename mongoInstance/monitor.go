package mongoInstance

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Monitor struct {
	MasterIP *string
}

func (m *Monitor) HealthCheck(healthChannel chan bool) {
	res, resErr := http.Get(fmt.Sprintf("http://%s", *m.MasterIP))
	if resErr != nil {
		log.Fatal(resErr)
		healthChannel <- false
	} else if res.Status != "200 OK" {
		log.Fatal(errors.New(fmt.Sprintf("healthcheck responded with status other than 200 OK, %s", res.Status)))
		healthChannel <- false
	}
	healthChannel <- true
}

func newMonitor(masterIP *string) *Monitor {
	return &Monitor{masterIP}
}
