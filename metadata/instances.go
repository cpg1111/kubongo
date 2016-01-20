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
		castInst := inst[i].(hostProvider.GcloudInstance)
		instanceMap[fmt.Sprintf("%v", castInst.Name)] = inst[i]
	}
	return
}

// New creates a new Instance slice
func New(firstInstance *hostProvider.Instance) *Instances {
	if firstInstance != nil {
		return &Instances{*firstInstance}
	}
	return &Instances{}
}

// AddInstance will add an instance to the Instances slice
func AddInstance(list Instances, instance hostProvider.Instance) Instances {
	return append(list, instance)
}

// RemoveInstance will remove an instance from the slice
func RemoveInstance(list Instances, instance hostProvider.Instance) Instances {
	newList := make(Instances, len(list)-1)
	for i := range list {
		castInst := list[i].(hostProvider.GcloudInstance)
		castOther := instance.(hostProvider.GcloudInstance)
		if castInst.ID != castOther.ID {
			newList[i] = list[i]
		}
	}
	return newList
}
