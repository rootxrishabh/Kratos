package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/rootxrishabh/DynamicClient/pkg/rishabh.dev/v1alpha1"
	"github.com/rootxrishabh/dynamic-client/controller"
	"github.com/rootxrishabh/dynamic-client/pkg/apis/rishabh.dev/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/rishabh/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		// handle error
		fmt.Printf("erorr %s building config from flags\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s, getting inclusterconfig", err.Error())
		}
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, getting dyn client\n", err.Error())
	}

	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynClient, 10*time.Minute)

	c := controller.NewController(dynClient, infFactory)
	infFactory.Start(make(<-chan struct{}))
	c.run(make(<-chan struct{}))
	fmt.Printf("the concrete type that we got is %+v\n", k)
}