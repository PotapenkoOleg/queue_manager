package monitor

import (
	"Monitor/config"
	"Monitor/queue_reader"
	"Monitor/queue_writer"
	"Monitor/sql_server_checker"
	"context"
	"log"
	"sync"
)

type Monitor struct {
	ctx   context.Context
	wg    *sync.WaitGroup
	cfg   *config.Config
	mutex *sync.RWMutex
}

func NewMonitor(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, mutex *sync.RWMutex) *Monitor {
	return &Monitor{
		ctx:   ctx,
		wg:    wg,
		cfg:   cfg,
		mutex: mutex,
	}
}

func (m *Monitor) Start() {
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()

		writeChan := make(chan string)
		checkChan := make(chan string)

		queueReader := queue_reader.NewQueueReader(m.ctx, m.wg, m.cfg, m.mutex, writeChan, checkChan)
		queueReader.Start()

		sqlServerChecker := sql_server_checker.NewSqlServerChecker(m.ctx, m.wg, m.cfg, m.mutex, writeChan, checkChan)
		sqlServerChecker.Start()

		queueWriter := queue_writer.NewQueueWriter(m.ctx, m.wg, m.cfg, m.mutex, writeChan)
		queueWriter.Start()

		log.Printf("Monitor started")

		for {
			select {
			case <-m.ctx.Done():
				// TODO: cleanup resources
				log.Printf("Monitor stopped")
				return
			}
		}
	}()
}
