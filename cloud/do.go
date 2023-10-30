package do

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The following functions are used to create and delete a cluster on Digital Ocean, due to specific nature of output of dynamic client, deletion of cluster could not be implemented using token available in the secrets and hence a new token is acquired from os.Getenv to delete the cluster.
// Currently getToken is still in use for creating the cluster, but will be replaced with os.Getenv in future or any other method to acquire token.

func Create(c kubernetes.Interface, spec map[string]interface{}) (string, error) {
	token, err := getToken(c, spec["tokenSecret"].(string))
	if err != nil {
		return "", err
	}
	client := godo.NewFromToken(token)
	nodePools := spec["nodePools"].([]interface{})
	nodePool := nodePools[0].(map[string]interface{})
	ct := nodePool["count"].(string)
	count, _ := strconv.Atoi(ct)
	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec["name"].(string),
		RegionSlug:  spec["region"].(string),
		VersionSlug: spec["version"].(string),
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			&godo.KubernetesNodePoolCreateRequest{
				Size:  nodePool["size"].(string),
				Name:  nodePool["name"].(string),
				Count: count,
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

func DeleteCluster(c kubernetes.Interface, name string) (string, error){
	token := os.Getenv("DIGITAL_OCEAN_TOKEN")

    client := godo.NewFromToken(token)
    ctx := context.TODO()
	ClusterID := retriveClusterID(c, token, name)
    conformation, err := client.Kubernetes.Delete(ctx, ClusterID)
	if err != nil {
		return "", err
	}
	return  "Cluster Deleted" + string(conformation.Status), nil
}

func retrieveClusters(c kubernetes.Interface, token string) ([]*godo.KubernetesCluster, error){

    client := godo.NewFromToken(token)
    ctx := context.TODO()

    opt := &godo.ListOptions{
        Page:    1,
        PerPage: 200,
    }

    clusters, _, err := client.Kubernetes.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func retriveClusterID(client kubernetes.Interface, token string, name string) string{
	cluster, _ := retrieveClusters(client, token)
	for _, c := range cluster {
		if(c.Name == name){
			log.Println("true")
			return c.ID
		}
	}
	return "Error retrieving cluster ID"
}