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
	conf  *config.Config
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

func initConfig() {
	conf = config.Init()
}

type Pusher struct {
	grpcServer       *grpcServer
	pushRegister     servicediscovery.Register
	gatewayDiscovery servicediscovery.Discovery
	db               *db.DataBases
}

func initPusher() {
	p.grpcServer = newGrpcServer(conf.Pusher.GrpcPort)
	p.pushRegister = servicediscovery.NewEtcdRegister(
		conf.Pusher.PusherServiceName,
		fmt.Sprintf("%s:%d", podIP, conf.Pusher.GrpcPort),
	)
	p.gatewayDiscovery = servicediscovery.NewEtcdDiscovery(
		conf.Pusher.GatewayServiceName,
	)
	p.db = db.NewDataBases(
		db.MysqlConfig{
			Addr:     conf.Mysql.MysqlAddr,
			Username: conf.Mysql.MysqlUsername,
			Password: conf.Mysql.MysqlPassword,
		},
		db.RedisConfig{
			Addr:     conf.Redis.RedisAddr,
			Username: conf.Redis.RedisUsername,
			Password: conf.Redis.RedisPassword,
		},
		db.MongodbConfig{
			Addr:     conf.Mongo.MongoAddr,
			Username: conf.Mongo.MongoUsername,
			Password: conf.Mongo.MongoPassword,
		},
	)
}

func Run() {
	initConfig()
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
