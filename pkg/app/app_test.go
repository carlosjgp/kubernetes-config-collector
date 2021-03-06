package app

import (
	"context"
	"testing"
	"time"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/handler"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

var (
	ChannelHandler = func(events chan *handler.Event) cache.ResourceEventHandlerFuncs {
		return cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c := obj.(*apiv1.ConfigMap)
				events <- &handler.Event{
					Object: c.GetObjectMeta(),
					Action: "add",
				}
			},
			DeleteFunc: func(obj interface{}) {
				c := obj.(*apiv1.ConfigMap)
				events <- &handler.Event{
					Object: c.GetObjectMeta(),
					Action: "delete",
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				c := newObj.(*metav1.ObjectMeta)
				events <- &handler.Event{
					Object: c.GetObjectMeta(),
					Action: "update",
				}
			},
		}
	}
)

// TestFakeClient
func TestAddWatchedConfigMap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the fake client.
	client := fake.NewSimpleClientset()

	cmClient := client.CoreV1().ConfigMaps("")

	events := make(chan *handler.Event, 1)

	controller := NewConfigMapInformer(cmClient, metav1.ListOptions{}, ChannelHandler(events))

	informers := informers.NewSharedInformerFactory(client, 0)
	informers.Start(ctx.Done())

	stop := make(chan struct{})
	go controller.Run(stop)

	// This is not required in tests, but it serves as a proof-of-concept by
	// ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	cache.WaitForCacheSync(ctx.Done(), controller.HasSynced)

	// Inject an event into the fake client.
	cm := &apiv1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "my-cm"}}
	t.Logf("Create new configmap: %s", cm)
	_, err := cmClient.Create(cm)
	if err != nil {
		t.Fatalf("error injecting resource add: %v", err)
	}

	select {
	case e := <-events:
		t.Logf("Got event from channel: %s", e)
		if e.GetName() != cm.GetName() {
			t.Errorf("Not the expected event resource: %s", e)
		}
		if e.Action != "add" {
			t.Errorf("Not the expected event action: %s", e)
		}
	case <-time.After(wait.ForeverTestTimeout):
		t.Error("Informer did not get the added event")
	}
}
