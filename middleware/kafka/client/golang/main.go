package main

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
)

func main() {
	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup([]string{"127.0.0.1:9092"}, "test-group", config)
	if err != nil {
		panic(err)
	}

	err = consumerGroup.Consume(context.Background(), []string{"test1", "test2"}, &handler{})
	if err != nil {
		panic(err)
	}
}

type handler struct {
}

func (h *handler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println("-------------------------------------------")
		fmt.Printf("Message Topic: %s\n", msg.Topic)
		fmt.Printf("Message Value: %s\n", string(msg.Value))
		fmt.Printf("Message Partition: %d\n", msg.Partition)
		fmt.Printf("Message Offset: %d\n", msg.Offset)
		fmt.Println("-------------------------------------------")
		session.MarkMessage(msg, "")
	}
	return nil
}
