package database

import (
	"github.com/elmanelman/oracle-judge/config"
	"github.com/jmoiron/sqlx"

	_ "github.com/godror/godror"
)

const oracleDriverName = "godror"

func ConnectWithConfig(cfg config.OracleConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect(oracleDriverName, cfg.ConnectionString())
	if err != nil {
		return nil, err
	}
	return db, nil
}
