package models

import (
	"fmt"
	"log"
	"time"

	"github.com/asdine/storm/v3"
)

const (
	LogTypeDebug LogType = "DEBUG"
	LogTypeInfo  LogType = "INFO"
	LogTypeWarn  LogType = "WARN"
	LogTypeError LogType = "ERROR"
)

type LogType string

type Logger struct {
	db *storm.DB
}

type Log struct {
	ID        int     `storm:"id,increment"`
	LogType   LogType `storm:"index"`
	Message   string
	CreatedAt time.Time `storm:"index"`
}

func NewLogger(db *storm.DB) *Logger {
	return &Logger{
		db: db,
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.storeLog(LogTypeDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.storeLog(LogTypeInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.storeLog(LogTypeWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.storeLog(LogTypeError, fmt.Sprintf(format, v...))
}

func (l *Logger) storeLog(logType LogType, msg string) {
	newLog := &Log{
		LogType:   logType,
		Message:   msg,
		CreatedAt: time.Now().UTC(),
	}

	if err := l.db.Save(newLog); err != nil {
		log.Print("Error adding log to database: ", err)
	}

	fmt.Println(newLog.String())
}

func (l *Log) String() string {
	return fmt.Sprintf("%s %-7s %s", l.CreatedAt.Format("2006-01-02 15:04:05.000"), "["+l.LogType+"]", l.Message)
}

func GetLogs(db *storm.DB, limit int) ([]Log, error) {
	var logs []Log
	err := db.AllByIndex("CreatedAt", &logs, storm.Limit(limit))
	return logs, err
}
