package middleware

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

// MessageHandler 消息发送通道
type MessageHandler struct {

	// messageCounter 消息计数器
	messageCounter uint64

	// writer 消息通道
	writer *kafka.Writer
}

// MessageStats 消息统计信息
type MessageStats struct {

	// writer统计信息
	kafka.WriterStats

	// 发送消息总量
	MessageCount uint64
}

// Send 发送多条消息
//
// messages 多条消息
func (this MessageHandler) Send(messages ...kafka.Message) error {
	if messages == nil || len(messages) <= 0 {
		return errors.New("未传递message")
	}
	this.messageCounter += uint64(len(messages))
	return this.writer.WriteMessages(context.Background(), messages...)
}

// Stats 获取消息统计信息
func (this MessageHandler) Stats() MessageStats {
	return MessageStats{
		WriterStats:  this.writer.Stats(),
		MessageCount: this.messageCounter,
	}
}

// CreateMessageHandler 创建消息发送通道
//
// kafkaServers kafka broker 地址,使用,分隔集群
//
// timeoutSeconds 超时时间, 单位秒
func CreateMessageHandler(kafkaServers string, timeoutSeconds int) MessageHandler {
	dialer := &kafka.Dialer{
		Timeout:   time.Second * time.Duration(timeoutSeconds),
		DualStack: true, // DualStack enables RFC 6555-compliant "Happy Eyeballs"
	}
	return MessageHandler{
		messageCounter: 0,
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  strings.Split(kafkaServers, ","),
			Balancer: &kafka.RoundRobin{},
			Dialer:   dialer,
		}),
	}
}