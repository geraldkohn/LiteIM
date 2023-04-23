package transfer

import (
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/kafka"
	servicediscovery "github.com/geraldkohn/im/pkg/common/service-discovery"
	"github.com/geraldkohn/im/pkg/common/setting"
)

type handle func(msg []byte, msgKey string) error // 处理函数, 处理接收到的消息

var (
	pushDiscovery servicediscovery.Discovery
	producer      *kafka.Producer
)

func init() {
	initServiceDiscovery()
	initKafkaProducer()
}

func initServiceDiscovery() {
	pushDiscovery = servicediscovery.NewEtcdDiscovery(setting.APPSetting.ServiceDiscovery.PushServiceName)
	go pushDiscovery.Watch()
}

func initKafkaProducer() {
	producer = kafka.NewKafkaProducer(constant.KafkaChatPushRetryTopic)
}

// 启动
func Run() {
	monogoConsumerHandler := NewMsgConsumerHandler()
	pushConsumerHandler := NewPushConsumerGroupHandler()
	go monogoConsumerHandler.consumerGroup.RegisterHandleAndConsumer(monogoConsumerHandler)
	go pushConsumerHandler.consumerGroup.RegisterHandleAndConsumer(pushConsumerHandler)
}
