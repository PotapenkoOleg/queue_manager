package main

import (
	"Monitor/config"
	"Monitor/monitor"
	"Monitor/version"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	version.PrintSeparator()
	version.PrintBanner()
	version.PrintSeparator()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	mutex := &sync.RWMutex{}

	mutex.Lock()
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		panic(err)
	}
	log.Printf("Config loaded ...")
	mutex.Unlock()

	monitorApp := monitor.NewMonitor(ctx, &wg, cfg, mutex)
	monitorApp.Start()

	<-ctx.Done()

	fmt.Println("Shutting down:", ctx.Err())
	wg.Wait()
	fmt.Println("Shutting down: DONE", ctx.Err())
}
