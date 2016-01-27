/*
Copyright 2014 Christian Grabowski All rights reserved.
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

package kubeClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	kube "golang.org/x/build/kubernetes"
	api "golang.org/x/build/kubernetes/api"
	"golang.org/x/net/context"
)

type Controller struct {
	APIServerIP string
	Client      *kube.Client
	rawClient   *http.Client
	Namespace   string
	EnvVarName  string
}

func New(apiServerIP, namespace, evn string) *Controller {
	httpClient := &http.Client{}
	kubeClient, kubeErr := kube.NewClient(apiServerIP, httpClient)
	if kubeErr != nil {
		panic(kubeErr)
	}
	return &Controller{
		APIServerIP: apiServerIP,
		Client:      kubeClient,
		rawClient:   httpClient,
		Namespace:   namespace,
		EnvVarName:  evn,
	}
}

type execPayload struct {
	Command   string `json:"command"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (c *Controller) kubeExec(pod api.Pod, cmd string, execErrs chan error) {
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/pods/%s/exec", c.APIServerIP, c.Namespace, pod.ObjectMeta.Name)
	payload := &execPayload{
		Command:   cmd,
		Name:      pod.ObjectMeta.Name,
		Namespace: c.Namespace,
	}
	payloadJSON, pErr := json.Marshal(payload)
	if pErr != nil {
		execErrs <- pErr
	}
	resp, resErr := c.rawClient.Post(url, "Application/JSON", bytes.NewBuffer(payloadJSON))
	if resErr != nil {
		execErrs <- resErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		execErrs <- fmt.Errorf("%s received a status code of %d", url, resp.StatusCode)
	}
}

func (c *Controller) UpdateServiceEndPoint(newEndPoint string) error {
	ctx := context.Background()
	pods, pErr := c.Client.GetPods(ctx)
	if pErr != nil {
		return pErr
	}
	errsChan := make(chan error)
	for i := range pods {
		go c.kubeExec(pods[i], fmt.Sprintf("export %s=%s", c.EnvVarName, newEndPoint), errsChan)
		select {
		case err := <-errsChan:
			if err != nil {
				return err
			}
		}
	}
	return nil
}
