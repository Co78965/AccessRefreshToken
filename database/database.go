package database

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

type DataBase struct {
	Hash      []byte `json:"hash_refresh"`
	TimeAlive int64  `json:"time_alive"`
	IpAddress string `json:"ip_address"`
	Guid      string `json:"guid"`
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
		log.Println("[INFO] paramConn and paramEnv and paramEnv['param']: ", paramConn, paramEnv, paramEnv["param"])
		param, exists := os.LookupEnv(paramEnv["param"])
		if !exists {
			param = paramEnv["default"]
		}
		connStr += fmt.Sprintf("%s=%s ", paramConn, param)
	}
	log.Println("[INFO] db connect: ", connStr)
	DB, err = sqlx.Connect("postgres", connStr)
	return err
}

func (db *DataBase) IsExcist() (bool, error) {

	return false, nil
}

func (db *DataBase) AddToken() error {

	return nil
}
