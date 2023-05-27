package kafka

import (
	"sync"

	"LiteIM/pkg/common/config"

	"github.com/Shopify/sarama"
)

type Consumer struct {
	WG            sync.WaitGroup
	Topic         string
	PartitionList []int32
	Consumer      sarama.Consumer
}

func NewKafkaConsumer(topic string) *Consumer {
	consumerConfig := sarama.NewConfig()
	if config.Conf.Kafka.KafkaSASLUsername != "" && config.Conf.Kafka.KafkaSASLPassword != "" {
		consumerConfig.Net.SASL.Enable = true
		consumerConfig.Net.SASL.User = config.Conf.Kafka.KafkaSASLUsername
		consumerConfig.Net.SASL.Password = config.Conf.Kafka.KafkaSASLPassword
	}

	p := Consumer{}
	p.Topic = topic
	consumer, err := sarama.NewConsumer(config.Conf.Kafka.KafkaBrokerAddr, consumerConfig)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.Consumer = consumer

	partitionList, err := consumer.Partitions(p.Topic)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.PartitionList = partitionList

	return &p
}
