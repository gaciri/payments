package utils

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func NewConsumer(host string, groupId string) (*kafka.Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": host,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	return c, err
}
