package yukari

import (
	"time"
)

type KodomoScheduler struct {
	Name         string
	Decision     map[string]int32
	window       int32
	sleepTime    int8
	Okasan       *OkasanScheduler
	ScheduleStop *StopChan
}

type StopChan struct {
	Kodomo chan bool
	Okasan chan bool
}

func NewKodomoScheduler(
	name string, sleepTime int8,
) *KodomoScheduler {
	atarashiiKodomoScheduler := &KodomoScheduler{
		Name:         name,
		sleepTime:    sleepTime,
		Decision:     map[string]int32{},
		ScheduleStop: NewStopChan(),
	}

	for _, nodename := range NODENAMES {
		atarashiiKodomoScheduler.Decision[nodename] = int32(0)
	}

	go atarashiiKodomoScheduler.schedule()

	return atarashiiKodomoScheduler
}

func NewStopChan() *StopChan {
	newStopChan := &StopChan{
		Kodomo: make(chan bool),
		Okasan: make(chan bool),
	}
	return newStopChan
}

func (s *StopChan) Stop() {
	s.Kodomo <- true
	s.Okasan <- true
}

func (k *KodomoScheduler) schedule() {
	for {
		select {
		case <-k.ScheduleStop.Kodomo:
			return
		default:
			k.Decision = k.Okasan.KPADecision[k.Name]
			time.Sleep(time.Duration(k.sleepTime) * time.Second)
		}
	}
}
