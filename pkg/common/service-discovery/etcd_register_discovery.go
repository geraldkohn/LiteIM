package servicediscovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"Lite_IM/pkg/common/setting"

	"github.com/golang/glog"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd "go.etcd.io/etcd/client/v3"
)

var (
	keyServiceName = "%s-%s" // serviceName-endpoint
)

type EtcdDiscovery struct {
	cli       *etcd.Client
	service   string
	lock      sync.RWMutex
	endpoints []string
	close     chan error
}

func NewEtcdDiscovery(serviceName string) Discovery {
	cli, err := etcd.New(etcd.Config{
		Endpoints: setting.APPSetting.Etcd.Addr,
		Username:  setting.APPSetting.Etcd.Username,
		Password:  setting.APPSetting.Etcd.Password,
	})
	if err != nil {
		glog.Errorf("failed to create a etcd discovery, error: %s", err)
		return nil
	}
	d := &EtcdDiscovery{
		cli:       cli,
		service:   serviceName,
		lock:      sync.RWMutex{},
		endpoints: make([]string, 0),
		close:     make(chan error),
	}
	return d
}

func (d *EtcdDiscovery) addOne(endpoint string) {
	d.lock.Lock()
	d.endpoints = append(d.endpoints, endpoint)
	d.lock.Unlock()
}

func (d *EtcdDiscovery) deleteOne(endpoint string) {
	d.lock.Lock()
	var index int
	for k, v := range d.endpoints {
		if endpoint == v {
			index = k
			break
		}
	}
	for i := index + 1; i < len(d.endpoints); i++ {
		d.endpoints[i-1] = d.endpoints[i]
	}
	d.lock.Unlock()
}

func (d *EtcdDiscovery) PickOne() string {
	endpoints := d.pick()
	return endpoints[rand.Intn(len(endpoints))]
}

func (d *EtcdDiscovery) PickAll() []string {
	return d.pick()
}

func (d *EtcdDiscovery) pick() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.endpoints
}

func (d *EtcdDiscovery) Watch() {
	watcher := etcd.NewWatcher(d.cli)
	watcherChan := watcher.Watch(context.Background(), d.service, etcd.WithPrefix())
	for {
		select {
		case resp := <-watcherChan:
			for _, ev := range resp.Events {
				switch ev.Type {
				case etcd.EventTypePut:
					k := string(ev.Kv.Key)
					endpoint := strings.Split(k, "-")
					d.addOne(endpoint[1])
				case etcd.EventTypeDelete:
					k := string(ev.Kv.Key)
					endpoint := strings.Split(k, "-")
					d.deleteOne(endpoint[1])
				}
			}
		case <-d.close:
			return
		}
	}
}

func (d *EtcdDiscovery) Exit() {
	close(d.close)
	d.cli.Close()
}

type EtcdRegister struct {
	cli      *etcd.Client
	name     string
	endpoint string
	close    chan error
}

func NewEtcdRegister(name string, endpoint string) Register {
	cli, err := etcd.New(etcd.Config{
		Endpoints: setting.APPSetting.Etcd.Addr,
		Username:  setting.APPSetting.Etcd.Username,
		Password:  setting.APPSetting.Etcd.Password,
	})
	if err != nil {
		glog.Errorf("failed to create a etcd discovery, error: %s", err)
		return nil
	}
	r := &EtcdRegister{
		cli:      cli,
		name:     name,
		endpoint: endpoint,
	}
	return r
}

func (r *EtcdRegister) Run() {
	ctx := context.Background()
	ticker := time.NewTicker(4 * time.Second)
	leaseID, err := etcdPut(ctx, r.cli, fmt.Sprintf(keyServiceName, r.name, r.endpoint), r.endpoint, 10)
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ticker.C:
			_, err := r.cli.KeepAlive(ctx, etcd.LeaseID(leaseID))
			if err != nil {
				if err == rpctypes.ErrLeaseNotFound {
					leaseID, _ = etcdPut(ctx, r.cli, fmt.Sprintf(keyServiceName, r.name, r.endpoint), r.endpoint, 10)
					err = nil
				}
				if err != nil {
					glog.Error(errors.New("Register KeepAlive error: " + err.Error()))
					return
				}
			}
		case <-r.close:
			r.cli.Revoke(ctx, etcd.LeaseID(leaseID))
		}
	}
}

func (r *EtcdRegister) Exit() {
	close(r.close)
}

func etcdPut(ctx context.Context, cli *etcd.Client, key string, val string, second int64) (int64, error) {
	leaseResp, err := cli.Grant(ctx, second)
	if err != nil {
		return -1, err
	}
	_, err = cli.Put(ctx, key, val, etcd.WithLease(leaseResp.ID))
	if err != nil {
		return -1, err
	}
	return int64(leaseResp.ID), nil
}
