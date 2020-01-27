package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func GetClient() (*kubernetes.Clientset, error) {
	var clientconfig *rest.Config

	// Check if we are running inside the cluster
	// All the PODs have a service account and a mounted volume
	// https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#service-account-automation
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if os.IsNotExist(err) {
		kubeconfig := ""
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		clientconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		clientconfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(clientconfig)
}
