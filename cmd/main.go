package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ialidzhikov/efk-stress-test/pkg/dto"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	seedKubeconfig := flag.String("seed-kubeconfig", "", "Seed kubeconfig")
	shootNamespace := flag.String("shoot-namespace", "", "Shoot namespace")
	documents := flag.Int("log-messages", 10000, "Number of log messages to log by client")
	clients := flag.Int("clients", 2, "Number of clients")
	elasticHost := flag.String("elastic-host", "http://localhost:9200", "Elasticsearch host")

	flag.Parse()

	waitInterval := 60

	config, err := clientcmd.BuildConfigFromFlags("", *seedKubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	file, err := os.Open("./assets/deployment.yaml")
	if err != nil {
		panic(err)
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 1024)
	var deployment appsv1.Deployment
	decoder.Decode(&deployment)
	deployment.Spec.Template.Spec.Containers[0].Env[0].Value = strconv.Itoa(*documents)

	deploymentsClient := clientset.AppsV1().Deployments(*shootNamespace)
	result, err := deploymentsClient.Create(&deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	for i := 0; i < *clients; i++ {
		replicas := int32(i + 1)
		deployment.Spec.Replicas = &replicas
		_, err := deploymentsClient.Update(&deployment)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Updated the %q replicas to %d.\n", deployment.GetObjectMeta().GetName(), replicas)

		fmt.Printf("Sleeping %d seconds.\n", waitInterval)
		time.Sleep(time.Duration(waitInterval) * time.Second)
	}

	actualHits := 0
	expectedHits := *clients * (*documents + 1)
	for actualHits < expectedHits {
		actualHits = int(getTotalHits(elasticHost))
		fmt.Printf("Expected %d hits, got - %d\n", expectedHits, actualHits)
		if actualHits == expectedHits {
			break
		}

		time.Sleep(time.Duration(10) * time.Second)
	}
}

func getTotalHits(elasticHost *string) uint64 {
	now := time.Now()
	formattedDate := fmt.Sprintf("%d.%02d.%d", now.Year(), now.Month(), now.Day())
	url := fmt.Sprintf("%s/logstash-%s/_search?q=kubernetes.pod_name:%s", *elasticHost, formattedDate, "logger")
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	searchResponse := &dto.SearchResponse{}
	json.NewDecoder(response.Body).Decode(searchResponse)

	return searchResponse.Hits.Total
}
