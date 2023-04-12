package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"itodo/app/models"
	"itodo/config"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

func top(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "public_navbar", "top")
}

func auth(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)
	fmt.Println(config.Authconfig)
	http.Redirect(w, r, config.Authconfig.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func callback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}
	oauth2Token, err := config.Authconfig.Exchange(config.Context, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	idToken, err := config.Verifier.Verify(config.Context, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	nonce, err := r.Cookie("nonce")
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce did not match", http.StatusBadRequest)
		return
	}

	oauth2Token.AccessToken = "*REDACTED*"
	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var jsondata map[string]interface{}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p, _ := json.Marshal(resp.IDTokenClaims)
	json.Unmarshal(p, &jsondata)

	user := models.User{
		OAUTHID: jsondata["sub"].(string),
		VENDER:  1,
	}
	already_signup_chk, err := user.GetUserByOAUTHID(jsondata["sub"].(string))
	if already_signup_chk.OAUTHID == "" {
		if err := user.CreateUser(); err != nil {
			log.Println(err)
		}
	}
	user_registed_chk, err := user.GetUserByOAUTHID(jsondata["sub"].(string))
	user.ID = user_registed_chk.ID
	before_session, err := models.GetSessionByUserID(user.ID)
	if before_session.UUID != "" {
		if err := before_session.DeleteSessionByID(); err != nil {
			log.Println(err)
		}
	}
	sessions, err := user.CreateSession()
	if err != nil {
		log.Println(err)
	}
	setCallbackCookie(w, r, "chech", sessions.UUID)
	http.Redirect(w, r, "/todos", http.StatusFound)
}

func index(w http.ResponseWriter, r *http.Request) {
	sess, err := session(w, r)
	if err != nil {
		http.Redirect(w, r, "/", 302)
	} else {
		user, err := sess.GetUserBySession()
		if err != nil {
			log.Println(err)
		}
		todos, _ := user.GetTodosByUser()
		user.Todos = todos
		generateHTML(w, user, "layout", "private_navbar", "index")
	}
}

func todoId(w http.ResponseWriter, r *http.Request) {
	sess, err := session(w, r)
	if err != nil {
		http.Redirect(w, r, "/", 302)
	} else {
		user, err := sess.GetUserBySession()
		if err != nil {
			http.Redirect(w, r, "/", 302)
		}
		if strings.HasSuffix(r.URL.String(), "/add") {
			todoid, _ := strconv.Atoi(r.URL.String()[7:strings.Index(r.URL.String(), "/add")])
			content := r.PostFormValue("content")
			if err := models.CreateItem(content, todoid); err != nil {
				log.Println(err)
			}
			http.Redirect(w, r, "/todos/"+r.URL.String()[7:strings.Index(r.URL.String(), "/add")], 302)
		} else if strings.HasSuffix(r.URL.String(), "/edit") {
			fmt.Println("edit")
		} else if strings.HasSuffix(r.URL.String(), "/del") {
			todoid, _ := strconv.Atoi(r.URL.String()[7:strings.Index(r.URL.String(), "/del")])
			if err := models.DeleteTodo(todoid); err != nil {
				log.Println(err)
			}
			http.Redirect(w, r, "/todos", 302)
		} else if strings.Count(r.URL.String(), "/") == 2 {
			todoid, _ := strconv.Atoi(r.URL.String()[7:])
			todo, _ := models.GetTodos(todoid, user.ID)
			items, _ := models.GetItems(todoid)
			todo.Items = items
			generateHTML(w, todo, "layout", "private_navbar", "item")
		} else {
			http.Redirect(w, r, "/todos", 302)
		}

	}
}

func todoNew(w http.ResponseWriter, r *http.Request) {
	_, err := session(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", 302)
	} else {
		generateHTML(w, nil, "layout", "private_navbar", "todo_new")
	}
}

func todoAdd(w http.ResponseWriter, r *http.Request) {
	sess, err := session(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", 302)
	} else {
		err = r.ParseForm()
		if err != nil {
			log.Println(err)

		}
		user, err := sess.GetUserBySession()
		if err != nil {
			log.Println(err)
		}
		content := r.PostFormValue("content")
		if err := user.CreateTodo(content); err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/todos", 302)
	}
}

func itemEdit(w http.ResponseWriter, r *http.Request) {
	sess, err := session(w, r)
	if err != nil {
		http.Redirect(w, r, "/", 302)
	} else {
		_, err := sess.GetUserBySession()
		if err != nil {
			http.Redirect(w, r, "/", 302)
		} else if strings.HasSuffix(r.URL.String(), "/edit") {
			itemid, _ := strconv.Atoi(strings.Replace(r.URL.String()[7:], "/edit", "", -1))
			priority, _ := strconv.Atoi(r.PostFormValue("priority"))
			todoid, _ := strconv.Atoi(r.PostFormValue("todoid"))
			models.UpdateItemPriority(priority, itemid, todoid)
			http.Redirect(w, r, "/todos/"+r.PostFormValue("todoid"), 302)
		} else if strings.HasSuffix(r.URL.String(), "/del") {
			itemid, _ := strconv.Atoi(strings.Replace(r.URL.String()[7:], "/del", "", -1))
			todoid, _ := strconv.Atoi(r.PostFormValue("todoid"))
			models.DeleteItem(itemid, todoid)
			http.Redirect(w, r, "/todos/"+r.PostFormValue("todoid"), 302)
		} else {
			http.Redirect(w, r, "/todos", 302)
		}

	}
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, c)
}
