package main

/*
 * Module Dependencies
 */

import (
	"fmt"
	"time"

	"github.com/mozzzzy/logger"
)

/*
 * Types
 */

/*
 * Constants
 */

/*
 * Functions
 */

func main() {
	logger, err := logger.New("./test.log", "DEBUG", 1024*1024, 5)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 100; i++ {
		time.Sleep(500 * time.Millisecond)
		logger.Log("test message")
		logger.Fatal("test message")
		logger.Error("test message")
		logger.Warn("test message")
		logger.Notice("test message")
		logger.Info("test message")
		logger.Debug("test message")
	}

	logger.Close()
}
