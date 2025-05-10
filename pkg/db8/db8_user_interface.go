package db8

import (
	"deifzar/reportingm8/pkg/model8"
)

type Db8User8Interface interface {
	GetReportOwner() (model8.User8, error)
	GetReportAuthor() (model8.User8, error)
	GetUsersByRole(role model8.Roletype) ([]model8.User8, error)
	ExistReportOwner() bool
	ExistReportAuthor() bool
}
