package model8

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Report8 struct {
	Id        uuid.UUID `json:"id"`
	Nvul      int       `json:"nvul"`
	Ndomain   int       `json:"ndomain"`
	Nhostname int       `json:"nhostname"`
	Ncritical int       `json:"ncritical"`
	Nhigh     int       `json:"nhigh"`
	Nmedium   int       `json:"nmedium"`
	Nlow      int       `json:"nlow"`
	Ninfo     int       `json:"ninfo"`
	Nservice  int       `json:"nservice"`
	Ntech     int       `json:"ntech"`
	Release   time.Time `json:"release"`
	Link      string    `json:"link"`
}
