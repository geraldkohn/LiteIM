package config

import (
	"LiteIM/pkg/common/logger"
	"os"

	"github.com/spf13/viper"
)

var Conf Config

func init() {
	viper.SetConfigType("toml")
	f, err := os.Open("config/config.toml")
	if err != nil {
		panic("无法打开配置文件 | " + err.Error())
	}
	viper.ReadConfig(f)
	Conf = Config{}
	autoFillConfig(&Conf)
	logger.Logger.Debug(Conf)
}

type Config struct {
	Gateway  Gateway
	Pusher   Pusher
	Transfer Transfer
	Kafka    Kafka
	Mysql    Mysql
	Redis    Redis
	Mongo    Mongo
	Etcd     Etcd
	Secret   Secret
}

type Gateway struct {
	GrpcPort         int
	HttpPort         int
	HttPReadTimeout  int
	HttPWriteTimeout int
	WebsocketPort    int
	WebsocketTimeout int
}

type Pusher struct {
	GrpcPort           int
	PusherServiceName  string
	GatewayServiceName string
}

type Transfer struct {
	PusherServiceName string
}

type Kafka struct {
	KafkaBrokerAddr     []string
	KafkaSASLUsername   string
	KafkaSASLPassword   string
	KafkaMsgTopic       string
	KafkaRetryPushTopic string
}

type Mysql struct {
	MysqlAddr     string
	MysqlUsername string
	MysqlPassword string
}

type Redis struct {
	RedisAddr     string
	RedisPassword string
}

type Mongo struct {
	MongoAddr     string
	MongoUsername string
	MongoPassword string
}

type Etcd struct {
	EtcdAddr     []string
	EtcdUsername string
	EtcdPassword string
}

type Secret struct {
	JwtSecret string
}

func autoFillConfig(config *Config) {
	config.Gateway = Gateway{
		GrpcPort:         viper.GetInt("gateway.grpcPort"),
		HttpPort:         viper.GetInt("gateway.httpPort"),
		HttPReadTimeout:  viper.GetInt("gateway.httPReadTimeout"),
		HttPWriteTimeout: viper.GetInt("gateway.httPWriteTimeout"),
		WebsocketPort:    viper.GetInt("gateway.websocketPort"),
		WebsocketTimeout: viper.GetInt("gateway.websocketTimeout"),
	}
	config.Transfer = Transfer{
		PusherServiceName: viper.GetString("transfer.pusherServiceName"),
	}
	config.Pusher = Pusher{
		GrpcPort:           viper.GetInt("pusher.grpcPort"),
		PusherServiceName:  viper.GetString("pusher.pusherServiceName"),
		GatewayServiceName: viper.GetString("pusher.gatewayServiceName"),
	}
	config.Kafka = Kafka{
		KafkaBrokerAddr:     viper.GetStringSlice("kafka.kafkaBrokerAddr"),
		KafkaSASLUsername:   viper.GetString("kafka.kafkaSASLUsername"),
		KafkaSASLPassword:   viper.GetString("kafka.kafkaSASLPassword"),
		KafkaMsgTopic:       viper.GetString("kafka.kafkaMsgTopic"),
		KafkaRetryPushTopic: viper.GetString("kafka.kafkaRetryPushTopic"),
	}
	config.Mysql = Mysql{
		MysqlAddr:     viper.GetString("mysql.mysqlAddr"),
		MysqlUsername: viper.GetString("mysql.mysqlUsername"),
		MysqlPassword: viper.GetString("mysql.mysqlUsername"),
	}
	config.Redis = Redis{
		RedisAddr:     viper.GetString("redis.redisAddr"),
		RedisPassword: viper.GetString("redis.redisPassword"),
	}
	config.Mongo = Mongo{
		MongoAddr:     viper.GetString("mongo.mongoAddr"),
		MongoUsername: viper.GetString("mongo.mongoUsername"),
		MongoPassword: viper.GetString("mongo.mongoPassword"),
	}
	config.Etcd = Etcd{
		EtcdAddr:     viper.GetStringSlice("etcd.etcdAddr"),
		EtcdUsername: viper.GetString("etcd.etcdUsername"),
		EtcdPassword: viper.GetString("etcd.etcdPassword"),
	}
	config.Secret = Secret{
		JwtSecret: viper.GetString("secret.jwtSecret"),
	}
}
