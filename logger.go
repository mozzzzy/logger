package logger

/*
 * Module Dependencies
 */

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"./rotator"
)

/*
 * Types
 */

type Logger struct {
	fileWriter     *os.File
	logDirPathStr  string
	logFileMutex   *sync.Mutex
	logFileNameStr string
	logger         *log.Logger
	logLevel       int
	rotator        *rotator.Rotator
}

/*
 * Constants and Package Scope Variables
 */

const FATAL int = 0
const ERROR int = 1
const WARN int = 2
const NOTICE int = 3
const INFO int = 4
const DEBUG int = 5

var logFileMutex sync.Mutex

/*
 * Functions
 */

func getFilePath(dirPath string, fileName string) string {
	if strings.HasSuffix(dirPath, "/") {
		return dirPath + fileName
	} else {
		return dirPath + "/" + fileName
	}
}

func getFileWriter(logFilePath string) (*os.File, error) {
	// Validate parameter
	if logFilePath == "" {
		err := errors.New("Failed to create file writer. Length of logFilePath is 0.")
		return nil, err
	}
	// Get file writer
	logfile, logFileOpenErr :=
		os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	return logfile, logFileOpenErr
}

func New(
	logDirPath string,
	logFileName string,
	logLevel int,
	maxLogFileBytes int64,
	maxOldLogFileNum int,
) (*Logger, error) {
	logger := new(Logger)

	// Set file path
	logger.logFileNameStr = logFileName
	logger.logDirPathStr = logDirPath

	// Set log level
	logger.logLevel = logLevel

	// Set mutex
	logger.logFileMutex = &logFileMutex

	// Get file writer
	fileWriter, getFileWriterErr := getFileWriter(getFilePath(logDirPath, logFileName))
	if getFileWriterErr != nil {
		fileWriter.Close()
		return nil, getFileWriterErr
	}
	logger.fileWriter = fileWriter

	// Create logger
	innerLogger := log.New(fileWriter, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	logger.logger = innerLogger

	// Create rotator
	logger.rotator = rotator.New(logDirPath, logFileName, maxLogFileBytes, maxOldLogFileNum)

	return logger, nil
}

func (logger *Logger) Close() error {
	logger.logFileMutex.Lock()
	defer logger.logFileMutex.Unlock()
	return logger.fileWriter.Close()
}

func (logger *Logger) Log(message string) error {
	// Write log
	logger.logFileMutex.Lock()
	logger.logger.Println(message)
	logger.logFileMutex.Unlock()

	isRotatable, isRotatableErr := logger.rotator.IsRotatable()
	if isRotatableErr != nil {
		return isRotatableErr
	}
	if isRotatable {
		// Close fileWriter
		logger.Close()
		// Rotate log file
		logger.logFileMutex.Lock()
		logger.rotator.Rotate()
		logger.rotator.RemoveOldFile()
		logger.logFileMutex.Unlock()
		// Reopen file writer
		fileWriter, getFileWriterErr :=
			getFileWriter(getFilePath(logger.logDirPathStr, logger.logFileNameStr))
		if getFileWriterErr != nil {
			return getFileWriterErr
		}
		logger.fileWriter = fileWriter
		// Recreate logger
		innerLogger := log.New(fileWriter, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		logger.logger = innerLogger
	}
	return nil
}

func (logger *Logger) Fatal(message string) error {
	if logger.logLevel >= FATAL {
		return logger.Log(fmt.Sprintf("[FATAL] %s", message))
	}
	return nil
}

func (logger *Logger) Error(message string) error {
	if logger.logLevel >= ERROR {
		return logger.Log(fmt.Sprintf("[ERROR] %s", message))
	}
	return nil
}

func (logger *Logger) Warn(message string) error {
	if logger.logLevel >= WARN {
		return logger.Log(fmt.Sprintf("[WARN] %s", message))
	}
	return nil
}

func (logger *Logger) Notice(message string) error {
	if logger.logLevel >= NOTICE {
		return logger.Log(fmt.Sprintf("[NOTICE] %s", message))
	}
	return nil
}

func (logger *Logger) Info(message string) error {
	if logger.logLevel >= INFO {
		return logger.Log(fmt.Sprintf("[INFO] %s", message))
	}
	return nil
}

func (logger *Logger) Debug(message string) error {
	if logger.logLevel >= DEBUG {
		return logger.Log(fmt.Sprintf("[DEBUG] %s", message))
	}
	return nil
}
