package gateway

import (
	"net"

	"github.com/geraldkohn/im/pkg/common/logger"
)

var (
	hServer *HServer
	wServer *WServer
	IP      string
)

func initNodeIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorf("Failed to node ip")
		return
	}

	for _, address := range addrs {
		// 过滤掉回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// 返回第一个网卡地址作为标识
				IP = ipnet.IP.String()
				break
			}
		}
	}

	logger.Infof("init node ip:%s", IP)
}

func initCronTask() {

}

func initAll() {
	initNodeIP()
	initCronTask()
	hServer.onInit()
	wServer.onInit()
}

func run() {
	go wServer.run()
	go hServer.run()
}

func Run() {
	initAll()
	run()
}
