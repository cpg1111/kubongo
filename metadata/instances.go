package metadata

import (
    "fmt"

	"github.com/cpg1111/kubongo/hostProvider"
)

type Instances []hostProvider.GcloudInstance

func (inst Instances) ToMap() (instanceMap map[string]hostProvider.GcloudInstance) {
	for i := range inst {
		instanceMap[fmt.Sprintf("%v", inst[i].Id)] = inst[i]
	}
	return
}

func New(firstInstance *hostProvider.GcloudInstance) Instances {
	if firstInstance != nil {
		return Instances{*firstInstance}
	}
	return Instances{}
}

func AddInstance(list Instances, instance hostProvider.GcloudInstance) Instances {
	return append(list, instance)
}

func RemoveInstance(list Instances, instance hostProvider.GcloudInstance) Instances {
	newList := make(Instances, len(list)-1)
	for i := range list {
		if list[i].Id != instance.Id {
			newList[i] = list[i]
		}
	}
	return newList
}
