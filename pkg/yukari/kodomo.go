package yukari

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
)

type KodomoScheduler struct {
	Name         string
	Decision     map[string]int32
	window       int32
	sleepTime    int8
	Okasan       *OkasanScheduler
	ScheduleStop chan bool
}

func NewKodomoScheduler(
	name string, sleepTime int8,
) *KodomoScheduler {
	atarashiiKodomoScheduler := &KodomoScheduler{
		Name:         name,
		sleepTime:    sleepTime,
		Decision:     map[string]int32{},
		ScheduleStop: make(chan bool),
	}

	for _, nodename := range NODENAMES {
		atarashiiKodomoScheduler.Decision[nodename] = int32(0)
	}

	go atarashiiKodomoScheduler.schedule()

	return atarashiiKodomoScheduler
}

func (k *KodomoScheduler) schedule() {
	decideInNode := map[string]int32{}
	for {
		select {
		case <-k.ScheduleStop:
			return
		default:
			// get desiredPod from KPA
			response, err := http.Get("http://autoscaler.knative-serving.svc.cluster.local:9999/metrics/kservices/" + k.Name)
			if err != nil {
				bonalib.Warn("Error in calling to Kn-Au")
				time.Sleep(5 * time.Second)
				continue
			}
			if err := json.NewDecoder(response.Body).Decode(&decideInNode); err != nil {
				bonalib.Warn("Failed to decode JSON: ", err)
				continue
			}
			response.Body.Close()

			k.Decision = decideInNode

			time.Sleep(time.Duration(k.sleepTime) * time.Second)
		}
	}
}
