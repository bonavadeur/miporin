package miporin

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Kubeconfig() *rest.Config {
	var config *rest.Config
	var err error
	if os.Getenv("MIPORIN_ENVIRONMENT") == "local" {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
		if err != nil {
			panic(err)
		}
	}
	if os.Getenv("MIPORIN_ENVIRONMENT") == "container" {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}
	return config
}

func GetClientSet() *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(KUBECONFIG)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func GetDynamicClient() *dynamic.DynamicClient {
	dynclient, err := dynamic.NewForConfig(KUBECONFIG)
	if err != nil {
		panic(err.Error())
	}
	return dynclient
}
