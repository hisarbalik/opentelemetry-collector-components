package k8sconfig

import (
	"fmt"
	"net/http"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type AuthType string

const (
	// AuthTypeServiceAccount means to use the built-in service account that
	// K8s automatically provisions for each pod.
	AuthTypeServiceAccount AuthType = "serviceAccount"
	// AuthTypeKubeConfig uses local credentials like those used by kubectl.
	AuthTypeKubeConfig AuthType = "kubeConfig"
)

var authTypes = map[AuthType]bool{
	AuthTypeServiceAccount: true,
	AuthTypeKubeConfig:     true,
}

// APIConfig contains options relevant to connecting to the K8s API
type APIConfig struct {
	// How to authenticate to the K8s API server.  This can be one of `none`
	// (for no auth), `serviceAccount` (to use the standard service account
	// token provided to the agent pod), or `kubeConfig` to use credentials
	// from `~/.kube/config`.
	AuthType AuthType `mapstructure:"auth_type"`

	// When using auth_type `kubeConfig`, override the current context.
	Context string `mapstructure:"context"`
}

// Validate validates the K8s API config
func (c APIConfig) Validate() error {
	if !authTypes[c.AuthType] {
		return fmt.Errorf("invalid authType for kubernetes: %v", c.AuthType)
	}

	return nil
}

// CreateRestConfig creates a Kubernetes API config from user configuration.
func CreateRestConfig(apiConf APIConfig) (*rest.Config, error) {
	var authConf *rest.Config
	var err error

	authType := apiConf.AuthType

	if authType != AuthTypeKubeConfig {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			return nil, fmt.Errorf("unable to load k8s config, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
		}
	}

	switch authType {
	case AuthTypeKubeConfig:
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		if apiConf.Context != "" {
			configOverrides.CurrentContext = apiConf.Context
		}
		authConf, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules, configOverrides).ClientConfig()

		if err != nil {
			return nil, fmt.Errorf("error connecting to k8s with auth_type=%s: %w", AuthTypeKubeConfig, err)
		}
	}

	authConf.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		// Don't use system proxy settings since the API is local to the
		// cluster
		if t, ok := rt.(*http.Transport); ok {
			t.Proxy = nil
		}
		return rt
	}

	return authConf, nil
}

func MakeClient(apiConf APIConfig) (kubernetes.Interface, error) {
	if err := apiConf.Validate(); err != nil {
		return nil, err
	}

	authConf, err := CreateRestConfig(apiConf)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(authConf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
