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
	if err := logger.AddCategory(
		"access",
		"./access.log",
		logger.INFO,
		1024*1024,
		5); err != nil {
		fmt.Println(err)
		return
	}
	if err := logger.AddCategory(
		"diagnostic",
		"./diag.log",
		logger.DEBUG,
		1024*1024,
		5); err != nil {
		fmt.Println(err)
		return
	}

	accessLogger, newAccessErr := logger.New("access")
	if newAccessErr != nil {
		fmt.Println(newAccessErr)
		return
	}

	diagLogger, newDiagErr := logger.New("diagnostic")
	if newDiagErr != nil {
		fmt.Println(newDiagErr)
		return
	}
	defer accessLogger.Close()
	defer diagLogger.Close()

	for i := 0; i < 100; i++ {
		time.Sleep(500 * time.Millisecond)

		accessLogger.Log("new access.")

		diagLogger.Log("test message")
		diagLogger.Fatal("test message")
		diagLogger.Error("test message")
		diagLogger.Warn("test message")
		diagLogger.Notice("test message")
		diagLogger.Info("test message")
		diagLogger.Debug("test message")
	}
}
