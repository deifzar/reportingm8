package db8

import (
	"database/sql"
	"deifzar/reportingm8/pkg/log8"
	// "time"
)

type Db8Service8 struct {
	Db *sql.DB
}

func NewDb8Service8(db *sql.DB) Db8Service8Interface {
	return &Db8Service8{Db: db}
}

func (m *Db8Service8) GetUniqueServices() ([]string, error) {
	query, err := m.Db.Query("SELECT DISTINCT(service) FROM public.cptm8service")
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return nil, err
	}
	var services []string
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
			services = append(services, s)
		}
	}
	return services, nil
}

func (m *Db8Service8) CountUniqueServices() (int, error) {
	query, err := m.Db.Query("SELECT COUNT (t.serv) FROM (SELECT DISTINCT(service) as serv FROM public.cptm8service) as t")
	if err != nil {
		log8.BaseLogger.Debug().Stack().Msg(err.Error())
		return 0, err
	}
	var nservices int
	if query != nil {
		for query.Next() {
			err := query.Scan(&nservices)
			if err != nil {
				log8.BaseLogger.Debug().Stack().Msg(err.Error())
				return 0, err
			}
		}
	}
	return nservices, nil
}
