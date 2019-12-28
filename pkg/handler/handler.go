package handler

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	log "github.com/sirupsen/logrus"
)

type HandlerEvent struct {
	metav1.Object
	Action string
}

var (
	LogHandler = cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Infof("configmap added: %s", obj)
		},
		DeleteFunc: func(obj interface{}) {
			log.Infof("configmap deleted: %s", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Infof("configmap changed")
		},
	}
	ChannelHandler = func(events chan *HandlerEvent, t *testing.T) cache.ResourceEventHandlerFuncs {
		return cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c := obj.(*apiv1.ConfigMap)
				t.Logf("Add event to channel: add/%s", c)
				events <- &HandlerEvent{
					Object: c.GetObjectMeta(),
					Action: "add",
				}
			},
			DeleteFunc: func(obj interface{}) {
				c := obj.(*apiv1.ConfigMap)
				t.Logf("Add event to channel: delete/%s", c)
				events <- &HandlerEvent{
					Object: c.GetObjectMeta(),
					Action: "delete",
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				c := newObj.(*metav1.ObjectMeta)
				t.Logf("Add event to channel: update/%s", c)
				events <- &HandlerEvent{
					Object: c.GetObjectMeta(),
					Action: "update",
				}
			},
		}
	}
)
