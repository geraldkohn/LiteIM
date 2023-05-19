package transfer

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/kafka"
	"LiteIM/pkg/common/logger"
	servicediscovery "LiteIM/pkg/common/service-discovery"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

var tfer *Transfer

type Transfer struct {
	consumerHandler   consumerHandler            // 消费者组+不同Topic处理函数
	pushDiscovery     servicediscovery.Discovery // Push 组件服务发现
	db                *db.DataBases              // DB
	retryPushProducer *kafka.Producer            // 发送给 Push 失败的消息重写入 kafka
	retryDBProducer   *kafka.Producer            // 写入 DB 失败的消息重写入 kafka
}

func (tf *Transfer) initialize() {
	tf.pushDiscovery = servicediscovery.NewEtcdDiscovery(
		viper.GetString(viper.GetString("PushServiceName")),
	)
	tf.db = db.NewDataBases(
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
	tf.retryPushProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   viper.GetStringSlice("KafkaBrokerAddr"),
		SASLUsername: viper.GetString("KafkaSASLUsername"),
		SASLPassword: viper.GetString("KafkaSASLPassword"),
		Topic:        viper.GetString("KafkaRetryPushTopic"),
	})
	tf.retryDBProducer = kafka.NewKafkaProducer(kafka.KafkaProducerConfig{
		BrokerAddr:   viper.GetStringSlice("KafkaBrokerAddr"),
		SASLUsername: viper.GetString("KafkaSASLUsername"),
		SASLPassword: viper.GetString("KafkaSASLPassword"),
		Topic:        viper.GetString("KafkaRetryDBTopic"),
	})
	tf.consumerHandler = consumerHandler{
		kafka.NewConsumerGroup([]string{viper.GetString("KafkaMsgTopic"), viper.GetString("KafkaRetryPushTopic"), viper.GetString("KafkaRetryDBTopic")}, viper.GetString("KafkaConsumerGroup")),
		make(map[string]handle),
	}
	tf.consumerHandler.topicHandle[viper.GetString("KafkaMsgTopic")] = tf.handleMsg
	tf.consumerHandler.topicHandle[viper.GetString("KafkaRetryPushTopic")] = tf.handleRetryPush
}

// 阻塞函数
func Run() {
	tfer.initialize()
	ctx, cancel := context.WithCancel(context.Background())
	go tfer.pushDiscovery.Watch()                                                 // 监听服务发现
	go tfer.consumerHandler.RegisterHandleAndConsumer(ctx, &tfer.consumerHandler) // 将消费者组注册

	// 实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	tfer.pushDiscovery.Exit()
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
