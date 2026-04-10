package section

type (
	Broker struct {
		Kafka BrokerKafka
	}

	BrokerKafka struct {
		Addresses     []string `envconfig:"APP_BROKER_KAFKA_ADDRESSES"`
		ConsumerGroup string   `envconfig:"APP_BROKER_KAFKA_CONSUMER_GROUP"`
		ClientID      string   `envconfig:"APP_BROKER_KAFKA_CLIENT_ID" default:"worker-service"`

		ModelOrder BrokerKafkaModelOrder
	}

	BrokerKafkaModelOrder struct {
		Created BrokerKafkaModelOrderCreated
	}

	BrokerKafkaModelOrderCreated struct {
		Topic         string `envconfig:"APP_BROKER_KAFKA_MODEL_ORDER_CREATED_TOPIC" default:"order.created"`
		ConsumerGroup string `envconfig:"APP_BROKER_KAFKA_MODEL_ORDER_CREATED_CONSUMER_GROUP"`
	}
)
