package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	KafkaTopics struct {
		TransactionTopic string `envconfig:"TRANSACTION_TOPIC"`
		CallbackTopic    string `envconfig:"CALLBACK_TOPIC"`
		DispatcherTopic  string `envconfig:"DISPATCHER_TOPIC"`
	}
	Kafka struct {
		Server string `envconfig:"KAFKA_SERVER"`
	}
	Database struct {
		HostName string `envconfig:"PG_HOST"`
		Database string `envconfig:"PG_DATABASE"`
		Username string `envconfig:"PG_USER"`
		Password string `envconfig:"PG_PASSWORD"`
	}
	Redis struct {
		HostName string `envconfig:"REDIS_HOST"`
		Password string `envconfig:"REDIS_PASSWORD"`
		Username string `envconfig:"REDIS_USER"`
	}
	Network struct {
		CallbackPrefix string `envconfig:"API_CALLBACK_PREFIX"`
		GateWayAUrl    string `envconfig:"GATEWAY_A_URL"`
		GateWayBUrl    string `envconfig:"GATEWAY_B_URL"`
	}
}

func ReadConfig() (*Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
