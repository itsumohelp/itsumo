package controllers

import (
	"itodo/app/models"
)

type (
	sessionHandler struct {
		session models.Session
	}
)

func IniSessionHandler(u models.Session) *sessionHandler {
	return &sessionHandler{u}
}

func (h sessionHandler) CreateSession(userid int) {
	h.session.CreateSession(userid)
}

func (h sessionHandler) Checksession(cookieValue string) bool {
	return h.session.CheckSession(cookieValue)
}
