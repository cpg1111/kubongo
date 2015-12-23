package hostProvider

type instance interface {
	GetInternalIP() string
}

type HostProvider interface {
	GetServers(namespace string) ([]instance, error)
	CreateServer(namespace, zone, name, machineType, sourceImage, source string) ([]byte, error)
}
