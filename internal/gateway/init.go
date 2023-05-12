package gateway

import (
	"net"
	"os"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/logger"

	"github.com/spf13/viper"
)

var (
	hServer    *HServer
	wServer    *WServer
	pushServer *GServer
	nodeIP     string
)

func initConfig() {
	viper.SetConfigType("json")
	f, err := os.Open("conf/conf.json")
	if err != nil {
		panic("无法打开配置文件 | " + err.Error())
	}
	viper.ReadConfig(f)
}

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
				nodeIP = ipnet.IP.String()
				break
			}
		}
	}

	logger.Infof("init node ip:%s", nodeIP)
}

func initDataBases() {
	database.Databases = db.NewDataBases(
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
}

// 阻塞方法
func initHttpServer() {
	hServer.onInit()
	hServer.Run()
}

// 阻塞方法
func initWebsocketServer() {
	wServer.onInit()
	wServer.Run()
}

// 阻塞方法
func initGrpcServer() {
	pushServer = NewGrpcPushServer(viper.GetInt("GRPCPort"))
	pushServer.Run()
}

func Run() {
	initConfig()
	initNodeIP()
	initDataBases()
	go initHttpServer()
	go initWebsocketServer()
	go initGrpcServer()
}
