package handler

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	configMapAddedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_cm_add_total",
		Help: "The total number of ConfigMaps added",
	})
	configMapUpdatedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_cm_update_total",
		Help: "The total number of ConfigMaps updated",
	})
	configMapDeletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_cm_delete_total",
		Help: "The total number of ConfigMaps deleted",
	})
	keyAddedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_add_total",
		Help: "The total number of files added",
	})
	keyUpdatedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_update_total",
		Help: "The total number of files updated",
	})
	keyDeletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_delete_total",
		Help: "The total number of files deleted",
	})
	keyAddedErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_add_error_total",
		Help: "The total number of errors adding files",
	})
	keyUpdatedErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_update_error_total",
		Help: "The total number of errors updating files",
	})
	keyDeletedErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_collector_key_delete_error_total",
		Help: "The total number of errors deleting files",
	})
)

/**
Events managed by the handler
*/
type Event struct {
	metav1.Object
	Action string
}

// configuration
type Config struct {
	FolderAnnotation string
	Folder           string
}

func ResolveFolder(config Config, meta metav1.Object) string {
	log.Debugf("Annotations: %s", meta.GetAnnotations())
	var folder string
	// Get folder annotation
	if f, ok := meta.GetAnnotations()[config.FolderAnnotation]; ok {
		// or default to config flag
		log.Infof("Using custom folder: %s", f)
		folder = f
	} else {
		log.Info("Folder annotation not found")
		folder = config.Folder
	}
	return strings.TrimSuffix(folder, "/")
}

func ResolveFilePath(config Config, meta metav1.Object, file string) string {
	return fmt.Sprintf("%s/%s", ResolveFolder(config, meta), file)
}

// Create a new FileHandler
func NewFileHandler(config Config) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c := obj.(*apiv1.ConfigMap)
			name := c.GetObjectMeta().GetName()
			log.Infof("configmap added: %s", name)
			configMapAddedCounter.Inc()

			for key, element := range c.Data {
				log.Infof("Processing key: %s/%s", name, key)
				err := ioutil.WriteFile(ResolveFilePath(config, c.GetObjectMeta(), key), []byte(element), 0644)
				if err == nil {
					keyAddedCounter.Inc()
				} else {
					log.Error(err)
					keyAddedErrorCounter.Inc()
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			c := obj.(*apiv1.ConfigMap)
			name := c.GetObjectMeta().GetName()
			log.Infof("configmap deleted: %s", name)

			configMapDeletedCounter.Inc()

			for key := range c.Data {
				log.Infof("Processing key: %s/%s", name, key)
				err := os.Remove(ResolveFilePath(config, c.GetObjectMeta(), key))
				if err == nil {
					keyDeletedCounter.Inc()
				} else {
					log.Error(err)
					keyDeletedErrorCounter.Inc()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			cmOld := oldObj.(*apiv1.ConfigMap)
			cmNew := newObj.(*apiv1.ConfigMap)

			name := cmOld.GetObjectMeta().GetName()
			log.Infof("configmap changed: %s", name)

			configMapUpdatedCounter.Inc()

			// Delete missing keys
			for key := range cmOld.Data {
				log.Infof("Processing key: %s/%s", name, key)
				if _, ok := cmNew.Data[key]; ok {
					log.Infof("Deleting: %s/%s", name, key)

					err := os.Remove(ResolveFilePath(config, cmOld.GetObjectMeta(), key))
					if err == nil {
						keyDeletedCounter.Inc()
					} else {
						log.Error(err)
						keyDeletedErrorCounter.Inc()
					}
				}
			}
			// Add/Update files on the new ConfigMap
			for key, element := range cmNew.Data {
				log.Infof("Processing key: %s/%s", name, key)
				log.Infof("Adding/Updating: %s/%s", name, key)

				err := ioutil.WriteFile(ResolveFilePath(config, cmOld.GetObjectMeta(), key), []byte(element), 0644)
				if err == nil {
					keyAddedCounter.Inc()
				} else {
					log.Error(err)
					keyAddedErrorCounter.Inc()
				}
			}
		},
	}
}
