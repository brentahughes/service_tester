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

var logger *Logger

type LogType string

type Logger struct {
	db *storm.DB
	ch chan Log
}

type Log struct {
	ID        int     `storm:"id,increment"`
	LogType   LogType `storm:"index"`
	Message   string
	CreatedAt time.Time `storm:"index"`
}

func NewLogger(db *storm.DB) *Logger {
	l := &Logger{
		db: db,
		ch: make(chan Log, 0),
	}

	go l.logSaver()

	logger = l
	return l
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

	if err := l.db.From("log").Save(newLog); err != nil {
		log.Print("Error adding log to database: ", err)
	}

	fmt.Println(newLog.String())
}

func (l *Logger) logSaver() {
	t := time.NewTicker(time.Minute)

	var logs []Log
	for {
		select {
		case newLog := <-l.ch:
			logs = append(logs, newLog)

			if len(logs) > 100 {
				l.saveLogs(logs)
				logs = make([]Log, 0)
			}
		case <-t.C:
			if len(logs) > 0 {
				l.saveLogs(logs)
				logs = make([]Log, 0)
			}
		}
	}

}

func (l *Logger) saveLogs(logs []Log) {
	n := l.db.WithBatch(true).From("log")

	for _, lo := range logs {
		if err := n.Save(lo); err != nil {
			log.Print("Error adding log to database: ", err)
		}
	}
}

func (l *Log) String() string {
	return fmt.Sprintf("%s %-7s %s", l.CreatedAt.Format("2006-01-02 15:04:05.000"), "["+l.LogType+"]", l.Message)
}

func GetLogs(db *storm.DB, limit int) ([]Log, error) {
	var logs []Log
	err := db.From("log").AllByIndex("CreatedAt", &logs, storm.Limit(limit), storm.Reverse())
	return logs, err
}
