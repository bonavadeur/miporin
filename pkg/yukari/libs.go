package yukari

import (
	"context"
	"fmt"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func createSeika(ksvcName string) {
	namespace := "default"

	seikaGVR := schema.GroupVersionResource{
		Group:    "batch.bonavadeur.io",
		Version:  "v1",
		Resource: "seikas",
	}

	var deployment *v1.Deployment
	var err error
	for {
		deployment, err = CLIENTSET.AppsV1().
			Deployments(namespace).
			Get(context.TODO(), ksvcName+"-00001-deployment", metav1.GetOptions{})
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		} else {
			// delete some fields
			deployment.Spec.Template.ObjectMeta.CreationTimestamp = metav1.Time{}
			deployment.ObjectMeta.ResourceVersion = ""
			deployment.ObjectMeta.UID = ""
			// time.Sleep(5 * time.Second)
			break
		}
	}

	seikaInstance := &unstructured.Unstructured{
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
	result, err := DYNCLIENT.Resource(seikaGVR).Namespace("default").Create(context.TODO(), seikaInstance, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		bonalib.Info("Created Seika instance", result.GetName())
	}
}

func deleteSeika(ksvcName string) {
	namespace := "default"

	seikaGVR := schema.GroupVersionResource{
		Group:    "batch.bonavadeur.io",
		Version:  "v1",
		Resource: "seikas",
	}
	deletePolicy := metav1.DeletePropagationBackground
	err := DYNCLIENT.
		Resource(seikaGVR).
		Namespace(namespace).
		Delete(context.TODO(), ksvcName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		bonalib.Warn("Failed to delete Seika instance")
	} else {
		bonalib.Info("Deleted Seika instance", ksvcName)
	}
}
