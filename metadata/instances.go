package metadata

import (
	"fmt"

	"github.com/cpg1111/kubongo/hostProvider"
)

type Instances []hostProvider.Instance

func (inst Instances) ToMap() (instanceMap map[string]hostProvider.Instance) {
	for i := range inst {
		castInst := inst[i].(hostProvider.GcloudInstance)
		instanceMap[fmt.Sprintf("%v", castInst.Name)] = inst[i]
	}
	return
}

func New(firstInstance *hostProvider.Instance) *Instances {
	if firstInstance != nil {
		return &Instances{*firstInstance}
	}
	return &Instances{}
}

func AddInstance(list Instances, instance hostProvider.Instance) Instances {
	return append(list, instance)
}

func RemoveInstance(list Instances, instance hostProvider.Instance) Instances {
	newList := make(Instances, len(list)-1)
	for i := range list {
		castInst := list[i].(hostProvider.GcloudInstance)
		castOther := instance.(hostProvider.GcloudInstance)
		if castInst.Id != castOther.Id {
			newList[i] = list[i]
		}
	}
	return newList
}
