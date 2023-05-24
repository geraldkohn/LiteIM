package transfer

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"LiteIM/pkg/common/config"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/kafka"
	"LiteIM/pkg/common/logger"
	servicediscovery "LiteIM/pkg/common/service-discovery"
	"LiteIM/pkg/utils"

	"github.com/Shopify/sarama"
)

var (
	tf   *Transfer
	conf *config.Config
)

type Transfer struct {
	consumerHandler   consumerHandler            // 消费者组+不同Topic处理函数
	pushDiscovery     servicediscovery.Discovery // Push 组件服务发现
	db                *db.DataBases              // DB
	retryPushProducer *kafka.Producer            // 发送给 Push 失败的消息重写入 kafka
	retryDBProducer   *kafka.Producer            // 写入 DB 失败的消息重写入 kafka
}

func initTransfer() {
	tf = new(Transfer)
	tf.pushDiscovery = servicediscovery.NewEtcdDiscovery(
		conf.Transfer.PusherServiceName,
	)
	tf.db = db.NewDataBases(
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
	tf.retryPushProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   conf.Kafka.KafkaBrokerAddr,
		SASLUsername: conf.Kafka.KafkaSASLUsername,
		SASLPassword: conf.Kafka.KafkaSASLPassword,
		Topic:        conf.Kafka.KafkaMsgTopic,
	})
	tf.retryDBProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   conf.Kafka.KafkaBrokerAddr,
		SASLUsername: conf.Kafka.KafkaSASLUsername,
		SASLPassword: conf.Kafka.KafkaSASLPassword,
		Topic:        conf.Kafka.KafkaMsgTopic,
	})
	tf.consumerHandler = consumerHandler{
		kafka.NewConsumerGroup([]string{conf.Kafka.KafkaMsgTopic, conf.Kafka.KafkaRetryPushTopic}, utils.GenerateUID()),
		make(map[string]handle),
	}
	tf.consumerHandler.topicHandle[conf.Kafka.KafkaMsgTopic] = tf.handleMsg
	tf.consumerHandler.topicHandle[conf.Kafka.KafkaRetryPushTopic] = tf.handleRetryPush
}

func initConfig() {
	conf = config.Init()
}

// 阻塞函数
func Run() {
	initConfig()
	initTransfer()
	ctx, cancel := context.WithCancel(context.Background())
	go tf.pushDiscovery.Watch()                                               // 监听服务发现
	go tf.consumerHandler.RegisterHandleAndConsumer(ctx, &tf.consumerHandler) // 将消费者组注册

	// 实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	tf.pushDiscovery.Exit()
}

type handle func(value []byte, key string) error // 处理函数, 处理接收到的消息

// 处理消息
type consumerHandler struct {
	*kafka.ConsumerGroup                   // 封装的消费者组
	topicHandle          map[string]handle // 不同 Topic 的不同处理函数
}

func (consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Logger.Infof("MonogoConsumerHandler get a message %v", msg)
		fn := h.topicHandle[msg.Topic]
		err := fn(msg.Value, string(msg.Key))
		if err == nil {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}
