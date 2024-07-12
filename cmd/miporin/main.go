package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/miporin"
	"github.com/bonavadeur/miporin/pkg/scraper"
	"github.com/labstack/echo/v4"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var _ = bonalib.Baka()

var (
	PROMSERVER     = "http://prometheus-kube-prometheus-prometheus:9090/api/v1/query?query="
	NODENAMES      = []string{"node1", "node2", "node3"}
	KUBECONFIG     = miporin.Kubeconfig()
	SLEEPTIME      = 2
	METRIC_SCRAPED = make([][]int, 3)
	MAXPON         = []int{10, 10, 3}
)

func WatchEventCreateKsvc() {
	dynamicClient, _ := dynamic.NewForConfig(KUBECONFIG)
	clientset, _ := kubernetes.NewForConfig(KUBECONFIG)

	namespace := "default"

	ksvcGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}

	watcher, err := dynamicClient.Resource(ksvcGVR).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}

	seikaGVR := schema.GroupVersionResource{
		Group:    "batch.bonavadeur.io",
		Version:  "v1",
		Resource: "seikas",
	}
	seikaInstance := &unstructured.Unstructured{}

	bonalib.Use(clientset, seikaGVR, seikaInstance)

	for event := range watcher.ResultChan() {
		if event.Type == watch.Added {
			ksvc, _ := event.Object.(*unstructured.Unstructured)
			ksvcName, _, _ := unstructured.NestedString(ksvc.Object, "metadata", "name")
			bonalib.Warn("Ksvc has been created:", ksvcName)

			// get deployment named hello-00001-deployment
			var deployment *v1.Deployment
			for {
				deployment, err = clientset.AppsV1().Deployments(namespace).Get(context.TODO(), ksvcName+"-00001-deployment", metav1.GetOptions{})
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				} else {
					bonalib.Info("Found deployment", deployment.GetName())
					// delete some fields
					deployment.Spec.Template.ObjectMeta.CreationTimestamp = metav1.Time{}
					deployment.ObjectMeta.ResourceVersion = ""
					deployment.ObjectMeta.UID = ""
					time.Sleep(5 * time.Second)
					break
				}
			}

			// // prepare for seika instance
			seikaInstance = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch.bonavadeur.io/v1",
					"kind":       "Seika",
					"metadata": map[string]interface{}{
						"name": ksvcName,
					},
					"spec": map[string]interface{}{
						"repurika": map[string]interface{}{},
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"bonavadeur.io/seika": ksvcName,
							},
						},
						"template": deployment.Spec.Template,
					},
				},
			}
			repurika := seikaInstance.Object["spec"].(map[string]interface{})["repurika"].(map[string]interface{})
			for _, nodename := range NODENAMES {
				repurika[nodename] = 0
			}

			// // create seika instance
			result, err := dynamicClient.Resource(seikaGVR).Namespace("default").Create(context.TODO(), seikaInstance, metav1.CreateOptions{})
			if err != nil {
				fmt.Println(err)
			} else {
				bonalib.Info("Created Seika instance", result.GetName())
			}
		}
		if event.Type == watch.Deleted {
			bonalib.Warn("Ksvc has been deleted")
			deletePolicy := metav1.DeletePropagationBackground
			err := dynamicClient.Resource(seikaGVR).Namespace("default").Delete(context.TODO(), seikaInstance.GetName(), metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				bonalib.Warn("Failed to delete Seika instance")
			} else {
				bonalib.Info("Deleted Seika instance", seikaInstance.GetName())
			}
		}
	}
}

func SchedulerSeika() {
	currentDesiredPods := make([]int32, len(NODENAMES))
	newDesiredPods := make([]int32, len(NODENAMES))
	deltaDesiredPods := make([]int32, len(NODENAMES))
	firstTime := true
	var decideInNode decideInNode
	var minResponseTime, minIdx int

	for {
		// get desiredPod from KPA
		response, err := http.Get("http://autoscaler.knative-serving.svc.cluster.local:9999/metrics/kservice/hello")
		if err != nil {
			bonalib.Warn("Error in calling to Kn-Au")
			time.Sleep(5 * time.Second)
			continue
		}
		if err := json.NewDecoder(response.Body).Decode(&decideInNode); err != nil {
			bonalib.Warn("Failed to decode JSON: ", err)
			continue
		}

		// if firsttime, init prevDesiredPods
		if firstTime {
			currentDesiredPods = []int32{decideInNode.Node1, decideInNode.Node2, decideInNode.Node3}
			for i := range currentDesiredPods {
				deltaDesiredPods[i] = newDesiredPods[i] - currentDesiredPods[i]
			}
			firstTime = false
			continue
		}

		newDesiredPods = []int32{decideInNode.Node1, decideInNode.Node2, decideInNode.Node3}
		for i := range currentDesiredPods {
			deltaDesiredPods[i] = newDesiredPods[i] - currentDesiredPods[i]
		}

		// if no change in desiredPods, not call to Kubernetes for updating deployment
		if reflect.DeepEqual(deltaDesiredPods, []int32{0, 0, 0}) {
			time.Sleep(time.Duration(SLEEPTIME) * time.Second)
			continue
		}

		// if have change in desiredPods, update deployment
		bonalib.Info("RESPONSETIME", scraper.RESPONSETIME)
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

		Patch(int(currentDesiredPods[0]), int(currentDesiredPods[1]), int(currentDesiredPods[2]))
		// for i := range currentDesiredPods {
		// 	UpdateDeployment("default", "hello-00001-deployment-"+NODES[i], currentDesiredPods[i])
		// }

		response.Body.Close()
		time.Sleep(time.Duration(SLEEPTIME) * time.Second)
	}
}

type decideInNode struct {
	Node1 int32 `json:"node1"`
	Node2 int32 `json:"node2"`
	Node3 int32 `json:"node3"`
}

func Patch(node1 int, node2 int, node3 int) {
	dynamicClient, err := dynamic.NewForConfig(KUBECONFIG)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "batch.bonavadeur.io",
		Version:  "v1",
		Resource: "seikas",
	}

	// Define the patch data
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"repurika": map[string]interface{}{
				"node1": node1,
				"node2": node2,
				"node3": node3,
			},
		},
	}

	// Convert patch data to JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		fmt.Printf("Error marshalling patch data: %v", err)
	}

	// Namespace and resource name
	namespace := "default"
	resourceName := "hello"

	// Execute the patch request
	patchedResource, err := dynamicClient.Resource(gvr).
		Namespace(namespace).
		Patch(context.TODO(), resourceName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("Error patching resource: %v", err)
	}

	resource, _, _ := unstructured.NestedString(patchedResource.Object, "metadata", "name")
	bonalib.Info("Patched resource: ", resource)
}

func main() {
	bonalib.Log("Konnichiwa, Miporin-chan desu")

	// go scrapeMetrics()
	go scraper.ScrapeMetrics()

	// go Scheduler()

	go SchedulerSeika()

	go WatchEventCreateKsvc()

	// Start Echo Server
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Konnichiwa, Miporin-chan desu\n")
	})

	e.GET("/api/weighted", func(c echo.Context) error {
		return c.JSON(http.StatusOK, [][]int(scraper.WEIGHT))
	})

	e.Logger.Fatal(e.Start(":18080"))
}
