package zookeeper

const (
	allEths  = "0.0.0.0"
	envPodIP = "POD_IP"
)

// Options is the config item with the given options for zookeeper.
type Options struct {
	ListenOn       string
	Hosts          []string
	BasePath       string
	ServiceName    string
	Weight         float64
	Metadata       map[string]string
	SessionTimeout int
	ConnectTimeout int
}

type Option func(*Options)

func NewZookeeperConfig(listenOn string, opts ...Option) *Options {
	options := &Options{
		ListenOn:       listenOn,
		BasePath:       "/services",
		Metadata:       make(map[string]string),
		SessionTimeout: 60,
		ConnectTimeout: 5,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WithWeight(weight float64) Option {
	return func(o *Options) {
		o.Weight = weight
	}
}

func WithHosts(hosts []string) Option {
	return func(o *Options) {
		o.Hosts = hosts
	}
}

func WithBasePath(basePath string) Option {
	return func(o *Options) {
		o.BasePath = basePath
	}
}

func WithServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

func WithMetadata(metadata map[string]string) Option {
	return func(o *Options) {
		o.Metadata = metadata
	}
}

func WithSessionTimeout(timeout int) Option {
	return func(o *Options) {
		o.SessionTimeout = timeout
	}
}

func WithConnectTimeout(timeout int) Option {
	return func(o *Options) {
		o.ConnectTimeout = timeout
	}
}
