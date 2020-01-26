package handler

import (
	"fmt"
	"io/ioutil"
	"os"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LogHandler = cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Infof("configmap added: %s", obj)
		},
		DeleteFunc: func(obj interface{}) {
			log.Infof("configmap deleted: %s", obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			log.Infof("configmap changed: %s", oldObj)
		},
	}

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
		Name: "config_collector_key_add_total",
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

type HandlerEvent struct {
	metav1.Object
	Action string
}

func NewFileHandler(folder string) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c := obj.(*apiv1.ConfigMap)
			name := c.GetObjectMeta().GetName()
			log.Infof("configmap added: %s", name)
			configMapAddedCounter.Inc()

			for key, element := range c.Data {
				log.Infof("Processing key: %s/%s", name, key)
				err := ioutil.WriteFile(fmt.Sprintf("%s/%s", folder, key), []byte(element), 0644)
				keyAddedCounter.Inc()
				if err != nil {
					log.Fatal(err)
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
				keyDeletedCounter.Inc()
				err := os.Remove(fmt.Sprintf("%s/%s", folder, key))
				if err != nil {
					log.Fatal(err)
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
					keyDeletedCounter.Inc()

					err := os.Remove(fmt.Sprintf("%s/%s", folder, key))
					if err != nil {
						log.Fatal(err)
						keyDeletedErrorCounter.Inc()
					}
				}
			}
			// Add/Update files on the new ConfigMap
			for key, element := range cmNew.Data {
				log.Infof("Processing key: %s/%s", name, key)
				log.Infof("Adding/Updating: %s/%s", name, key)
				keyAddedCounter.Inc()

				err := ioutil.WriteFile(fmt.Sprintf("%s/%s", folder, key), []byte(element), 0644)
				if err != nil {
					log.Fatal(err)
					keyAddedErrorCounter.Inc()
				}
			}
		},
	}
}
