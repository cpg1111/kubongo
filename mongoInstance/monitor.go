package mongoInstance

import (
	"fmt"
	"log"
	"net/http"
)

// Monitor master mongo instance
type Monitor struct {
	MasterIP *string
}

// HealthCheck master mongo instance
func (m *Monitor) HealthCheck(healthChannel chan bool) {
	res, resErr := http.Get(fmt.Sprintf("http://%s", *m.MasterIP))
	if resErr != nil {
		log.Println(resErr)
		healthChannel <- false
	} else if res.Status != "200 OK" {
		log.Println(fmt.Errorf("healthcheck responded with status other than 200 OK, %s", res.Status))
		healthChannel <- false
	}
	log.Println(*m.MasterIP, "passes health check")
	healthChannel <- true
}

func newMonitor(masterIP *string) *Monitor {
	return &Monitor{masterIP}
}
