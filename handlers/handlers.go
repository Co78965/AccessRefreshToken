package handlers

import (
	"AccessRefreshToken/notify"
	"AccessRefreshToken/token"
	"encoding/json"
	"log"
	"net/http"
)

type INotify interface {
	Notify(guid string, ip string) error
}

type ITokenRefresh interface {
	IsValidRefreshToken() (int, error)
	GetIpAddress() string
	GetGuid() string
}

type IToken interface {
	CheckAccessToken() bool
	GenerateAccessToken() error
	GenerateRefreshToken() error
	GetTokenAccess() string
	GetTokenRefresh() string
}

type Notifyer struct {
	notifyer INotify
}

type TokenManager struct {
	manager ITokenRefresh
}

type TokensGenerator struct {
	generator IToken
}

func Init() error {
	return token.Init()
}

func newTokensManager(tokenGenerator ITokenRefresh) *TokenManager {
	return &TokenManager{manager: tokenGenerator}
}

func newTokensGenerator(tokenGenerator IToken) *TokensGenerator {
	return &TokensGenerator{generator: tokenGenerator}
}

func newNotifyer(notifyer INotify) *Notifyer {
	return &Notifyer{notifyer: notifyer}
}

func generateTokens(tokensInfo IToken) ([]byte, error) {
	type Response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	tokenGenerator := newTokensGenerator(tokensInfo)

	if err := tokenGenerator.generator.GenerateAccessToken(); err != nil {
		log.Printf("[ERROR] func: GetTokens --> GenerateAccessToken | error: %v\n", err)
		return nil, err
	}

	if err := tokenGenerator.generator.GenerateRefreshToken(); err != nil {
		log.Printf("[ERROR] func: GetTokens --> GenerateRefreshToken | error: %v\n", err)
		return nil, err
	}

	resp := new(Response)
	resp.AccessToken = tokenGenerator.generator.GetTokenAccess()
	resp.RefreshToken = tokenGenerator.generator.GetTokenRefresh()

	jsonResp, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		log.Printf("[ERROR] func: GetTokens --> json.Marshal | error: %v\n", err)
		return nil, err
	}

	return jsonResp, nil
}

func GetTokens(w http.ResponseWriter, r *http.Request) {
	tokensInfo := &token.TokensInfo{}
	tokensInfo.Guid = r.URL.Query().Get("guid")
	tokensInfo.IpAddress = r.RemoteAddr

	jsonResp, err := generateTokens(tokensInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Somthing wrong"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func RefreshTokens(w http.ResponseWriter, r *http.Request) {
	notifyerEmail := &notify.Notifyer{}
	tokensInfo := &token.TokensInfo{}

	err := json.NewDecoder(r.Body).Decode(&tokensInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Somthing wrong"))
		return
	}
	tokensInfo.IpAddress = r.RemoteAddr
	tokenManager := newTokensManager(tokensInfo)

	isValid, err := tokenManager.manager.IsValidRefreshToken()

	var code int
	var response []byte

	switch isValid {
	case token.DIFFERENT_IP:
		notifyer := newNotifyer(notifyerEmail)
		err := notifyer.notifyer.Notify(tokenManager.manager.GetGuid(), r.RemoteAddr)

		if err != nil {
			log.Printf("[ERROR] func: RefreshTokens --> Notify | error: %v\n", err)
		}

		code = http.StatusUnauthorized
		response = []byte("ip addresses are not similar")
	case token.ERROR:
		log.Printf("[ERROR] func: RefreshTokens --> IsValidRefreshToken | error: %v\n", err)
		code = http.StatusUnauthorized
		response = []byte("Somthing wrong")
	case (token.TOKEN_DEATH):
		code = http.StatusUnauthorized
		response = []byte("user is unauthorized")
	case token.IS_NOT_VALID:
		code = http.StatusUnauthorized
		response = []byte("token is not valid")
	case token.SUCCSESS:
		response, err = generateTokens(tokensInfo)
		if err != nil {
			code = http.StatusInternalServerError
			response = []byte("Somthing wrong")
			break
		}
		code = http.StatusOK
	default:
		log.Printf("[ERROR] func: RefreshTokens  | error: incorrect value switch")
		code = http.StatusInternalServerError
		response = []byte("Somthing wrong")
	}

	w.WriteHeader(code)
	w.Write(response)
}
