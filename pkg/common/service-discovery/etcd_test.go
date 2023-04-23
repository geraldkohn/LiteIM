package servicediscovery

import (
	"sync"
	"testing"

	"github.com/geraldkohn/im/pkg/common/setting"
)

func TestRegister(t *testing.T) {
	setting.APPSetting.Etcd = &setting.Etcd{
		Addr:     []string{},
		Username: "",
		Password: "",
	}
	register := NewEtcdRegister("service-test", "127.0.0.1:8080")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		register.Run()
		wg.Done()
	}()
	wg.Wait()
}
