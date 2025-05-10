package db8

import (
	"database/sql"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/model8"

	_ "github.com/lib/pq"
)

type Db8ReportSchedulingSettings8 struct {
	Db *sql.DB
}

func NewDb8ReportSchedulingSettings8(db *sql.DB) Db8ReportSchedulingSettings8Interface {
	return &Db8ReportSchedulingSettings8{Db: db}
}

func (m *Db8ReportSchedulingSettings8) Get() (model8.Reportschedulingsettings8, error) {
	query, err := m.Db.Query("SELECT id, settings FROM cptm8reportschedulingsettings")
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}
	var reportScheduling model8.Reportschedulingsettings8
	if query != nil {
		if query.Next() {
			err := query.Scan(&reportScheduling.Id, &reportScheduling.Settings)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return model8.Reportschedulingsettings8{}, err
			}

		}
	}
	return reportScheduling, nil
}

func (m *Db8ReportSchedulingSettings8) Set(post model8.PostReportschedulingsettings8) (model8.Reportschedulingsettings8, error) {
	var d model8.Reportschedulingsettings8
	d, err := m.Get()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}

	stmt, err := tx.Prepare("UPDATE cptm8reportschedulingsettings SET settings = $1 WHERE id = $2")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}
	_, err2 := stmt.Exec(post.Settings, d.Id)
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return model8.Reportschedulingsettings8{}, err2
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}
	d, err = m.Get()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Reportschedulingsettings8{}, err
	}
	return d, nil
}
