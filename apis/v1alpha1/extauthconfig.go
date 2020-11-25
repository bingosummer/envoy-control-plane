package v1alpha1

type ExtAuthConfig struct {
	Name              string `yaml:"name"`
	ExtAuthConfigSpec `yaml:"spec"`
}

type ExtAuthConfigSpec struct {
	RequiredBearerToken string            `yaml:"requiredToken"`
	AuthorizationToken  string            `yaml:"authToken"`
	AdditionalHeaders   map[string]string `yaml:"addtionalHeaders"`
}
