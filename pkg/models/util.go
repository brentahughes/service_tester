package models

import (
	"github.com/google/uuid"
)

const currentHostID = "current-host"

func getID() string {
	return uuid.New().String()
}
