package model8

import (
	"github.com/google/uuid"
)

type Scheduler8 struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Tags    []string  `json:"tags"`
	Nextrun string    `json:"nextrun"`
}
