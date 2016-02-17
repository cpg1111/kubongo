/*
Copyright 2015 Christian Grabowski All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	} else {
		log.Println(*m.MasterIP, "passes health check")
		healthChannel <- true
	}
	return
}

func newMonitor(masterIP *string) *Monitor {
	return &Monitor{masterIP}
}
