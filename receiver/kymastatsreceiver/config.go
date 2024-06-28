package kymastatsreceiver

import (
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	k8sconfig.APIConfig            `mapstructure:",squash"`
	makeClient                     func() (k8s.Interface, error)
	makeDynamicClient              func() (dynamic.Interface, error)
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
}

func (cfg *Config) Validate() error {
	err := cfg.ControllerConfig.Validate()
	if err != nil {
		return err
	}
	return cfg.APIConfig.Validate()
}

func (cfg *Config) getK8sClient() (k8s.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeClient()
	}
	return k8sconfig.MakeClient(cfg.APIConfig)
}

func (cfg *Config) getK8sDynamicClient() (dynamic.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeDynamicClient()
	}
	return k8sconfig.MakeDynamicClient(cfg.APIConfig)
}
