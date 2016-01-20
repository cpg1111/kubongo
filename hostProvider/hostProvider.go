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
