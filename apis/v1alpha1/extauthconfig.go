package v1alpha1

type ExtAuthConfig struct {
	Name              string `yaml:"name"`
	ExtAuthConfigSpec `yaml:"spec"`
}

type ExtAuthConfigSpec struct {
	RequiredTokenByRoutes map[string]string `yaml:"requiredTokenByRoute"`
	AuthorizationToken    string            `yaml:"authToken"`
	AdditionalHeaders     map[string]string `yaml:"additionalHeaders"`
}
