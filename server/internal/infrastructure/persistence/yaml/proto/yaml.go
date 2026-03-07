package proto

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed services.yaml
var serviceBytes []byte

// HTTP represents the HTTP configuration for a method
type HTTP struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Body   string `yaml:"body,omitempty"`
}

type Method struct {
	Name     string `yaml:"name"`
	Request  string `yaml:"request"`
	Response string `yaml:"response"`
	HTTP     HTTP   `yaml:"http"`
}

type Service struct {
	File    string   `yaml:"file"`
	Service string   `yaml:"service"`
	Methods []Method `yaml:"methods"`
}

type Services struct {
	Services []Service `yaml:"services"`
}

func SelectAllServices() []Service {
	var services Services
	err := yaml.Unmarshal(serviceBytes, &services)
	if err != nil {
		// In a real application, we would handle this error properly
		// For now, just return an empty slice
		return []Service{}
	}
	return services.Services
}
