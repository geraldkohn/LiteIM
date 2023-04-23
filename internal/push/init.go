package push

import (
	servicediscovery "github.com/geraldkohn/im/pkg/common/service-discovery"
	"github.com/geraldkohn/im/pkg/common/setting"
)

var (
	pushRegister     servicediscovery.Register
	gatewayDiscovery servicediscovery.Discovery
)

func Run() {
	initServiceDiscovery()
	initGrpcServer()
}

func initServiceDiscovery() {
	pushRegister = servicediscovery.NewEtcdRegister(setting.APPSetting.ServiceDiscovery.PushServiceName, setting.APPSetting.ServiceDiscovery.PushServiceEndpoint)
	go pushRegister.Run()

	gatewayDiscovery = servicediscovery.NewEtcdDiscovery(setting.APPSetting.ServiceDiscovery.GatewayServiceName)
	go gatewayDiscovery.Watch()
}

func initGrpcServer() {
	pushServer := NewGrpcPushServer(int(setting.APPSetting.RPC.PushRPCPort))
	go pushServer.Run()
}
