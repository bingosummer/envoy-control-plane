package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/weinong/envoy-control-plane/apis/v1alpha1"
	"gopkg.in/yaml.v2"
)

// ParseEnvoyConfig takes in a yaml envoy config and returns a typed version
func ParseEnvoyConfig(file string) (*v1alpha1.EnvoyConfig, error) {
	var config v1alpha1.EnvoyConfig

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading YAML file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
