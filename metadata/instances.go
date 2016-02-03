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

package metadata

import (
	"fmt"

	"github.com/cpg1111/kubongo/hostProvider"
)

// Instances is a slice of instances
type Instances []hostProvider.Instance

// ToMap converts slice of instances to a map of instances
func (inst Instances) ToMap() (instanceMap map[string]hostProvider.Instance) {
	for i := range inst {
		castInst := inst[i].(hostProvider.LocalInstance)
		instanceMap[fmt.Sprintf("%v", castInst.Name)] = inst[i]
	}
	return
}

var current *Instances

// New creates a new Instance slice
func New(firstInstance *hostProvider.Instance) *Instances {
	if current != nil {
		return current
	}
	if firstInstance != nil {
		current = &Instances{*firstInstance}
		return current
	}
	current = &Instances{}
	return current
}

// AddInstance will add an instance to the Instances slice
func AddInstance(list *Instances, instance hostProvider.Instance) *Instances {
	newList := append(*list, instance)
	list = &newList
	return list
}

// RemoveInstance will remove an instance from the slice
func RemoveInstance(list Instances, instance hostProvider.Instance) Instances {
	newList := make(Instances, len(list)-1)
	for i := range list {
		castInst := list[i].(hostProvider.LocalInstance)
		castOther := instance.(hostProvider.LocalInstance)
		if castInst.Name != castOther.Name {
			newList[i] = list[i]
		}
	}
	return newList
}
