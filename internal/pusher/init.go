package pusher

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"LiteIM/pkg/common/config"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/logger"
	servicediscovery "LiteIM/pkg/common/service-discovery"
)

var (
	podIP string
	p     *Pusher
)

func initNodeIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Logger.Errorf("Failed to node ip")
		return
	}

	for _, address := range addrs {
		// 过滤掉回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// 返回第一个网卡地址作为标识
				podIP = ipnet.IP.String()
				break
			}
		}
	}

	logger.Logger.Infof("init node ip:%s", podIP)
}

type Pusher struct {
	grpcServer       *grpcServer
	pushRegister     servicediscovery.Register
	gatewayDiscovery servicediscovery.Discovery
	db               *db.DataBases
}

func initPusher() {
	p.grpcServer = newGrpcServer(config.Conf.Pusher.GrpcPort)
	p.pushRegister = servicediscovery.NewEtcdRegister(
		config.Conf.Pusher.PusherServiceName,
		fmt.Sprintf("%s:%d", podIP, config.Conf.Pusher.GrpcPort),
	)
	p.gatewayDiscovery = servicediscovery.NewEtcdDiscovery(
		config.Conf.Pusher.GatewayServiceName,
	)
	p.db = db.NewDataBases(
		db.MysqlConfig{
			Addr:     config.Conf.Mysql.MysqlAddr,
			Username: config.Conf.Mysql.MysqlUsername,
			Password: config.Conf.Mysql.MysqlPassword,
		},
		db.RedisConfig{
			Addr:     config.Conf.Redis.RedisAddr,
			Password: config.Conf.Redis.RedisPassword,
		},
		db.MongodbConfig{
			Addr:     config.Conf.Mongo.MongoAddr,
			Username: config.Conf.Mongo.MongoUsername,
			Password: config.Conf.Mongo.MongoPassword,
		},
	)
}

func Run() {
	initNodeIP()
	initPusher()

	go p.pushRegister.Run()
	go p.gatewayDiscovery.Watch()
	go p.grpcServer.Run()

	// 实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	p.pushRegister.Exit()
	p.gatewayDiscovery.Exit()
	p.grpcServer.Exit()
}
