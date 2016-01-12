package hostProvider

type Instance interface {
	GetInternalIP() string
}

type HostProvider interface {
	GetServers(namespace string) ([]Instance, error)
	GetServer(project, zone, name string) (Instance, error)
	CreateServer(namespace, zone, name, machineType, sourceImage, source string) ([]byte, error)
	DeleteServer(namespace, zone, name string) error
}
