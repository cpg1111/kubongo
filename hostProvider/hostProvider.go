package hostProvider

type HostProvider interface {
	GetServerNames(namespace string)
}
