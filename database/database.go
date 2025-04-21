package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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

func generateRefreshTokenHash(token string) ([]byte, error) {
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, err
	}
	return tokenHash, err
}

func (db *DataBase) GetIpAddress() string {
	return db.IpAddress
}

func (db *DataBase) GetTimeAlive() int64 {
	return db.TimeAlive
}

func (db *DataBase) IsExcist(token string) (bool, *DataBase, error) {
	var tokensInfo []DataBase
	var err error

	err = DB.Select(&tokensInfo, "SELECT hash, guid, time_alive, ip FROM refresh_tokens WHERE guid=$1", db.Guid)
	if err != nil {
		return false, nil, err
	}

	for _, tokenInfo := range tokensInfo {
		err = bcrypt.CompareHashAndPassword(tokenInfo.Hash, []byte(token))
		if err == nil {
			return true, &tokenInfo, nil
		}
	}

	return false, nil, nil
}

func (db *DataBase) DeleteToken() error {
	_, err := DB.NamedExec(`DELETE FROM refresh_tokens WHERE hash=:hash`, db)
	return err
}

func (db *DataBase) AddToken(token string) error {
	var err error

	db.Hash, err = generateRefreshTokenHash(token)
	if err != nil {
		return err
	}
	_, err = DB.NamedExec(`INSERT INTO refresh_tokens (guid, hash, ip, time_alive) VALUES (:guid, :hash, :ip, :time_alive)`, db)
	return err
}
