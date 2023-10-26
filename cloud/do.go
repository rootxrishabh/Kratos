package do

import (
	"context"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/emicklei/go-restful/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Create(c kubernetes.Interface, spec map[string]interface{}) (string, error) {
	token, err := getToken(c, spec["tokenSecret"].(string))
	if err != nil {
		return "", err
	}
	log.Printf("token from cloud file %s", token)
	client := godo.NewFromToken(token)
	nodePools := spec["nodePools"].([]interface{})
	nodePool := nodePools[0].(map[string]interface{})

	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec["name"].(string),
		RegionSlug:  spec["region"].(string),
		VersionSlug: spec["version"].(string),
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			&godo.KubernetesNodePoolCreateRequest{
				Size:  nodePool["size"].(string),
				Name:  nodePool["name"].(string),
				Count: nodePool["count"].(int),
			},
		},
	}

	cluster, _, err := client.Kubernetes.Create(context.Background(), request)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}

func ClusterState(c kubernetes.Interface, spec map[string]interface{}, id string) (string, error) {
	token, err := getToken(c, spec["tokenSecret"].(string))
	if err != nil {
		return "", err
	}

	client := godo.NewFromToken(token)

	cluster, _, err := client.Kubernetes.Get(context.Background(), id)
	return string(cluster.Status.State), err
}

func getToken(client kubernetes.Interface, sec string) (string, error) {
	namespace := strings.Split(sec, "/")[0]
	name := strings.Split(sec, "/")[1]
	s, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(s.Data["token"]), nil
}
