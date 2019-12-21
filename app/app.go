package app

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"
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

// Application configuration
type Config struct {
	Verbose          bool
	Labels           []string
	FolderAnnotation string
	Folder           string
	Namespaces       []string
	ConfigMaps       bool
	Secrets          bool
}

// Execute the app
func Execute(config *Config) {
	var kubeconfig *string
	var clientconfig *rest.Config

	// Check if we are running inside the cluster
	// All the PODs have a service account and a mounted volume
	// https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#service-account-automation
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if os.IsNotExist(err) {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		clientconfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		clientconfig, err = rest.InClusterConfig()
	}

	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(clientconfig)
	if err != nil {
		panic(err.Error())
	}

	cmClient := clientset.CoreV1().ConfigMaps("")

	cmWatchList := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return cmClient.List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return cmClient.Watch(metav1.ListOptions{})
		},
	}

	// All namespaces
	// watchlist := cache.NewListWatchFromClient(clientset.RESTClient(),
	// 	"configmap", "", fields.Everything())
	_, controller := cache.NewInformer(
		cmWatchList,
		&apiv1.ConfigMap{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Infof("configmap added: %s", obj)
			},
			DeleteFunc: func(obj interface{}) {
				log.Infof("configmap deleted: %s", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Infof("configmap changed")
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
