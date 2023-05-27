package servicediscovery

import (
	"sync"
	"testing"
)

func TestRegister(t *testing.T) {
	register := NewEtcdRegister("service-test", "127.0.0.1:8080")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		register.Run()
		wg.Done()
	}()
	wg.Wait()
}
