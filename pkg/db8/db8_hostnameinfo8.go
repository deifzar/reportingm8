package db8

import (
	"database/sql"
	"deifzar/reportingm8/pkg/log8"
)

type Db8Hostnameinfo8 struct {
	Db *sql.DB
}

func NewDb8Hostnameinfo8(db *sql.DB) Db8Hostnameinfo8Interface {
	return &Db8Hostname8{Db: db}
}

func (m *Db8Hostname8) GetUniqueSoftware() ([]string, error) {
	query, err := m.Db.Query("SELECT DISTINCT (UNNEST(software)) AS soft FROM public.cptm8hostnameinfo")
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return nil, err
	}
	var software []string
	if query != nil {
		for query.Next() {
			var (
				s string
			)
			err := query.Scan(&s)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return nil, err
			}
			software = append(software, s)
		}
	}
	return software, nil
}

func (m *Db8Hostname8) CountUniqueSoftware() (int, error) {
	query, err := m.Db.Query("SELECT COUNT (t.soft) FROM (SELECT DISTINCT (UNNEST(software)) AS soft FROM public.cptm8hostnameinfo) as t")
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return 0, err
	}
	var nsoftware int
	if query != nil {
		for query.Next() {
			err := query.Scan(&nsoftware)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return 0, err
			}
		}
	}
	return nsoftware, nil
}
