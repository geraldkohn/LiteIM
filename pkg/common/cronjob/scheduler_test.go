package cronjob

import (
	"fmt"
	"testing"
	"time"
)

func TestSchedulerRun(t *testing.T) {
	testCases := []struct {
		name     string
		dofunc   func()
		duration time.Duration
	}{
		{name: "job1", dofunc: func() { fmt.Printf("job1: %v\n", time.Now()) }, duration: time.Duration(10 * time.Second)},
		{name: "job2", dofunc: func() { fmt.Printf("job2: %v\n", time.Now()) }, duration: time.Duration(5 * time.Second)},
		{name: "job3", dofunc: func() { fmt.Printf("job3: %v\n", time.Now()) }, duration: time.Duration(3 * time.Second)},
	}

	scheduler := NewScheduler()

	for _, c := range testCases {
		j := NewJober(c.name, c.dofunc, c.duration, 0)
		scheduler.Add(j)
	}

	scheduler.Start()
	<-time.After(20 * time.Second)
	j4 := NewJober("job-add1", func() { fmt.Printf("job-add1 %v\n", time.Now()) }, time.Duration(2*time.Second), 0)
	scheduler.Add(j4)
	<-time.After(20 * time.Second)
	scheduler.Remove(j4.GetName())
	<-time.After(10 * time.Second)
	scheduler.Stop()
}
