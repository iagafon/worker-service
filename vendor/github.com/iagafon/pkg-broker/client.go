package broker

import (
	"github.com/IBM/sarama"
	butil "github.com/iagafon/pkg-broker/util"
)

type KafkaConfig struct {
	Addresses     []string // Адреса брокеров: ["localhost:9092"]
	ConsumerGroup string   // Consumer group по умолчанию
	ClientID      string   // Идентификатор клиента для логов Kafka
}

type KafkaClient struct {
	addresses            []string
	saramaCfg            *sarama.Config
	producer             sarama.SyncProducer
	defaultConsumerGroup string
}

func NewKafkaClient(cfg KafkaConfig) (*KafkaClient, error) {
	clientID := butil.Coalesce(cfg.ClientID, cfg.ConsumerGroup)
	defaultGroup := butil.Coalesce(cfg.ConsumerGroup, cfg.ClientID)

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.ClientID = clientID

	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	producer, err := sarama.NewSyncProducer(cfg.Addresses, config)
	if err != nil {
		return nil, err
	}

	return &KafkaClient{
		addresses:            cfg.Addresses,
		saramaCfg:            config,
		producer:             producer,
		defaultConsumerGroup: defaultGroup,
	}, nil
}

func (c *KafkaClient) Producer() sarama.SyncProducer {
	return c.producer
}

func (c *KafkaClient) NewConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	return sarama.NewConsumerGroup(c.addresses, groupID, c.saramaCfg)
}

func (c *KafkaClient) DefaultConsumerGroup() string {
	return c.defaultConsumerGroup
}

func (c *KafkaClient) Close() error {
	if c.producer != nil {
		return c.producer.Close()
	}
	return nil
}
