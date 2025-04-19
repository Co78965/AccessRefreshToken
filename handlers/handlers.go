package handlers

import (
	"AccessRefreshToken/token"
	"encoding/json"
	"log"
	"net/http"
)

type IToken interface {
	CheckAccessToken() bool
	GenerateAccessToken() error
	GenerateRefreshToken() error
	GetTokenAccess() string
	GetTokenRefresh() string
}

type TokensGenerator struct {
	generator IToken
}

func Init() error {
	return token.Init()
}

func newTokensGenerator(tokenGenerator IToken) *TokensGenerator {
	return &TokensGenerator{generator: tokenGenerator}
}

func GetTokens(w http.ResponseWriter, r *http.Request) {
	tokensInfo := &token.TokensInfo{}
	tokensInfo.Guid = r.URL.Query().Get("guid")
	tokensInfo.IpAddress = r.RemoteAddr

	tokenGenerator := newTokensGenerator(tokensInfo)

	type Response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := tokenGenerator.generator.GenerateAccessToken(); err != nil {
		log.Printf("[ERROR] func: GetTokens --> GenerateAccessToken | error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Somthing wrong"))
		return
	}

	if err := tokenGenerator.generator.GenerateRefreshToken(); err != nil {
		log.Printf("[ERROR] func: GetTokens --> GenerateRefreshToken | error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Somthing wrong"))
		return
	}

	resp := new(Response)
	resp.AccessToken = tokenGenerator.generator.GetTokenAccess()
	resp.RefreshToken = tokenGenerator.generator.GetTokenRefresh()

	jsonResp, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		log.Printf("[ERROR] func: GetTokens --> json.Marshal | error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Somthing wrong"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func RefreshTokens(w http.ResponseWriter, r *http.Request) {

}
