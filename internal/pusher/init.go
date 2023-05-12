package pusher

import (
	"fmt"
	"net"
	"os"

	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/logger"
	servicediscovery "LiteIM/pkg/common/service-discovery"

	"github.com/spf13/viper"
)

var (
	podIP string
	p     *Pusher
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
				podIP = ipnet.IP.String()
				break
			}
		}
	}

	logger.Infof("init node ip:%s", podIP)
}

func initConfig() {
	viper.SetConfigType("json")
	f, err := os.Open("conf/conf.json")
	if err != nil {
		panic("无法打开文件" + err.Error())
	}
	viper.ReadConfig(f)
}

type Pusher struct {
	grpcServer       *grpcServer
	pushRegister     servicediscovery.Register
	gatewayDiscovery servicediscovery.Discovery
	db               *db.DataBases
	exit             chan error
}

func (p *Pusher) initialize() {
	p.grpcServer = newGrpcServer(viper.GetInt("GRPCPort"))
	p.pushRegister = servicediscovery.NewEtcdRegister(
		viper.GetString("PushServiceName"),
		fmt.Sprintf("%s:%d", podIP, viper.GetInt("GRPCPort")),
	)
	p.gatewayDiscovery = servicediscovery.NewEtcdDiscovery(
		viper.GetString("GatewayServiceName"),
	)
	p.db = db.NewDataBases(
		db.MysqlConfig{
			Addr:     viper.GetString("MysqlAddr"),
			Username: viper.GetString("MysqlUsername"),
			Password: viper.GetString("MysqlPassword"),
		},
		db.RedisConfig{
			Addr:     viper.GetString("RedisAddr"),
			Username: viper.GetString("RedisUsername"),
			Password: viper.GetString("RedisPassword"),
		},
		db.MongodbConfig{
			Addr:     viper.GetStringSlice("MongoAddr"),
			Username: viper.GetString("MongoUsername"),
			Password: viper.GetString("MongoPassword"),
		},
	)
	p.exit = make(chan error)
}

func Run() {
	initConfig()
	initNodeIP()
	p.initialize()
	go p.pushRegister.Run()
	go p.gatewayDiscovery.Watch()
	go p.grpcServer.Run()
	<-p.exit
	p.pushRegister.Exit()
	p.gatewayDiscovery.Exit()
	p.grpcServer.Exit()
}

func Exit() {
	close(p.exit)
}
