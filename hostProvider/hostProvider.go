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

package hostProvider

// Instance is an interface for a platform's instances
type Instance interface {
	// GetInternalIP returns a string of the instance's internal IP
	GetInternalIP() string
}

//HostProvider is the interface for HostProviders for each platform to control instances on the platform
type HostProvider interface {
	// GetServers returns a slice of instances
	GetServers(namespace string) ([]Instance, error)
	// GetServer returns a specific instance
	GetServer(project, zone, name string) (Instance, error)
	// CreateServer creates an instance for the platform
	CreateServer(namespace, zone, name, machineType, sourceImage, source string) (Instance, error)
	// DeleteServer deletes an instance for the platform
	DeleteServer(namespace, zone, name string) error
}
