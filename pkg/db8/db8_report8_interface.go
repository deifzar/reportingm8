package db8

import (
	"deifzar/reportingm8/pkg/model8"
	// "github.com/gofrs/uuid/v5"
)

type Db8Report8Interface interface {
	InsertReport(model8.Report8) (bool, error)
}
