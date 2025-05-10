package db8

import (
	"deifzar/reportingm8/pkg/model8"
)

type Db8ReportSchedulingSettings8Interface interface {
	Get() (model8.Reportschedulingsettings8, error)
	Set(model8.PostReportschedulingsettings8) (model8.Reportschedulingsettings8, error)
}
