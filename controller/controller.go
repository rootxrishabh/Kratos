package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type contrller struct {
	client   dynamic.Interface
	informer cache.SharedIndexInformer
}

func NewController(client dynamic.Interface, dynInformer dynamicinformer.DynamicSharedInformerFactory) *contrller {
	inf := dynInformer.ForResource(schema.GroupVersionResource{
		Group:    "rishabh.dev",
		Version:  "v1alpha1",
		Resource: "kratos",
	}).Informer()

	inf.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("resource was created")
			},
		},
	)

	return &contrller{
		client:   client,
		informer: inf,
	}

}

func (c *contrller) run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *contrller) worker() {
	for c.processItem() {

	}
}

func (c *contrller) processItem() bool {
	return true
}