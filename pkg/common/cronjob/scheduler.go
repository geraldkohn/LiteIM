package cronjob

import (
	"sync"
	"time"
)

// Scheduler interfacee
type Scheduler interface {
	Remove(string)
	Add(Jober)
	Start()
	Stop()
}

type schedule struct {
	lock      sync.RWMutex
	startFlag bool

	stopCh   chan struct{}
	removeCh chan string
	addCh    chan Jober

	ticker  *time.Ticker
	jobList map[string]Jober
}

func NewScheduler() Scheduler {
	return &schedule{
		lock:      sync.RWMutex{},
		startFlag: false,
		stopCh:    make(chan struct{}),
		removeCh:  make(chan string, 128),
		addCh:     make(chan Jober, 128),
		ticker:    time.NewTicker(1 * time.Second),
		jobList:   make(map[string]Jober),
	}
}

func (s *schedule) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.startFlag {
		return
	}
	s.startFlag = true
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.run()
			case <-s.stopCh:
				s.stop()
				return
			case name := <-s.removeCh:
				s.remove(name)
			case j := <-s.addCh:
				s.add(j)
			}
		}
	}()
}

func (s *schedule) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.startFlag {
		return
	}
	s.startFlag = false
	s.stopCh <- struct{}{}
}

func (s *schedule) Add(jober Jober) {
	s.addCh <- jober
}

func (s *schedule) add(jober Jober) {
	s.jobList[jober.GetName()] = jober
}

func (s *schedule) Remove(name string) {
	s.removeCh <- name
}

func (s *schedule) remove(name string) {
	delete(s.jobList, name)
}

func (s *schedule) run() {
	for _, j := range s.jobList {
		if j.ShouldRun() {
			go j.Run()
		}
	}
}

func (s *schedule) stop() {
	s.ticker.Stop()
	close(s.stopCh)
	close(s.removeCh)
	close(s.addCh)
}
