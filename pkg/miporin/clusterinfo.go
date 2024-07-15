package miporin

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

var (
	KUBECONFIG *rest.Config
	CLIENTSET  *kubernetes.Clientset
	NODENAMES  []string
	PODCIDRS   []PodCIDR
)

func init() {
	KUBECONFIG = Kubeconfig()
	CLIENTSET = GetClientSet()
	NODENAMES = GetNodenames()
	PODCIDRS = GetPodsCIDRs()
}

type PodCIDR struct {
	Nodename   string
	NodeIP     string
	PodIPRange string
	PodPrefix  int32
}

type IPAMBlock struct {
	Spec struct {
		CIDR     string  `json:"cidr"`
		Affinity *string `json:"affinity"`
	} `json:"spec"`
}

type IPAMBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAMBlock `json:"items"`
}

func (obj *IPAMBlockList) DeepCopyObject() runtime.Object {
	dst := &IPAMBlockList{}
	dst.TypeMeta = obj.TypeMeta
	dst.ListMeta = obj.ListMeta
	dst.Items = make([]IPAMBlock, len(obj.Items))
	objcopy := make([]IPAMBlock, len(obj.Items))
	copy(objcopy, obj.Items)
	for i := range objcopy {
		dst.Items[i] = obj.Items[i]
	}
	return dst
}

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

func GetNodenames() []string {
	nodes, err := CLIENTSET.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		bonalib.Warn("Error listing nodes: %v\n", err)
		return []string{}
	}

	ret := []string{}
	for _, node := range nodes.Items {
		ret = append(ret, node.Name)
	}
	sort.Strings(ret)
	return ret
}

func GetDynamicClient() *dynamic.DynamicClient {
	dynclient, err := dynamic.NewForConfig(KUBECONFIG)
	if err != nil {
		panic(err.Error())
	}
	return dynclient
}

func GetPodsCIDRs() []PodCIDR {
	nodes, err := CLIENTSET.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	var ipamBlockList IPAMBlockList
	retry.OnError(retry.DefaultRetry, func(error) bool { return true }, func() error {
		return CLIENTSET.RESTClient().
			Get().
			AbsPath("/apis/crd.projectcalico.org/v1/ipamblocks").
			Do(context.TODO()).
			Into(&ipamBlockList)
	})

	ret := make([]PodCIDR, 0, len(NODENAMES))

	for i := 0; i < len(NODENAMES); i++ {
		nodeip := []corev1.NodeAddress{}
		node := NODENAMES[i]
		for j := 0; j < len(nodes.Items); j++ {
			if nodes.Items[j].Name == node {
				nodeip = nodes.Items[j].Status.Addresses
			}
		}
		for _, item := range ipamBlockList.Items {
			if item.Spec.Affinity != nil && strings.Contains(*item.Spec.Affinity, node) {
				ip, ipNet, err := net.ParseCIDR(item.Spec.CIDR)
				ones, _ := ipNet.Mask.Size()
				if err != nil {
					panic(err)
				}
				ret = append(ret, PodCIDR{
					Nodename:   node,
					NodeIP:     nodeip[0].Address,
					PodIPRange: ip.String(),
					PodPrefix:  int32(ones),
				})
			}
		}
	}

	return ret
}

func CheckIPInNode(ip string) string {
	for _, podcidr := range PODCIDRS {
		_, ipNet, _ := net.ParseCIDR(podcidr.PodIPRange + "/" + strconv.Itoa(int(podcidr.PodPrefix)))
		if ipNet.Contains(net.ParseIP(ip)) {
			return podcidr.Nodename
		}
	}
	return ""
}
