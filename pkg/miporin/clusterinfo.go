package miporin

import (
	"context"
	"net"
	"sort"
	"strconv"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	KUBECONFIG = Kubeconfig()
	CLIENTSET  = GetClientSet()
	NODENAMES  = GetNodenames()
	PODCIDRS   = GetPodsCIDRs()
)

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

func CheckIPInNode(ip string) string {
	for _, podcidr := range PODCIDRS {
		_, ipNet, _ := net.ParseCIDR(podcidr.PodIPRange + "/" + strconv.Itoa(int(podcidr.PodPrefix)))
		if ipNet.Contains(net.ParseIP(ip)) {
			return podcidr.Nodename
		}
	}
	return ""
}
