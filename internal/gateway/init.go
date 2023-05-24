package gateway

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/config"
	"LiteIM/pkg/common/cronjob"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/kafka"
	"LiteIM/pkg/common/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	hServer    *HServer
	wServer    *WServer
	pushServer *GServer
	nodeIP     string
	conf       *config.Config
)

func initConfig() {
	conf = config.Init()
}

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
				nodeIP = ipnet.IP.String()
				break
			}
		}
	}

	logger.Logger.Infof("init node ip:%s", nodeIP)
}

func initDataBases() {
	database.Databases = db.NewDataBases(
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

// 阻塞方法
func initHttpServer() {
	router := gin.Default()
	initRouter(router)
	hServer = new(HServer)
	hServer.server = &http.Server{
		Addr:         ":" + strconv.Itoa(conf.Gateway.HttpPort),
		Handler:      router,
		ReadTimeout:  time.Duration(conf.Gateway.HttPReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.Gateway.HttPWriteTimeout) * time.Second,
	}
}

// 阻塞方法
func initWebsocketServer() {
	wServer = new(WServer)
	wServer.port = conf.Gateway.WebsocketPort
	wServer.wsUpgrader = &websocket.Upgrader{
		HandshakeTimeout: time.Duration(conf.Gateway.WebsocketTimeout) * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	wServer.connMap = make(map[string]*UserConn)
	wServer.connMapLock = new(sync.RWMutex)
	wServer.producer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   conf.Kafka.KafkaBrokerAddr,
		SASLUsername: conf.Kafka.KafkaSASLUsername,
		SASLPassword: conf.Kafka.KafkaSASLPassword,
		Topic:        conf.Kafka.KafkaMsgTopic,
	})
	wServer.scheduler = cronjob.NewScheduler()
	wServer.exit = make(chan error)
}

// 阻塞方法
func initGrpcServer() {
	pushServer = new(GServer)
	pushServer.port = conf.Gateway.GrpcPort
}

func Run() {
	initConfig()
	initNodeIP()
	initDataBases()
	initHttpServer()
	initWebsocketServer()
	initGrpcServer()

	go hServer.Run()
	go wServer.Run()
	go pushServer.Run()
}
