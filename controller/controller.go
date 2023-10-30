package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kanisterio/kanister/pkg/poll"
	do "github.com/rootxrishabh/dynamic-client/cloud"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

var cloudResource = schema.GroupVersionResource{
	Group:    "rishabh.dev",
	Version:  "v1alpha1",
	Resource: "kratoses",
}

type contrller struct {
	dynamicClient dynamic.Interface
	informer      cache.SharedIndexInformer
	stopper       chan struct{}
	queue         workqueue.RateLimitingInterface
	staticClient  kubernetes.Interface
}

func NewController(dynamicClient dynamic.Interface, dynInformer dynamicinformer.DynamicSharedInformerFactory, staticClient kubernetes.Interface) *contrller {
	informer := dynInformer.ForResource(cloudResource).Informer()

	c := &contrller{
		dynamicClient: dynamicClient,
		informer:      informer,
		stopper:       make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Kratos"),
		staticClient:  staticClient,
	}

	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
			// UpdateFunc: c.handleUpdate,
		},
	)

	return c
}

func (c *contrller) Run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *contrller) worker() {
	for c.processItem() {
		// This loops helps running the processItem function as long as it returns true
	}
}

func (c *contrller) processItem() bool {
	item, shutDown := c.queue.Get()
	if shutDown {
		// logs as well
		return false
	}

	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name, error %s\n", err.Error())
		return false
	}

	kratos, err := c.dynamicClient.Resource(cloudResource).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err := do.DeleteCluster(c.staticClient, name)
			if err != nil {
				log.Printf("error %s, deleting the cluster", err.Error())
				return false
			} else {
				return true
				}
			}
			log.Printf("error %s, Getting the kluster resource from lister", err.Error())
			return false
		}

		spec := kratos.Object["spec"].(map[string]interface{})
		clusterID, err := do.Create(c.staticClient, spec)
		if err != nil {
			log.Printf("error %s, creating the cluster", err.Error())
		}
		
		if clusterID == "" {
		log.Printf("clusterID is nil, cluster creation failed as clusters are already created")
		return false
		} else {
		// Getting cluster state to wait for cluster to be running
		err = c.waitForCluster(spec, clusterID)
		if err != nil {
			log.Printf("error %s, waiting for cluster to be running or cluster already exists", err.Error())
		}
	}
	return true
}

func (c *contrller) waitForCluster(spec map[string]interface{}, clusterID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := do.ClusterState(c.staticClient, spec, clusterID)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}

		return false, nil
	})
}

func (c *contrller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.queue.Add(obj)
}

func (c *contrller) handleDel(obj interface{}) {
	log.Println("handleDelete was called")
	c.queue.Add(obj)
}

// func (c *contrller) handleUpdate(ondObj, newObj interface{}) {
// 	// get the kluster resource
// 	kluster, ok := newObj.(*v1alpha1.Kluster)
// 	if !ok {
// 		log.Printf("can not convert newObj to kluster resource\n")
// 		return
// 	}
// 	ctx := context.Background()
// 	// if the finalizer is set or not
// 	// check if the cluster has prod namespace
// 	_, err := c.staticClient.CoreV1().Namespaces().Get(ctx, protectedNS, metav1.GetOptions{}) // this would requrie role change to be able to get ns
// 	if err == nil {
// 		// prod ns is available, do nothing
// 		return
// 	}
// 	// if it has, do nothing
// 	// otherwise, remove finalizer `viveksingh.dev/prod-protection` from resource
// 	// if we are here, there is an err set, to be explicit you can check this says resource not found
// 	k := kluster.DeepCopy()
// 	finals := []string{}
// 	for _, f := range k.Finalizers {
// 		if f == klusterFinalizer {
// 			continue
// 		}
// 		finals = append(finals, f)
// 	}
// 	k.Finalizers = finals

// 	// change role to be able to update the kluster resource
// 	if _, err = c.klient.ViveksinghV1alpha1().Klusters(k.Namespace).Update(ctx, k, metav1.UpdateOptions{}); err != nil {
// 		log.Printf("Update of the kluster resource failed: %s\n", err.Error())
// 		return
// 	}
// 	log.Println("Finalizer was removed from the resource")
// }
