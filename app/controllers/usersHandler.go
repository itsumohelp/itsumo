package controllers

import (
	"encoding/json"
	"fmt"
	"itodo/app/models"
	"itodo/config"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type (
	userhandler struct {
		user models.User
	}
)

func IniHandler(u models.User) *userhandler {
	return &userhandler{u}
}

func (h userhandler) GetReq(c echo.Context) error {
	var userid int = 1
	h.user.Fetch(userid)
	if h.user.GetUuId() == "" {
		fmt.Println("404")
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.String(http.StatusOK, h.user.GetUuId())
}

func getOpenIdConnect(c echo.Context) error {
	state, err := randString(16)
	if err != nil {
		return c.String(http.StatusInternalServerError, "/")
	}
	nonce, err := randString(16)
	if err != nil {
		return c.String(http.StatusInternalServerError, "/")
	}
	writeCookie(c, "state", state)
	writeCookie(c, "nonce", nonce)
	return c.Redirect(http.StatusFound, config.Authconfig.AuthCodeURL(state, oidc.Nonce(nonce)))
}

func getGoogleCallback(c echo.Context) error {
	state, err := c.Cookie("state")
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	if c.QueryParam("state") != state.Value {
		return c.String(http.StatusInternalServerError, "")
	}
	oauth2Token, err := config.Authconfig.Exchange(config.Context, c.QueryParam("code"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return c.String(http.StatusInternalServerError, "")
	}
	idToken, err := config.Verifier.Verify(config.Context, rawIDToken)
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	nonce, err := c.Cookie("nonce")
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	if idToken.Nonce != nonce.Value {
		return c.String(http.StatusInternalServerError, "")
	}

	oauth2Token.AccessToken = "*REDACTED*"
	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	var jsondata map[string]interface{}
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	p, _ := json.Marshal(resp.IDTokenClaims)
	json.Unmarshal(p, &jsondata)

	userHandler := IniHandler(models.NewUserModel(models.InitDataBase()))
	userHandler.user.GetUserByOAUTHID(jsondata["sub"].(string), config.Constants.GoogleOpenIDConnect)
	if userHandler.user.GetUuId() == "" {
		if err := userHandler.user.CreateUser(); err != nil {
			return c.String(http.StatusInternalServerError, "")
		}
	}

	sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
	sessionHandler.session.CreateSession(userHandler.user.GetUserId())
	writeCookie(c, "chech", sessionHandler.session.GetSession().Uuid)

	// user := models.User{
	// 	OAUTHID: jsondata["sub"].(string),
	// 	VENDER:  1,
	// }
	// checkSignupUser(jsondata["sub"].(string), 1)
	// already_signup_chk, err := user.GetUserByOAUTHID(jsondata["sub"].(string))
	// if already_signup_chk.OAUTHID == "" {
	// 	if err := user.CreateUser(); err != nil {
	// 		log.Println(err)
	// 	}
	// 	registed_user, _ := user.GetUserByOAUTHID(jsondata["sub"].(string))
	// 	signupData(registed_user.ID)
	// }
	// user_registed_chk, err := user.GetUserByOAUTHID(jsondata["sub"].(string))
	// user.ID = user_registed_chk.ID
	// before_session, err := models.GetSessionByUserID(user.ID)
	// if before_session.UUID != "" {
	// 	if err := before_session.DeleteSessionByID(); err != nil {
	// 		log.Println(err)
	// 	}
	// }
	// sessions, err := user.CreateSession()
	// if err != nil {
	// 	log.Println(err)
	// }
	// writeCookie(c, "chech", sessions.UUID)
	// return c.Redirect(http.StatusFound, "/todo")
	return c.Redirect(http.StatusFound, "/todo")
}
