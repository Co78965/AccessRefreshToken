package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type DataBase struct {
	Hash      []byte `json:"hash_refresh"`
	TimeAlive int64  `json:"time_alive"`
	IpAddress string `json:"ip_address"`
	Guid      string `json:"guid"`
}

func Connect() error {
	var err error
	connStr := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable"

	params := map[string]map[string]string{
		"host": {
			"param":   "HOST",
			"default": "::1",
		},
		"port": {
			"param":   "PORT",
			"default": "5432",
		},
		"user": {
			"param":   "USER",
			"default": "postgres",
		},
		"password": {
			"param":   "PASSWORD",
			"default": "postgres",
		},
		"dbname": {
			"param":   "DBNAME",
			"default": "postgres",
		},
		"sslmode": {
			"param":   "SSLMODE",
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

	DB, err = sql.Open("postgres", connStr)
	return err
}

func (db *DataBase) IsExcist() (bool, error) {

	return false, nil
}

func (db *DataBase) AddToken() error {
	result, err := DB.Exec("insert into Products (model, company, price) values ('iPhone X', $1, $2)", "Apple", 72000)

	if err != nil {
		panic(err)
	}

	fmt.Println(result.LastInsertId()) // не поддерживается
	fmt.Println(result.RowsAffected()) // количество добавленных строк
	return nil
}
