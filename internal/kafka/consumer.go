package kafka

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/models"
	"order-service/internal/storage"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer sarama.Consumer
	storage  *storage.PostgresStorage
	cache    *storage.Cache
	topic    string
}

func NewKafkaConsumer(brokers []string, topic string, storage *storage.PostgresStorage, cache *storage.Cache) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V4_0_0_0

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		storage:  storage,
		cache:    cache,
		topic:    topic,
	}, nil
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	partitionConsumer, err := kc.consumer.ConsumePartition(kc.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			if err := kc.storage.SaveOrder(order); err != nil {
				log.Printf("Failed to save order: %v", err)
				continue
			}

			kc.cache.Set(order)
			log.Printf("Order %s processed and cached", order.OrderUID)

		case err := <-partitionConsumer.Errors():
			log.Printf("Kafka error: %v", err)

		case <-ctx.Done():
			return nil
		}
	}
}
