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

	diagLogger1, newDiag1Err := logger.New("diagnostic")
	if newDiag1Err != nil {
		fmt.Println(newDiag1Err)
		return
	}

	diagLogger2, newDiag2Err := logger.New("diagnostic")
	if newDiag2Err != nil {
		fmt.Println(newDiag2Err)
		return
	}

	defer accessLogger.Close()
	defer diagLogger1.Close()
	defer diagLogger2.Close()

	for i := 0; i < 100; i++ {
		time.Sleep(500 * time.Millisecond)

		accessLogger.Log("new access.")

		diagLogger1.Log("test message")
		diagLogger1.Fatal("test message")
		diagLogger1.Error("test message")
		diagLogger1.Warn("test message")
		diagLogger1.Notice("test message")
		diagLogger1.Info("test message")
		diagLogger1.Debug("test message")
	}

	for i := 0; i < 100; i++ {
		time.Sleep(500 * time.Millisecond)

		accessLogger.Log("new access.")

		diagLogger2.Log("test message")
		diagLogger2.Fatal("test message")
		diagLogger2.Error("test message")
		diagLogger2.Warn("test message")
		diagLogger2.Notice("test message")
		diagLogger2.Info("test message")
		diagLogger2.Debug("test message")
	}
}
