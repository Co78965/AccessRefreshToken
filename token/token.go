package token

import (
	"AccessRefreshToken/database"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type IDataBase interface {
	IsExcist() (bool, error)
	AddToken() error
}

type DataBase struct {
	db IDataBase
}

type TokensInfo struct {
	Guid         string `json:"guid"`
	IpAddress    string `json:"ip_address"`
	TokenAccess  string `json:"token_acces"`
	TokenRefresh string `json:"token_refresh"`
}

var mySigningKey = []byte("secret")

func Init() error {
	return database.Connect()
}

func getTimeAlive(typeToken string) time.Duration {
	paramToken := map[string]map[string]string{
		"access": {
			"param":            "TIMEALIVE_ACCESS_TOKEN",
			"default":          "10",
			"duration":         "TYPE_DURATION_ACCESS",
			"default_duration": "minute",
		},
		"refresh": {
			"param":            "TIMEALIVE_REFRESH_TOKEN",
			"default":          "10",
			"duration":         "TYPE_REFRESH_ACCESS",
			"default_duration": "hour",
		},
	}

	timeAliveStr, excist := os.LookupEnv(paramToken[typeToken]["param"])
	if !excist {
		timeAliveStr = paramToken[typeToken]["default"]
	}

	timeAlive, err := strconv.Atoi(timeAliveStr)
	if err != nil {
		log.Printf("[ERROR] func: getTimeAlive() --> strconv.Atoi() | error: %v\n", err)
		return 100000
	}

	durationsType := map[string]time.Duration{
		"hour":   time.Hour,
		"second": time.Second,
		"minute": time.Minute,
	}

	durationType, excist := os.LookupEnv(paramToken[typeToken]["duration"])
	if !excist {
		durationType = paramToken[typeToken]["default_duration"]
	}

	return time.Duration(timeAlive) * durationsType[durationType]
}

func generateRefreshToken() (string, error) {
	lenghtStr, excist := os.LookupEnv("LENGHT_REFRESH_TOKEN")
	if !excist {
		lenghtStr = "64"
	}

	lenght, err := strconv.Atoi(lenghtStr)

	if err != nil {
		return "", err
	}

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890*&^%$/?#№")

	b := make([]rune, lenght)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b), nil
}

func newDbManager(database IDataBase) *DataBase {
	return &DataBase{db: database}
}

func (t *TokensInfo) GetTokenAccess() string {
	return t.TokenAccess
}

func (t *TokensInfo) GetTokenRefresh() string {
	return t.TokenRefresh
}

func (t *TokensInfo) CheckAccessToken() bool {
	token, err := jwt.ParseWithClaims(t.TokenAccess, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secret_key"), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}

func (t *TokensInfo) GenerateAccessToken() error {
	token := jwt.New(jwt.SigningMethodHS512)

	claims := token.Claims.(jwt.MapClaims)

	claims["ip"] = t.IpAddress
	claims["id"] = t.Guid

	claims["exp"] = time.Now().Add(getTimeAlive("access")).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return err
	}

	t.TokenAccess = tokenString
	return nil
}

func (t *TokensInfo) GenerateRefreshToken() error {
	var err error

	dbInfo := new(database.DataBase)
	dbInfo.IpAddress = t.IpAddress
	dbInfo.Guid = t.Guid
	dbInfo.TimeAlive = time.Now().Add(getTimeAlive("refresh")).Unix()

	t.TokenRefresh, err = generateRefreshToken()

	fmt.Println("[INFO] token refresh: ", t.TokenRefresh)

	if err != nil {
		log.Printf("[ERROR] func: GenerateRefreshToken() --> generateRefreshToken() | error: %v\n", err)
		return err
	}

	dbInfo.Hash, err = bcrypt.GenerateFromPassword([]byte(t.TokenRefresh), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] func: GenerateRefreshToken() --> bcrypt.GenerateFromPassword() | error: %v\n", err)
		return err
	}

	dbManager := newDbManager(dbInfo)

	if err := dbManager.db.AddToken(); err != nil {
		log.Printf("[ERROR] func: GenerateRefreshToken() --> dbManager.db.AddToken() | error: %v\n", err)
		return err
	}

	return nil
}
