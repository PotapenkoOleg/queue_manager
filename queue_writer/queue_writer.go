package queue_writer

import (
	"Monitor/config"
	"context"
	"log"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const ProducerFlushMilliseconds = 5000

type QueueWriter struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	cfg       *config.Config
	mutex     *sync.RWMutex
	writeChan chan string
	producer  *kafka.Producer
	topic     string
}

func NewQueueWriter(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, mutex *sync.RWMutex, writeChan chan string) *QueueWriter {
	mutex.RLock()
	brokers := cfg.KafkaConfig.Brokers
	topic := cfg.KafkaConfig.WriteTopic
	mutex.RUnlock()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	qw := &QueueWriter{
		ctx:       ctx,
		wg:        wg,
		cfg:       cfg,
		mutex:     mutex,
		writeChan: writeChan,
		producer:  producer,
		topic:     topic,
	}
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Fatalf("delivery failed: %v\n", ev.TopicPartition.Error)
				}
				log.Printf("delivered to %s [%d] at offset %v\n",
					*ev.TopicPartition.Topic,
					ev.TopicPartition.Partition,
					ev.TopicPartition.Offset,
				)
			}
		}
	}()
	return qw
}

func (qw *QueueWriter) Start() {
	go func() {
		for {
			select {
			case <-qw.ctx.Done():
				qw.close()
				log.Printf("Queue Writer stopped")
				return
			case message := <-qw.writeChan:
				log.Printf("Queue Writer: Chan Message: %s\n", message)
				err := qw.sendMessage(message)
				if err != nil {
					log.Fatalf("Failed to send message: %v\n", err)
				}
			}
		}
	}()
	log.Printf("Queue Writer started ...")
}

func (qw *QueueWriter) sendMessage(message string) error {
	return qw.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &qw.topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
	}, nil)
}

func (qw *QueueWriter) close() {
	qw.producer.Flush(ProducerFlushMilliseconds)
	qw.producer.Close()
}
