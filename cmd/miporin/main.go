package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/libs"
	"github.com/bonavadeur/miporin/pkg/miporin"
	"github.com/bonavadeur/miporin/pkg/scraper"
	"github.com/bonavadeur/miporin/pkg/yukari"
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

var (
	PROMSERVER        = "http://prometheus-kube-prometheus-prometheus:9090/api/v1/query?query="
	NODENAMES         = []string{"node1", "node2", "node3"}
	KUBECONFIG        = miporin.Kubeconfig()
	SLEEPTIME         = 2
	METRIC_SCRAPED    = make([][]int, 3)
	MAXPON            = []int{10, 10, 3}
	OKASAN_SCRAPERS   = map[string]*scraper.OkasanScraper{}
	OKASAN_SCHEDULERS = map[string]*yukari.OkasanScheduler{}
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
	currentDesiredPods := map[string]int32{}
	newDesiredPods := map[string]int32{}
	deltaDesiredPods := map[string]int32{}
	decideInNode := map[string]int32{}
	for _, nodename := range NODENAMES {
		currentDesiredPods[nodename] = 0
		newDesiredPods[nodename] = 0
		deltaDesiredPods[nodename] = 0
	}
	firstTime := true
	var minResponseTime, minIdx int

	bonalib.Use(currentDesiredPods, newDesiredPods, deltaDesiredPods, decideInNode, minResponseTime, minIdx, firstTime)

	for {
		// get desiredPod from KPA
		response, err := http.Get("http://autoscaler.knative-serving.svc.cluster.local:9999/metrics/kservices/hello")
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
			currentDesiredPods = decideInNode
			firstTime = false
		} else {
			newDesiredPods = decideInNode
		}
		for k_cdp := range currentDesiredPods {
			deltaDesiredPods[k_cdp] = newDesiredPods[k_cdp] - currentDesiredPods[k_cdp]
		}

		// check if no change in desiredPods, not call to Kubernetes for updating deployment
		for k_dpp, v_dpp := range deltaDesiredPods {
			if v_dpp != 0 { // if have any change in delta, break and go to following steps
				break
			}
			if k_dpp == NODENAMES[len(NODENAMES)-1] { // if no change, sleep and continue
				time.Sleep(time.Duration(SLEEPTIME) * time.Second)
				continue
			}
		}

		nodeidx := map[string]int{}
		for i, nodename := range NODENAMES {
			nodeidx[nodename] = i
		}

		for k_ddp := range deltaDesiredPods {
			for i := deltaDesiredPods[k_ddp]; i != 0; {
				if i < 0 {
					currentDesiredPods[k_ddp]--
					i++
				}
				if i > 0 {
					minResponseTime = 1000000
					minIdx = -1
					for i_rpt := range scraper.RESPONSETIME[nodeidx[k_ddp]] { // loop each row of RESPONSETIME
						if currentDesiredPods[k_ddp] >= int32(MAXPON[nodeidx[k_ddp]]) {
							continue
						} else {
							if scraper.RESPONSETIME[nodeidx[k_ddp]][i_rpt] < minResponseTime {
								minResponseTime = scraper.RESPONSETIME[nodeidx[k_ddp]][i_rpt]
								minIdx = i_rpt
							}
						}
					}
					if minIdx != -1 {
						currentDesiredPods[NODENAMES[minIdx]]++
					}
					i--
				}
			}
		}

		bonalib.Log("currentDesiredPods", currentDesiredPods)

		Patch(currentDesiredPods)

		response.Body.Close()
		time.Sleep(time.Duration(SLEEPTIME) * time.Second)
	}
}

func Patch(desiredPods map[string]int32) {
	dynamicClient, err := dynamic.NewForConfig(KUBECONFIG)

	gvr := schema.GroupVersionResource{
		Group:    "batch.bonavadeur.io",
		Version:  "v1",
		Resource: "seikas",
	}

	// Define the patch data
	repurika := map[string]interface{}{}
	for _, nodename := range NODENAMES {
		repurika[nodename] = desiredPods[nodename]
	}
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"repurika": repurika,
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
		bonalib.Warn("Error patching resource: ", err)
	} else {
		resource, found, _ := unstructured.NestedString(patchedResource.Object, "metadata", "name")
		if !found {
			bonalib.Warn("Seika not found:", err)
		}
		bonalib.Info("Patched resource:", resource)
	}
}

func init() {
	scraper.OKASAN_SCRAPERS = OKASAN_SCRAPERS
	yukari.OKASAN_SCRAPERS = OKASAN_SCRAPERS
	yukari.OKASAN_SCHEDULERS = OKASAN_SCHEDULERS
}

func main() {
	bonalib.Log("Konnichiwa, Miporin-chan desu")

	go libs.License()

	go scraper.Scraper(OKASAN_SCRAPERS)

	if miporin.Cm2Bool("ikukantai-miporin-enable-yukari") {
		// go WatchEventCreateKsvc()
		// go SchedulerSeika()
		go yukari.Scheduler(OKASAN_SCHEDULERS)
	}

	// Start Echo Server
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Konnichiwa, Miporin-chan desu\n")
	})

	e.GET("/api/weight/okasan/:okasan/kodomo/:kodomo", func(c echo.Context) error {
		okasanScraper := c.Param("okasan")
		kodomoScraper := c.Param("kodomo")
		return c.JSON(http.StatusOK, OKASAN_SCRAPERS[okasanScraper].Kodomo[kodomoScraper].Weight)
	})

	e.Logger.Fatal(e.Start(":18080"))
}
