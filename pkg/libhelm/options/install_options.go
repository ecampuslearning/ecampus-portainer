package options

import "time"

type InstallOptions struct {
	Name                    string
	Chart                   string
	Version                 string
	Namespace               string
	Repo                    string
	Wait                    bool
	ValuesFile              string
	PostRenderer            string
	Atomic                  bool
	Timeout                 time.Duration
	KubernetesClusterAccess *KubernetesClusterAccess

	// Optional environment vars to pass when running helm
	Env []string
}
