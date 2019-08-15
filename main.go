package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var n17RawAPIKey string

func notify17(httpClient *http.Client, title string, content string) {
	log.Printf("Triggering notification: [title=" + title + "] [content=" + content + "]")

	formData := url.Values{
		"title":   {title},
		"content": {content},
	}

	hookURL := fmt.Sprintf("https://hook.notify17.net/api/raw/%s", n17RawAPIKey)
	resp, err := http.PostForm(hookURL, formData)
	if err != nil {
		log.Printf("ERROR, failed to trigger notification: %s", err)
	}

	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)
}

func main() {
	// Check basic notify17 configuration
	n17RawAPIKey = os.Getenv("N17_RAW_API_KEY")
	if n17RawAPIKey == "" {
		log.Fatalf("Missing ENV variable N17_RAW_API_KEY")
	}

	httpClient := &http.Client{}

	var k8sConfig *rest.Config
	var err error
	kubeConfigPath := os.Getenv("KUBE_CONFIG_PATH")
	if kubeConfigPath != "" {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("failed to load k8s config: %s", err)
		}
	} else {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("failed to load k8s in-cluster config: %s", err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("failed to create k8s client: %s", err)
	}

	api := clientSet.CoreV1()

	// Fetches initial nodes, to prevent useless notifications when the watch initially starts up
	initialNodes, err := api.Nodes().List(metaV1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	var initialNodesUIDs []types.UID
	for _, node := range initialNodes.Items {
		initialNodesUIDs = append(initialNodesUIDs, node.UID)
	}

	listOptions := metaV1.ListOptions{}
	watcher, err := api.Nodes().Watch(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	ch := watcher.ResultChan()

	for event := range ch {
		node, ok := event.Object.(*v1.Node)
		if !ok {
			log.Fatal("unexpected type")
		}

		switch event.Type {
		case watch.Added:
			foundInitial := false
			for _, initialNodeUID := range initialNodesUIDs {
				if initialNodeUID == node.UID {
					foundInitial = true
					break
				}
			}
			// Ignore added notification for an initial node
			if foundInitial {
				continue
			}

			notify17(httpClient, "Node added", fmt.Sprintf("Node %s has been added", node.Name))
		case watch.Deleted:
			for idx, initialNodeUID := range initialNodesUIDs {
				if initialNodeUID == node.UID {
					// Remove node from initials list
					initialNodesUIDs = append(initialNodesUIDs[:idx], initialNodesUIDs[idx+1:]...)
				}
			}

			notify17(httpClient, "Node deleted", fmt.Sprintf("Node %s has been deleted", node.Name))
		}
	}
}
