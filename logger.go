package logger

/*
 * Module Dependencies
 */

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mozzzzy/logger/rotator"
)

/*
 * Types
 */

type Logger struct {
	fileWriter     *os.File
	logDirPathStr  string
	logFileMutex   *sync.Mutex
	logFileNameStr string
	innerLogger    *log.Logger
	logLevel       int
	rotator        *rotator.Rotator
}

type category struct {
	categoryName  string
	filePath      string
	logLevel      int
	logFileBytes  int64
	oldLogFileNum int
	logger        *Logger
}

/*
 * Constants and Package Scope Variables
 */

const (
	FATAL  int = 0
	ERROR  int = 1
	WARN   int = 2
	NOTICE int = 3
	INFO   int = 4
	DEBUG  int = 5
)

var (
	fatalStr   string = "FATAL"
	errorStr   string = "ERROR"
	warnStr    string = "WARN"
	noticeStr  string = "NOTICE"
	infoStr    string = "INFO"
	debugStr   string = "DEBUG"
	categories []category
)

/*
 * Functions
 */

func contain(str string, candidates []string) bool {
	for _, c := range candidates {
		if c == str {
			return true
		}
	}
	return false
}

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

func getCategoryNames() []string {
	var categoryNames []string
	for _, c := range categories {
		categoryNames = append(categoryNames, c.categoryName)
	}
	return categoryNames
}

func getCategory(categoryName string) *category {
	for index, c := range categories {
		if c.categoryName == categoryName {
			return &categories[index]
		}
	}
	return nil
}

func validateLogLevel(logLevel int) error {
	if logLevel < FATAL || logLevel > DEBUG {
		return errors.New(fmt.Sprintf("Unknown log level %v.", logLevel))
	}
	return nil
}

func AddCategory(
	categoryName string,
	filePath string,
	logLevel int,
	logFileBytes int64,
	oldLogFileNum int) error {
	// Get category instance from category name
	categoryNames := getCategoryNames()
	// If the category has already exist
	if contain(categoryName, categoryNames) {
		return errors.New(fmt.Sprintf("Category %v has already exist.", categoryName))
	}
	// Validate log level
	if err := validateLogLevel(logLevel); err != nil {
		return err
	}
	// Add new category
	categories = append(
		categories,
		category{
			categoryName,
			filePath,
			logLevel,
			logFileBytes,
			oldLogFileNum,
			nil,
		},
	)
	return nil
}

func GetLogLevelByStr(levelStr string) (int, error) {
	if contain(levelStr, []string{"DEBUG"}) {
		return DEBUG, nil
	}
	if contain(levelStr, []string{"INFO"}) {
		return INFO, nil
	}
	if contain(levelStr, []string{"NOTICE"}) {
		return NOTICE, nil
	}
	if contain(levelStr, []string{"WARN"}) {
		return WARN, nil
	}
	if contain(levelStr, []string{"ERROR"}) {
		return ERROR, nil
	}
	if contain(levelStr, []string{"FATAL"}) {
		return FATAL, nil
	}
	return -1, errors.New(fmt.Sprintf("Unknown log level \"%v\".", levelStr))
}

func New(categoryName string) (*Logger, error) {
	// Get category from category name
	category := getCategory(categoryName)
	if category == nil {
		return nil, errors.New(fmt.Sprintf("Category %v is not found.", categoryName))
	}
	if category.logger != nil {
		return category.logger, nil
	}
	// Create new Logger instance
	logger := new(Logger)
	// Set file path
	logDirPath := filepath.Dir(category.filePath)
	logFileName := filepath.Base(category.filePath)
	logger.logFileNameStr = logFileName
	logger.logDirPathStr = logDirPath

	// Set log level
	logger.logLevel = category.logLevel

	// Set mutex
	logger.logFileMutex = new(sync.Mutex)

	// Get file writer
	fileWriter, getFileWriterErr := getFileWriter(getFilePath(logDirPath, logFileName))
	if getFileWriterErr != nil {
		fileWriter.Close()
		return nil, getFileWriterErr
	}
	logger.fileWriter = fileWriter

	// Create logger
	innerLogger := log.New(fileWriter, "", log.LstdFlags|log.Lmicroseconds)
	logger.innerLogger = innerLogger

	// Create rotator
	logger.rotator = rotator.New(logDirPath, logFileName, category.logFileBytes, category.oldLogFileNum)

	return logger, nil
}

func (logger *Logger) noLockClose() error {
	return logger.fileWriter.Close()
}

func (logger *Logger) Close() error {
	// Lock
	logger.logFileMutex.Lock()
	defer logger.logFileMutex.Unlock()
	// Close
	return logger.noLockClose()
}

func (logger *Logger) Log(message string) error {
	// Lock
	logger.logFileMutex.Lock()
	defer logger.logFileMutex.Unlock()
	// Write log
	logger.innerLogger.Println(message)
	// Check rotatable or not
	isRotatable, isRotatableErr := logger.rotator.IsRotatable()
	if isRotatableErr != nil {
		return isRotatableErr
	}
	if isRotatable {
		// Close fileWriter
		logger.noLockClose()
		// Rotate log file
		logger.rotator.Rotate()
		logger.rotator.RemoveOldFile()
		// Reopen file writer
		fileWriter, getFileWriterErr :=
			getFileWriter(getFilePath(logger.logDirPathStr, logger.logFileNameStr))
		if getFileWriterErr != nil {
			return getFileWriterErr
		}
		logger.fileWriter = fileWriter
		// Recreate logger
		innerLogger := log.New(fileWriter, "", log.LstdFlags|log.Lmicroseconds)
		logger.innerLogger = innerLogger
	}
	return nil
}

func (logger *Logger) levelLog(message string, logLevel int, logLevelStr string) error {
	if logger.logLevel >= logLevel {
		return logger.Log(fmt.Sprintf("[%v] %s", logLevelStr, message))
	}
	return nil
}

func (logger *Logger) Fatal(message string) error {
	return logger.levelLog(message, FATAL, fatalStr)
}

func (logger *Logger) Error(message string) error {
	return logger.levelLog(message, ERROR, errorStr)
}

func (logger *Logger) Warn(message string) error {
	return logger.levelLog(message, WARN, warnStr)
}

func (logger *Logger) Notice(message string) error {
	return logger.levelLog(message, NOTICE, noticeStr)
}

func (logger *Logger) Info(message string) error {
	return logger.levelLog(message, INFO, infoStr)
}

func (logger *Logger) Debug(message string) error {
	return logger.levelLog(message, DEBUG, debugStr)
}
