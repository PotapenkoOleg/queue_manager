package queue_writer

import (
	"Monitor/config"
	"context"
	"log"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type QueueWriter struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	cfg       *config.Config
	mutex     *sync.RWMutex
	writeChan chan string
	producer  *kafka.Producer
}

func NewQueueWriter(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, mutex *sync.RWMutex, writeChan chan string) *QueueWriter {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Brokers,
	})
	if err != nil {
		log.Printf("failed to create producer: %w", err)
	}
	qw := &QueueWriter{
		ctx:       ctx,
		wg:        wg,
		cfg:       cfg,
		mutex:     mutex,
		writeChan: writeChan,
		producer:  producer,
	}
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("delivered to %s [%d] at offset %v\n",
						*ev.TopicPartition.Topic,
						ev.TopicPartition.Partition,
						ev.TopicPartition.Offset,
					)
				}
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
					log.Printf("Failed to send message: %v\n", err)
				}
			}
		}
	}()
	log.Printf("Queue Writer started ...")
}

func (qw *QueueWriter) sendMessage(message string) error {
	return qw.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &qw.cfg.Kafka.WriteTopic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
	}, nil)
}

func (qw *QueueWriter) close() {
	qw.producer.Flush(5000)
	qw.producer.Close()
}
