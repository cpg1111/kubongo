package hostProvider

type instance interface {
	GetInternalIP() string
}

type HostProvider interface {
	GetServerNames(namespace string) []instance, error
}
