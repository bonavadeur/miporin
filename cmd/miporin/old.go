package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/scraper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type decideInNode struct {
	Node1 int32 `json:"node1"`
	Node2 int32 `json:"node2"`
	Node3 int32 `json:"node3"`
}

func TriggerChangeCRD() {
	dynamicClient, _ := dynamic.NewForConfig(KUBECONFIG)

	namespace := "default"
	name := "hello"

	gvr := schema.GroupVersionResource{
		Group:    "networking.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "ingresses",
	}

	crd, _ := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	labels := crd.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["bonachange"] = generateRandomString(10)
	crd.SetLabels(labels)

	_, err := dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), crd, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
}

func makeHTTPRequest() {
	url := "http://net-kourier-controller.knative-serving.svc.cluster.local:1323/api/percentage"

	jsonData, err := json.Marshal(METRIC_SCRAPED)
	if err != nil {
		bonalib.Warn("Error marshaling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		bonalib.Warn("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// body, _ := ioutil.ReadAll(resp.Body)
	// bonalib.Info("Response Body:", resp.Status, string(body))
}

func Scheduler() {
	currentDesiredPods := []int32{0, 0, 0}
	newDesiredPods := []int32{0, 0, 0}
	deltaDesiredPods := []int32{0, 0, 0}
	firstTime := true
	var decideInNode decideInNode
	var minResponseTime, minIdx int
	for {
		// get desiredPod from KPA
		response, err := http.Get("http://autoscaler.knative-serving.svc.cluster.local:9999/metrics/kservice/hello")
		if err != nil {
			bonalib.Warn("Error in calling to Kn-Au")
			time.Sleep(time.Duration(SLEEPTIME) * time.Second)
			continue
		}
		if err := json.NewDecoder(response.Body).Decode(&decideInNode); err != nil {
			bonalib.Warn("Failed to decode JSON: ", err)
			continue
		}

		// if firsttime, init prevDesiredPods
		if firstTime {
			currentDesiredPods = []int32{decideInNode.Node1, decideInNode.Node2, decideInNode.Node3}
			firstTime = false
			continue
		}

		newDesiredPods = []int32{decideInNode.Node1, decideInNode.Node2, decideInNode.Node3}
		for i := range currentDesiredPods {
			deltaDesiredPods[i] = newDesiredPods[i] - currentDesiredPods[i]
		}

		// bonalib.Log("currentDesiredPods", currentDesiredPods)
		// bonalib.Log("newDesiredPods", newDesiredPods)
		// bonalib.Log("deltaDesiredPods", deltaDesiredPods)

		// if no change in desiredPods, not call to Kubernetes for updating deployment
		if reflect.DeepEqual(deltaDesiredPods, []int32{0, 0, 0}) {
			time.Sleep(time.Duration(SLEEPTIME) * time.Second)
			continue
		}

		for i_ddp := range deltaDesiredPods {
			for j := deltaDesiredPods[i_ddp]; j != 0; {
				if j < 0 {
					currentDesiredPods[i_ddp]--
					j++
				}
				if j > 0 {
					minResponseTime = 1000000
					minIdx = -1
					for i_rpt := range scraper.RESPONSETIME[i_ddp] { // loop each row of RESPONSETIME
						if currentDesiredPods[i_ddp] >= int32(MAXPON[i_ddp]) {
							continue
						} else {
							if scraper.RESPONSETIME[i_ddp][i_rpt] < minResponseTime {
								minResponseTime = scraper.RESPONSETIME[i_ddp][i_rpt]
								minIdx = i_rpt
							}
						}
					}
					if minIdx != -1 {
						currentDesiredPods[minIdx]++
					}
					j--
				}
			}
		}

		bonalib.Log("currentDesiredPods", currentDesiredPods)

		for i := range currentDesiredPods {
			UpdateDeployment("default", "hello-00001-deployment-"+NODENAMES[i], currentDesiredPods[i])
		}

		response.Body.Close()
		time.Sleep(time.Duration(SLEEPTIME) * time.Second)
	}
}

func UpdateDeployment(namespace string, name string, replicas int32) {
	bonalib.Info("UpdateDeployment", name, replicas)
	clientset, err := kubernetes.NewForConfig(KUBECONFIG)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting deployment: %v\n", err)
		return
	}

	deployment.Spec.Replicas = &replicas

	_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error updating deployment: %v\n", err)
		return
	}
}

func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
