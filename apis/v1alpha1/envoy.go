package v1alpha1

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type Spec struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type Listener struct {
	Name     string  `yaml:"name"`
	Address  string  `yaml:"address"`
	Port     uint32  `yaml:"port"`
	Routes   []Route `yaml:"routes"`
	CertFile string  `yaml:"certFile"`
	KeyFile  string  `yaml:"keyFile"`
}

type Route struct {
	Name        string `yaml:"name"`
	Prefix      string `yaml:"prefix"`
	Header      string `yaml:"header"`
	HostRewrite string `yaml:"hostRewrite"`
}

type Cluster struct {
	Name      string     `yaml:"name"`
	IsHTTPS   bool       `yaml:"isHTTPS"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}
