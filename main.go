package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/rootxrishabh/dynamic-client/controller"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
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

	staticClient, err := kubernetes.NewForConfig(config)

	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynClient, 10*time.Minute)

	c := controller.NewController(dynClient, infFactory, staticClient)
	infFactory.Start(make(<-chan struct{}))
	c.Run(make(<-chan struct{}))
}
