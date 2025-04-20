package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

type DataBase struct {
	Hash      []byte `json:"hash_refresh" db:"hash"`
	TimeAlive int64  `json:"time_alive" db:"time_alive"`
	IpAddress string `json:"ip_address" db:"ip"`
	Guid      string `json:"guid" db:"guid"`
}

func Connect() error {
	var err error
	connStr := ""
	params := map[string]map[string]string{
		"host": {
			"param":   "DBHOST",
			"default": "::1",
		},
		"port": {
			"param":   "DBPORT",
			"default": "5432",
		},
		"user": {
			"param":   "DBUSER",
			"default": "postgres",
		},
		"password": {
			"param":   "DBPASSWORD",
			"default": "postgres",
		},
		"dbname": {
			"param":   "DBNAME",
			"default": "postgres",
		},
		"sslmode": {
			"param":   "SSLMODEDB",
			"default": "disable",
		},
	}

	for paramConn, paramEnv := range params {
		param, exists := os.LookupEnv(paramEnv["param"])
		if !exists {
			param = paramEnv["default"]
		}
		connStr += fmt.Sprintf("%s=%s ", paramConn, param)
	}
	DB, err = sqlx.Connect("postgres", connStr)
	return err
}

func (db *DataBase) IsExcist() (bool, error) {
	return false, nil
}

func (db *DataBase) AddToken() error {
	_, err := DB.NamedExec(`INSERT INTO refresh_tokens (guid, hash, ip, time_alive) VALUES (:guid, :hash, :ip, :time_alive)`, db)
	return err
}
