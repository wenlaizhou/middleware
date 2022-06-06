package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

// MessageHandler 消息发送通道
type MessageHandler struct {

	// sendMessageCounter 消息计数器
	sendMessageCounter uint64

	// receiveMessageCounter 消息计数器
	receiveMessageCounter uint64

	// startTime 启动时间
	startTime time.Time

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
	SendMessageCounter uint64

	// receiveMessageCounter 消息计数器
	ReceiveMessageCounter uint64

	// 启动时间
	StartTime time.Time
}

// Send 发送多条消息
//
// messages 多条消息
func (this *MessageHandler) Send(messages ...kafka.Message) error {
	if messages == nil || len(messages) <= 0 {
		return errors.New("未传递message")
	}
	if this.writer == nil {
		return errors.New("kafka连接初始化错误")
	}
	this.sendMessageCounter += uint64(len(messages))
	return this.writer.WriteMessages(context.Background(), messages...)
}

// Stats 获取消息统计信息
func (this *MessageHandler) Stats() MessageStats {
	res := MessageStats{
		KafkaServers:          this.kafkaServers,
		TimeoutSeconds:        this.timeoutSeconds,
		SendMessageCounter:    this.sendMessageCounter,
		ReceiveMessageCounter: this.receiveMessageCounter,
		StartTime:             this.startTime,
	}
	if this.writer != nil {
		res.WriterStats = this.writer.Stats()
	}
	return res
}

// CreateMessageHandler 创建消息发送通道
//
// kafkaServers kafka broker 地址,使用,分隔集群
//
// timeoutSeconds 超时时间, 单位秒
func CreateMessageHandler(kafkaServers string, timeoutSeconds int) (MessageHandler, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}
	dialer := &kafka.Dialer{
		Timeout:   time.Second * time.Duration(timeoutSeconds),
		DualStack: true, // DualStack enables RFC 6555-compliant "Happy Eyeballs"
	}
	res := MessageHandler{
		sendMessageCounter:    0,
		receiveMessageCounter: 0,
		kafkaServers:          kafkaServers,
		timeoutSeconds:        timeoutSeconds,
		startTime:             time.Now(),
	}
	if writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  strings.Split(kafkaServers, ","),
		Balancer: &kafka.RoundRobin{},
		Dialer:   dialer,
	}); writer != nil {
		res.writer = writer
		return res, nil
	} else {
		return res, errors.New("kafka 连接错误")
	}
}

// RegisterConsumer 注册消息订阅
func (this *MessageHandler) RegisterConsumer(topic string, groupId string, cacheSeconds int, handler func([]kafka.Message)) {
	go func() {
		logger := GetLogger(fmt.Sprintf("consumer-%v-%v", topic, groupId))
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(this.kafkaServers, ","),
			Topic:   topic,
			GroupID: groupId,
			Dialer: &kafka.Dialer{
				Timeout:   time.Second * time.Duration(this.timeoutSeconds),
				DualStack: true, // DualStack enables RFC 6555-compliant "Happy Eyeballs"
			},

			// flushes commits to Kafka every second
			// By default, CommitMessages will synchronously commit offsets to Kafka.
			// For improved performance, you can instead periodically commit offsets to Kafka by setting CommitInterval on the ReaderConfig.
			CommitInterval: time.Second,
		})
		cache := []kafka.Message{}
		next := time.Now().Add(time.Duration(cacheSeconds) * time.Second)
		for {
			if cacheSeconds > 0 && time.Now().After(next) {
				if len(cache) > 0 {
					handler(cache)
				}
				cache = nil
				next = time.Now().Add(time.Duration(cacheSeconds) * time.Second)
			}
			if m, err := r.ReadMessage(context.Background()); err == nil {
				this.receiveMessageCounter++
				logger.InfoF("消费消息: offset: %v, 消息时间: %v, val: %v", m.Offset, m.Time.Format(TimeFormat), string(m.Value))
				if cacheSeconds <= 0 {
					handler([]kafka.Message{m})
					continue
				}
				cache = append(cache, m)
			} else {
				logger.ErrorF("消费消息错误, %v, %v, %v, 停止消费, 错误信息: %v", this.kafkaServers, topic, groupId, err.Error())
				break
			}
		}
	}()
}

type KafkaPartition struct {
	// Name of the topic that the partition belongs to, and its index in the
	// topic.
	Topic string `json:"topic"`
	ID    int    `json:"id"`

	// Leader, replicas, and ISR for the partition.
	Leader   KafkaBroker   `json:"leader"`
	Replicas []KafkaBroker `json:"replicas"`
	Isr      []KafkaBroker `json:"isr"`
}

type KafkaBroker struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	ID   int    `json:"id"`
	Rack string `json:"rack"`
}

// ClusterInfo 获取kafka集群信息
func (this *MessageHandler) ClusterInfo() ([]KafkaPartition, error) {
	result := []KafkaPartition{}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(this.timeoutSeconds))
	defer cancel()
	conn, err := kafka.DialContext(timeoutCtx, "tcp", this.kafkaServers)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return result, err
	}

	for _, p := range partitions {
		result = append(result, KafkaPartition{
			Topic:    p.Topic,
			ID:       p.ID,
			Leader:   toKafkaBroker(p.Leader),
			Replicas: toKafkaBrokers(p.Replicas),
			Isr:      toKafkaBrokers(p.Isr),
		})
	}

	return result, nil
}

func toKafkaBroker(broker kafka.Broker) KafkaBroker {

	return KafkaBroker{
		Host: broker.Host,
		Port: broker.Port,
		ID:   broker.ID,
		Rack: broker.Rack,
	}
}

func toKafkaBrokers(brokers []kafka.Broker) []KafkaBroker {
	res := []KafkaBroker{}
	if len(brokers) <= 0 {
		return res
	}
	for _, broker := range brokers {
		res = append(res, KafkaBroker{
			Host: broker.Host,
			Port: broker.Port,
			ID:   broker.ID,
			Rack: broker.Rack,
		})
	}
	return res
}
