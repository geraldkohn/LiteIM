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
	tf *Transfer
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
		config.Conf.Transfer.PusherServiceName,
	)
	tf.db = db.NewDataBases(
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
	tf.retryPushProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   config.Conf.Kafka.KafkaBrokerAddr,
		SASLUsername: config.Conf.Kafka.KafkaSASLUsername,
		SASLPassword: config.Conf.Kafka.KafkaSASLPassword,
		Topic:        config.Conf.Kafka.KafkaMsgTopic,
	})
	tf.retryDBProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   config.Conf.Kafka.KafkaBrokerAddr,
		SASLUsername: config.Conf.Kafka.KafkaSASLUsername,
		SASLPassword: config.Conf.Kafka.KafkaSASLPassword,
		Topic:        config.Conf.Kafka.KafkaMsgTopic,
	})
	tf.consumerHandler = consumerHandler{
		kafka.NewConsumerGroup([]string{config.Conf.Kafka.KafkaMsgTopic, config.Conf.Kafka.KafkaRetryPushTopic}, utils.GenerateUID()),
		make(map[string]handle),
	}
	tf.consumerHandler.topicHandle[config.Conf.Kafka.KafkaMsgTopic] = tf.handleMsg
	tf.consumerHandler.topicHandle[config.Conf.Kafka.KafkaRetryPushTopic] = tf.handleRetryPush
}

// 阻塞函数
func Run() {
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
