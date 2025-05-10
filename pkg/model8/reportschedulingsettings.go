package model8

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gofrs/uuid/v5"
)

type DocumentSchedulingSettings struct {
	DocumentTimebase  string `json:"documenttimebase,omitempty" binding:"required,oneof=weekly monthly"`
	DocumentFrequency int    `json:"documentfrequency,omitempty" binding:"required,number,min=1,max=12"`
}

type EmailSchedulingSettings struct {
	EmailTimebase  string `json:"emailtimebase,omitempty" binding:"required,oneof=weekly monthly"`
	EmailFrequency int    `json:"emailfrequency,omitempty" binding:"required,number,min=1,max=12"`
}

type SchedulingSettings8 struct {
	CompanyNameReport string                     `json:"companynamereport,omitempty" binding:"required,ascii,min=2,max=20"`
	Document          DocumentSchedulingSettings `json:"document" binding:"required"`
	Email             EmailSchedulingSettings    `json:"email" binding:"required"`
}

type PostReportschedulingsettings8 struct {
	Settings SchedulingSettings8 `json:"settings" binding:"required"`
}

type Reportschedulingsettings8 struct {
	Id       uuid.UUID           `json:"id,omitempty"`
	Settings SchedulingSettings8 `json:"settings"`
}

// Make the SchedulingSettings8 struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (s SchedulingSettings8) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Make the SchedulingSettings8 struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (s *SchedulingSettings8) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}
