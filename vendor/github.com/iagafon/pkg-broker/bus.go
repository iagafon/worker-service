package broker

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/iagafon/pkg-broker/codec"
	butil "github.com/iagafon/pkg-broker/util"
)

type MessageHandler[T any] func(ctx context.Context, msg *T, headers map[string]string) error

type consumerGroupHandler[T any] struct {
	codec   codec.Codec[T]
	handler MessageHandler[T]
}

// Bus — интерфейс для работы с топиком Kafka.
type Bus[T any] interface {
	Send(ctx context.Context, msg *T) error

	SendWithHeaders(ctx context.Context, msg *T, headers map[string]string) error

	Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error

	QueueName() string

	Close() error
}

type kafkaBus[T any] struct {
	client        *KafkaClient
	codec         codec.Codec[T]
	topic         string
	consumerGroup string
	consumer      sarama.ConsumerGroup
}

func NewBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) (Bus[T], error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	if topic == "" {
		return nil, errors.New("topic is empty")
	}

	group := butil.Coalesce(consumerGroup, client.DefaultConsumerGroup())
	if group == "" {
		return nil, errors.New("consumerGroup is empty")
	}

	return &kafkaBus[T]{
		client:        client,
		codec:         c,
		topic:         topic,
		consumerGroup: group,
	}, nil
}

func MustKafkaBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) Bus[T] {
	b, err := NewBus[T](client, c, topic, consumerGroup)
	if err != nil {
		panic(err)
	}
	return b
}

func (b *kafkaBus[T]) Send(ctx context.Context, msg *T) error {
	return b.SendWithHeaders(ctx, msg, nil)
}

func (b *kafkaBus[T]) SendWithHeaders(ctx context.Context, msg *T, headers map[string]string) error {
	data, err := b.codec.Encode(msg)
	if err != nil {
		return err
	}

	messageKey := uuid.New().String()

	saramaMsg := &sarama.ProducerMessage{
		Topic: b.topic,
		Key:   sarama.StringEncoder(messageKey),
		Value: sarama.ByteEncoder(data),
	}

	if len(headers) > 0 {
		saramaMsg.Headers = make([]sarama.RecordHeader, 0, len(headers))
		for k, v := range headers {
			saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}

	_, _, err = b.client.Producer().SendMessage(saramaMsg)
	return err
}

func (b *kafkaBus[T]) Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error {
	consumer, err := b.client.NewConsumerGroup(b.consumerGroup)
	if err != nil {
		return err
	}
	b.consumer = consumer

	consumerHandler := &consumerGroupHandler[T]{codec: b.codec, handler: handler}

	if wg != nil {
		wg.Add(1)
	}
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		defer consumer.Close()

		for {
			if err := consumer.Consume(ctx, []string{b.topic}, consumerHandler); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Consumer error: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (b *kafkaBus[T]) QueueName() string {
	return b.topic
}

func (b *kafkaBus[T]) Close() error {
	if b.consumer != nil {
		return b.consumer.Close()
	}
	return nil
}

func (h *consumerGroupHandler[T]) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler[T]) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		headers := make(map[string]string)
		for _, header := range msg.Headers {
			headers[string(header.Key)] = string(header.Value)

		}

		decoded, err := h.codec.Decode(msg.Value)
		if err != nil {
			log.Printf("Failed to decode message: %v\n", err)
			session.MarkMessage(msg, "")
			continue
		}

		if err := h.handler(session.Context(), decoded, headers); err != nil {
			if butil.IsNotCriticalError(err) {
				session.MarkMessage(msg, "")
				continue
			}
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
