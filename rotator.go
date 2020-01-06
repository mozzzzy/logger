package logger

/*
 * Module Dependencies
 */

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

/*
 * Types
 */

type Rotator struct {
	logFileNameStr   string
	logDirPathStr    string
	maxLogFileBytes  int64
	maxOldLogFileNum int
}

/*
 * Constants and Package Scope Variables
 */

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
	logDirPath string, logFileName string, maxLogFileBytes int64, maxOldLogFileNum int) *Rotator {
	rotator := new(Rotator)

	// Set file path
	rotator.logFileNameStr = logFileName
	rotator.logDirPathStr = logDirPath
	// Set max log file size
	rotator.maxLogFileBytes = maxLogFileBytes
	// Set max old log file number
	rotator.maxOldLogFileNum = maxOldLogFileNum

	return rotator
}

func (rotator *Rotator) getLogFiles() ([]string, error) {
	// Get all file list
	var logFiles []string
	files, readDirErr := ioutil.ReadDir(rotator.logDirPathStr)
	if readDirErr != nil {
		return nil, readDirErr
	}
	// Extract only log files
	for _, file := range files {
		// Skip if file is directory
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name()+"-", rotator.logFileNameStr) {
			logFiles = append(logFiles, file.Name())
		}
	}
	return logFiles, nil
}

func (rotator *Rotator) RemoveOldFile() error {
	// Get log file list
	logFiles, getLogFilesErr := rotator.getLogFiles()
	if getLogFilesErr != nil {
		return getLogFilesErr
	}
	// Sort log file list
	sort.Strings(logFiles)
	// Delete old log files
	for len(logFiles) > 1 && len(logFiles) > rotator.maxOldLogFileNum {
		os.Remove(getFilePath(rotator.logDirPathStr, logFiles[0]))
		logFiles = unset(logFiles, 0)
	}
	// Delete old log files
	return nil
}

func (rotator *Rotator) IsRotatable() (bool, error) {
	// Get log file size
	logFileBytes, getFileBytesErr :=
		getFileBytes(getFilePath(rotator.logDirPathStr, rotator.logFileNameStr))
	if getFileBytesErr != nil {
		return false, getFileBytesErr
	}

	// If file size is bigger than max size
	if logFileBytes >= rotator.maxLogFileBytes {
		return true, nil
	}
	return false, nil
}

func (rotator *Rotator) Rotate() error {
	// Rename file
	oldFilePath :=
		fmt.Sprintf(
			"%v-%v",
			getFilePath(rotator.logDirPathStr, rotator.logFileNameStr),
			time.Now().Unix(),
		)
	renameErr := os.Rename(
		getFilePath(rotator.logDirPathStr, rotator.logFileNameStr), oldFilePath)
	if renameErr != nil {
		return renameErr
	}
	return nil
}
