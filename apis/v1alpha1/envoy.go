package v1alpha1

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type Spec struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
	ExtAuthz  `yaml:"ext-authz"`
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

type DiscoveryType string

const (
	LogicalDNS DiscoveryType = "LogicalDNS"
	StrictDNS                = "StrictDNS"
	Static                   = "Static"
)

type Cluster struct {
	Name          string `yaml:"name"`
	IsHTTPS       bool   `yaml:"isHTTPS"`
	DiscoveryType `yaml:"discoveryType"`
	Endpoints     []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}

type ExtAuthz struct {
	RouteKey string          `yaml:"routeKey"`
	Routes   []ExtAuthzRoute `yaml:"routes"`
}

type ExtAuthzRoute struct {
	Cluster           string            `yaml:"cluster"`
	RequiredToken     string            `yaml:"requiredToken"`
	OutgoingToken     string            `yaml:"outgoingToken"`
	RewriteHost       string            `yaml:"rewriteHost"`
	RewriteRoute      string            `yaml:"rewriteRoute"`
	AdditionalHeaders map[string]string `yaml:"additionalHeaders"`
}
