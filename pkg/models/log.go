package models

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger"
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
	db *badger.DB
}

type Log struct {
	ID        string
	LogType   LogType
	Message   string
	CreatedAt time.Time
}

func NewLogger(db *badger.DB) *Logger {
	l := &Logger{
		db: db,
	}

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

	err := l.db.Update(func(txn *badger.Txn) error {
		newLog.ID = getID()

		jsonLog, _ := json.Marshal(newLog)
		return txn.Set([]byte("logs."+newLog.ID), jsonLog)
	})
	if err != nil {
		log.Print("Error adding log to database: ", err)
	}

	fmt.Println(newLog.String())
}

func (l *Log) String() string {
	return fmt.Sprintf("%s %-7s %s", l.CreatedAt.Format("2006-01-02 15:04:05.000"), "["+l.LogType+"]", l.Message)
}

func GetLogs(db *badger.DB, limit int) ([]Log, error) {
	var logs []Log

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   limit,
			Reverse:        true,
			AllVersions:    false,
			Prefix:         []byte("logs."),
		}
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("logs.")); it.ValidForPrefix([]byte("logs.")); it.Next() {
			item := it.Item()

			var log Log
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &log)
			})
			if err != nil {
				return err
			}

			logs = append(logs, log)
			if len(logs) >= limit {
				return nil
			}
		}
		return nil
	})
	return logs, err
}
