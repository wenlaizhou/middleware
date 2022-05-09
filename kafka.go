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

	// kafkaServers 后端服务地址
	kafkaServers string

	// timeoutSeconds 超时时间,单位 秒, 默认为20秒
	timeoutSeconds int
}

// MessageStats 消息统计信息
//
// 转换为json需要进行二次封装
type MessageStats struct {

	// KafkaServers 后端服务地址
	KafkaServers string

	// TimeoutSeconds 超时时间,单位 秒
	TimeoutSeconds int

	// writer统计信息
	kafka.WriterStats

	// 发送消息总量
	MessageCount uint64
}

// Send 发送多条消息
//
// messages 多条消息
func (this *MessageHandler) Send(messages ...kafka.Message) error {
	if messages == nil || len(messages) <= 0 {
		return errors.New("未传递message")
	}
	this.messageCounter += uint64(len(messages))
	return this.writer.WriteMessages(context.Background(), messages...)
}

// Stats 获取消息统计信息
func (this *MessageHandler) Stats() MessageStats {
	return MessageStats{
		KafkaServers:   this.kafkaServers,
		TimeoutSeconds: this.timeoutSeconds,
		WriterStats:    this.writer.Stats(),
		MessageCount:   this.messageCounter,
	}
}

// CreateMessageHandler 创建消息发送通道
//
// kafkaServers kafka broker 地址,使用,分隔集群
//
// timeoutSeconds 超时时间, 单位秒
func CreateMessageHandler(kafkaServers string, timeoutSeconds int) MessageHandler {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}
	dialer := &kafka.Dialer{
		Timeout:   time.Second * time.Duration(timeoutSeconds),
		DualStack: true, // DualStack enables RFC 6555-compliant "Happy Eyeballs"
	}
	return MessageHandler{
		messageCounter: 0,
		kafkaServers:   kafkaServers,
		timeoutSeconds: timeoutSeconds,
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  strings.Split(kafkaServers, ","),
			Balancer: &kafka.RoundRobin{},
			Dialer:   dialer,
		}),
	}
}
