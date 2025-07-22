package db8

import (
	"database/sql"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
)

type Db8User8 struct {
	Db *sql.DB
}

func NewDb8User8(db *sql.DB) Db8User8Interface {
	return &Db8User8{Db: db}
}

func (m *Db8User8) GetReportOwner() (model8.User8, error) {
	query, err := m.Db.Query("SELECT id, name, email, role, report FROM \"user\" WHERE report = true AND role = $1", model8.RoleUser)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return model8.User8{}, err
	}
	var user8 model8.User8
	if query != nil {
		if query.Next() {
			var (
				id     uuid.UUID
				name   string
				email  string
				role   model8.Roletype
				report bool
			)
			err := query.Scan(&id, &name, &email, &role, &report)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return model8.User8{}, err
			}
			user8 = model8.User8{Id: id, Name: name, Email: email, Role: role, Report: report}
		}
	}
	return user8, nil
}

func (m *Db8User8) GetReportAuthor() (model8.User8, error) {
	query, err := m.Db.Query("SELECT id, name, email, role, report FROM \"user\" WHERE report = true AND role = $1", model8.RoleAdmin)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return model8.User8{}, err
	}
	var user8 model8.User8
	if query != nil {
		if query.Next() {
			var (
				id     uuid.UUID
				name   string
				email  string
				role   model8.Roletype
				report bool
			)
			err := query.Scan(&id, &name, &email, &role, &report)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return model8.User8{}, err
			}
			user8 = model8.User8{Id: id, Name: name, Email: email, Role: role, Report: report}
		}
	}
	return user8, nil
}

func (m *Db8User8) GetUsersByRole(role model8.Roletype) ([]model8.User8, error) {
	query, err := m.Db.Query("SELECT id, name, email, role, report FROM \"user\" WHERE role = $1", role)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return nil, err
	}
	var user8 []model8.User8
	if query != nil {
		for query.Next() {
			var (
				id     uuid.UUID
				name   string
				email  string
				role   model8.Roletype
				report bool
			)
			err := query.Scan(&id, &name, &email, &role, &report)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return nil, err
			}
			user8 = append(user8, model8.User8{Id: id, Name: name, Email: email, Role: role, Report: report})
		}
	}
	return user8, nil
}

func (m *Db8User8) ExistReportOwner() bool {
	var exists bool
	err := m.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM \"user\" WHERE report = true AND role = $1)", model8.RoleUser).Scan(&exists)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return false
	}
	return exists
}

func (m *Db8User8) ExistReportAuthor() bool {
	var exists bool
	err := m.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM \"user\" WHERE report = true AND role = $1)", model8.RoleAdmin).Scan(&exists)
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return false
	}
	return exists
}
