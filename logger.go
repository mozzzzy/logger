package logger

/*
 * Module Dependencies
 */

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

/*
 * Types
 */

type Logger struct {
	logFileNameStr   string
	logDirPathStr    string
	fileWriter       *os.File
	logFileMutex     *sync.Mutex
	logger           *log.Logger
	maxLogFileBytes  int64
	maxOldLogFileNum int
}

/*
 * Constants and Package Scope Variables
 */

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

func getFileBytes(logFilePath string) (int64, error) {
	fileStat, err := os.Stat(logFilePath)
	if err != nil {
		return -1, err
	}

	return fileStat.Size(), nil
}

func unset(s []string, i int) []string {
	if i >= len(s) {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

func New(
	logDirPath string,
	logFileName string,
	maxLogFileBytes int64,
	maxOldLogFileNum int,
) (*Logger, error) {
	logger := new(Logger)

	// Set file path
	logger.logFileNameStr = logFileName
	logger.logDirPathStr = logDirPath

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

	// Set max log file size
	logger.maxLogFileBytes = maxLogFileBytes

	// Set max old log file num
	logger.maxOldLogFileNum = maxOldLogFileNum

	return logger, nil
}

func (logger *Logger) Close() error {
	logger.logFileMutex.Lock()
	defer logger.logFileMutex.Unlock()
	return logger.fileWriter.Close()
}

func (logger *Logger) getLogFiles() ([]string, error) {
	var logFiles []string
	files, readDirErr := ioutil.ReadDir(logger.logDirPathStr)
	if readDirErr != nil {
		return nil, readDirErr
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), logger.logFileNameStr) {
			logFiles = append(logFiles, file.Name())
		}
	}
	return logFiles, nil
}

func (logger *Logger) removeOldFile() error {
	logFiles, getLogFilesErr := logger.getLogFiles()
	if getLogFilesErr != nil {
		return getLogFilesErr
	}
	sort.Strings(logFiles)
	for len(logFiles) > 1 && len(logFiles) > logger.maxOldLogFileNum {
		os.Remove(getFilePath(logger.logDirPathStr, logFiles[1]))
		logFiles = unset(logFiles, 1)
	}

	return nil
}

func (logger *Logger) lotate() error {
	// Get log file size
	logFileBytes, getFileBytesErr :=
		getFileBytes(getFilePath(logger.logDirPathStr, logger.logFileNameStr))
	if getFileBytesErr != nil {
		return getFileBytesErr
	}

	// If file size is bigger than max size
	if logFileBytes >= logger.maxLogFileBytes {
		// Close file writer
		logger.Close()

		// Rename file
		oldFilePath :=
			fmt.Sprintf(
				"%v-%v",
				getFilePath(logger.logDirPathStr, logger.logFileNameStr),
				time.Now().Unix(),
			)
		renameErr := os.Rename(
			getFilePath(logger.logDirPathStr, logger.logFileNameStr), oldFilePath)
		if renameErr != nil {
			return renameErr
		}

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

func (logger *Logger) Log(message string) {
	logger.logFileMutex.Lock()
	logger.logger.Println(message)
	logger.logFileMutex.Unlock()

	logger.lotate()
	logger.removeOldFile()
}
