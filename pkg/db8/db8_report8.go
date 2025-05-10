package db8

import (
	"database/sql"
	"deifzar/reportingm8/pkg/log8"
	"deifzar/reportingm8/pkg/model8"
)

type Db8Report8 struct {
	Db *sql.DB
}

func NewDbReport8(db *sql.DB) Db8Report8Interface {
	return &Db8Report8{Db: db}
}

func (m *Db8Report8) InsertReport(r8 model8.Report8) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	stmt, err := tx.Prepare("INSERT INTO cptm8report (nvul, ncritical, nhigh, nmedium, nlow, ninfo, ndomain, nhostname, nservice, ntech, release, link) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(r8.Nvul, r8.Ncritical, r8.Nhigh, r8.Nmedium, r8.Nlow, r8.Ninfo, r8.Ndomain, r8.Nhostname, r8.Nservice, r8.Ntech, r8.Release, r8.Link)
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return false, err2
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err2
	}
	return true, nil
}
