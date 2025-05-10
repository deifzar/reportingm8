package db8

import (
	"deifzar/reportingm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
)

type Db8Hostname8Interface interface {
	Get() ([]model8.Hostname8, error)
	GetName(enabled bool) ([]string, error)
	GetAllHostnameByParentID(uuid.UUID) ([]model8.Hostname8, error)
	GetAllEnabled() ([]model8.Hostname8, error)
	GetAllEnabledByParentID(uuid.UUID) ([]model8.Hostname8, error)
	GetOneHostnameByID(uuid.UUID) (model8.Hostname8, error)
	GetOneHostnameByName(string) (model8.Hostname8, error)
	ValidPostBody(model8.PostHostname8) bool
	UpdateHostname(uuid.UUID, uuid.UUID, model8.PHostname8) (model8.Hostname8, error)
	Count(enabled bool) (int, error)
	CountFoundCurrentMonth() (int, error)
	CountFoundLastMonth() (int, error)
}
