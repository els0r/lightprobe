package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/els0r/lightprobe/pkg/capture"
	"github.com/els0r/lightprobe/pkg/config"
	"github.com/els0r/lightprobe/pkg/flow"
)

const (
	FlushInterval = 300 // in seconds
)

func main() {

	// initialize command line
	//cliFlags := flags.Read()
	cfg := &config.Config{}

	// initiailze data & control channels
	writeChan := make(chan flow.Map)

	flushFlowsChan := make(chan time.Time)

	sigExitChan := make(chan os.Signal, 1)
	signal.Notify(sigExitChan, syscall.SIGTERM, os.Interrupt)

	// initialize capture process
	c := capture.New(cfg,
		//capture.WithPacketSource(&capture.MockPacketFetcher{}),
		capture.WithPacketSource(capture.NewConntrackPacketFetcher()),
	)
	err := c.Open()
	if err != nil {
		fmt.Printf("Error opening capture: %s\n", err)
		os.Exit(1)
	}

	c.Run(writeChan, flushFlowsChan)

	// initialize DB writer
	//w := write.New(
	//	writeChan,
	//	write.WithLocation(write.GoDB),
	//)

	// start ticker to flush all 5 minutes
	ticker := time.NewTicker(time.Second * time.Duration(FlushInterval))
	for {
		select {
		case <-ticker.C:
			// TODO: write out
			flushFlowsChan <- time.Now()

		case <-sigExitChan:
			fmt.Println("Shutting down")
			c.Close()

			time.Sleep(1 * time.Second)

			// TODO: clean up and last flush
			os.Exit(0)
		}
	}

}
