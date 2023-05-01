package cronjob

import "time"

type Jober interface {
	GetName() string
	ShouldRun() bool
	Run()
}

type DoFunc func()

type job struct {
	name     string
	doFunc   DoFunc
	execTime time.Time
	duration time.Duration
}

func NewJober(name string, do DoFunc, duration, delay time.Duration) Jober {
	return &job{
		name:     name,
		doFunc:   do,
		execTime: time.Now().Add(delay),
		duration: duration,
	}
}

func (j *job) GetName() string {
	return j.name
}

func (j *job) ShouldRun() bool {
	return time.Now().Unix() >= j.execTime.Unix()
}

func (j *job) Run() {
	j.doFunc()
	j.setExecTime()
}

func (j *job) setExecTime() {
	j.execTime = j.execTime.Add(j.duration)
}
