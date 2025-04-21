package token

import (
	"AccessRefreshToken/database"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	ERROR        = 0
	SUCCSESS     = 1
	DIFFERENT_IP = 2
	TOKEN_DEATH  = 3
	IS_NOT_VALID = 4
)

type IValidator interface {
	IsExcist(token string) (bool, *database.DataBase, error)
}

type IDataBase interface {
	AddToken(token string) error
	DeleteToken() error
	GetIpAddress() string
	GetTimeAlive() int64
}

type DataBase struct {
	db IDataBase
}

type Validator struct {
	validator IValidator
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

func generateRefreshTokenString() (string, error) {
	lenghtStr, excist := os.LookupEnv("LENGHT_REFRESH_TOKEN")
	if !excist {
		lenghtStr = "64"
	}

	lenght, err := strconv.Atoi(lenghtStr)

	if err != nil {
		return "", err
	}

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890*&^%$/?#№")

	tokenRune := make([]rune, lenght)
	for i := range tokenRune {
		tokenRune[i] = letters[rand.Intn(len(letters))]
	}

	return string(tokenRune), nil
}

func newDbManager(database IDataBase) *DataBase {
	return &DataBase{db: database}
}

func newValidator(database IValidator) *Validator {
	return &Validator{validator: database}
}

func (t *TokensInfo) GetIpAddress() string {
	return t.IpAddress
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
	fmt.Println(getTimeAlive("refresh"), time.Now().Add(getTimeAlive("refresh")).Unix(), time.Now().Unix())

	dbInfo.TimeAlive = time.Now().Add(getTimeAlive("refresh")).Unix()

	tokenString, err := generateRefreshTokenString()
	if err != nil {
		log.Printf("[ERROR] func: GenerateRefreshToken() --> generateRefreshToken() | error: %v\n", err)
		return err
	}

	t.TokenRefresh = base64.StdEncoding.EncodeToString([]byte(tokenString))

	dbManager := newDbManager(dbInfo)

	if err := dbManager.db.AddToken(tokenString); err != nil {
		log.Printf("[ERROR] func: GenerateRefreshToken() --> dbManager.db.AddToken() | error: %v\n", err)
		return err
	}

	return nil
}

func checkTimeAlive(timeAlive int64) bool {
	fmt.Println(timeAlive, time.Now().Unix())
	return timeAlive >= time.Now().Unix()
}

func compareIp(IpCorrect string, IpTest string) bool {
	return IpCorrect == IpTest
}

func (t *TokensInfo) IsValidRefreshToken() (int, error) {
	dbInfoValidator := new(database.DataBase)

	tokenRefreshDecode, err := base64.StdEncoding.DecodeString(t.TokenRefresh)
	if err != nil {
		return ERROR, err
	}

	dbInfoValidator.Guid = t.Guid
	dbValidator := newValidator(dbInfoValidator)

	isValid, dbInfo, err := dbValidator.validator.IsExcist(string(tokenRefreshDecode))

	if err != nil {
		log.Println("[WARNING] token isn't excist (*-*)")
		return ERROR, err
	}

	if !isValid {
		log.Println("[WARNING] token isn't valid (*-*)")
		return IS_NOT_VALID, err
	}

	dbManager := newDbManager(dbInfo)

	if isCmp := compareIp(dbManager.db.GetIpAddress(), t.IpAddress); !isCmp {
		log.Println("[WARNING] ip addresses are not similar (>o<)")
		return DIFFERENT_IP, nil //Отправка сообщения
	}

	if isNotDeath := checkTimeAlive(dbManager.db.GetTimeAlive()); !isNotDeath {
		log.Println("[WARNING] refresh token is death (x_x)")
		return TOKEN_DEATH, nil //Отправка на повторный вход в профиль
	}

	err = dbManager.db.DeleteToken()
	if err != nil {
		return ERROR, err
	}

	return SUCCSESS, nil
}
