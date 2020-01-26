package app

import (
	"context"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/handler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/tools/cache"
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
func Execute(clientset *kubernetes.Clientset, config *Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add default namespace
	var clients []corev1.ConfigMapInterface
	if len(config.Namespaces) == 0 {
		config.Namespaces = []string{""}
	}

	// Create a client per namespace
	for _, ns := range config.Namespaces {
		clients = append(clients, clientset.CoreV1().ConfigMaps(ns))
	}

	// Create a controller per client
	var controllers []cache.Controller
	for _, client := range clients {
		controllers = append(controllers,
			NewConfigMapInformer(client, metav1.ListOptions{}, handler.NewFileHandler(config.Folder)))
	}

	informers := informers.NewSharedInformerFactory(clientset, 0)
	informers.Start(ctx.Done())

	// Start all the controllers
	stop := make(chan struct{})
	for _, ctrl := range controllers {
		go ctrl.Run(stop)

		// This is not required in tests, but it serves as a proof-of-concept by
		// ensuring that the informer goroutine have warmed up and called List before
		// we send any events to it.
		cache.WaitForCacheSync(ctx.Done(), ctrl.HasSynced)
	}

	for {
		time.Sleep(time.Second)
	}
}

func NewConfigMapInformer(
	client corev1.ConfigMapInterface,
	filteringOptions metav1.ListOptions,
	handler cache.ResourceEventHandlerFuncs) cache.Controller {

	return NewInformer(
		func(options metav1.ListOptions) (runtime.Object, error) {
			return client.List(filteringOptions)
		},
		func(options metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(filteringOptions)
		},
		&apiv1.ConfigMap{},
		handler)
}

func NewInformer(
	resourceListFunc cache.ListFunc,
	resourceWatchFunc cache.WatchFunc,
	resource runtime.Object,
	handler cache.ResourceEventHandlerFuncs) cache.Controller {

	controller := &cache.ListWatch{
		ListFunc:  resourceListFunc,
		WatchFunc: resourceWatchFunc,
	}

	_, informer := cache.NewInformer(
		controller,
		resource,
		time.Second*0,
		handler,
	)
	return informer
}
