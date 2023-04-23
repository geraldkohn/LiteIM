package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/Shopify/sarama"
)

// 消费者组处理结构体
type consumerGroupHandler struct {
}

func (consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() { // 读消息 channel
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", msg.Value, msg.Timestamp, msg.Topic)
		session.MarkMessage(msg, "") // 将 Kafka 消息标记为已消费
	}
	return nil
}

// 消费者组
func SaramaConsumerGroup() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = false
	config.Version = sarama.V0_10_2_0                     // specify appropriate version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 未找到组消费位移的时候从哪边开始消费

	// broker, groupID, config
	group, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "my-group", config)
	if err != nil {
		panic(err)
	}
	defer func() { _ = group.Close() }()

	// Track errors
	go func() {
		for err := range group.Errors() {
			fmt.Println("ERROR", err)
		}
	}()
	fmt.Println("Consumed start")
	// Iterate over consumer sessions.
	ctx := context.Background()
	for {
		topics := []string{"my_topic"}
		handler := consumerGroupHandler{}

		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		// 这个函数需要写在 for 循环内, 因为可能重新分配分区
		// topic-name, consumer-group-handler
		err := group.Consume(ctx, topics, handler)
		if err != nil {
			panic(err)
		}
	}
}

// 消费者
func SaramaConsumer() {
	// broker, config
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, sarama.NewConfig())
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// topic-name, partition-name, offset
	partitionConsumer, err := consumer.ConsumePartition("my_topic", 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumed := 0
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed message offset %d\n", msg.Offset)
			consumed++
		case <-signals:
			break ConsumerLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
}

// 异步生产者Goroutines
func SyncProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	// broker, config(处理消息是否发送成功, 处理发送产生的错误)
	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		panic(err)
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var (
		wg                                  sync.WaitGroup
		enqueued, successes, producerErrors int
	)

	// config.Producer.Return.Successes = true
	// 必须读取生产者成功发送并持久化到 kafka 的消息
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.Successes() {
			successes++
		}
	}()

	// config.Producer.Return.Errors = true
	// 必须读取生产者在持久化消息的 kafka 中的错误
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			producerErrors++
		}
	}()

ProducerLoop:
	for {
		// producer 生产的消息, Topic, Value
		message := &sarama.ProducerMessage{Topic: "my_topic", Value: sarama.StringEncoder("testing 456")}
		select {
		// 写入消息
		case producer.Input() <- message:
			enqueued++

		case <-signals:
			// 退出信号
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}

	wg.Wait()

	log.Printf("Successfully produced: %d; errors: %d\n", successes, producerErrors)
}

// 异步生产者Select
func SyncProducerSelect() {
	// broker, config(不处理消息是否发送成功, 不处理发送的错误)
	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var enqueued, producerErrors int
ProducerLoop:
	for {
		select {
		case producer.Input() <- &sarama.ProducerMessage{Topic: "my_topic", Key: nil, Value: sarama.StringEncoder("testing 123")}:
			enqueued++
		case err := <-producer.Errors():
			log.Println("Failed to produce message", err)
			producerErrors++
		case <-signals:
			break ProducerLoop
		}
	}
	log.Printf("Enqueued: %d; errors: %d\n", enqueued, producerErrors)
}

// 同步生产模式
func SaramaProducer() {
	// broker, config(不处理发送消息是否成功, 不处理发送的错误)
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// 同步生产消息
	msg := &sarama.ProducerMessage{Topic: "my_topic", Value: sarama.StringEncoder("testing 123")}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}

}

func main() {
	//生产者
	go SyncProducer()
	//go SaramaProducer()
	//go SyncProducerSelect()

	//消费者
	SaramaConsumerGroup()
	//SaramaConsumer()

}
