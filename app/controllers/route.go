package controllers

import (
	"encoding/json"
	"errors"
	"itodo/app/models"
	"itodo/config"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

func Requestroute() {
	e := echo.New()
	e.Static("/static", "static")
	e.File("/", "static/top.html")

	e.GET("/get", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/socialauth", func(c echo.Context) error {
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
	})
	e.GET("/auth/google/callback/", func(c echo.Context) error {
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

		user := models.User{
			OAUTHID: jsondata["sub"].(string),
			VENDER:  1,
		}
		already_signup_chk, err := user.GetUserByOAUTHID(jsondata["sub"].(string))
		if already_signup_chk.OAUTHID == "" {
			if err := user.CreateUser(); err != nil {
				log.Println(err)
			}
			registed_user, _ := user.GetUserByOAUTHID(jsondata["sub"].(string))
			signupData(registed_user.ID)
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
		writeCookie(c, "chech", sessions.UUID)
		return c.Redirect(http.StatusFound, "/todo")
	})

	e.GET("/todo", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			_, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			return c.File("static/todos.html")
		}
	})

	e.GET("/todos", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			user, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			todos, _ := user.GetTodosByUser()
			return json.NewEncoder(c.Response()).Encode(todos)
		}
	})

	e.GET("/todos/:todoid", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			_, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			return c.File("static/item.html")
		}
	})

	e.GET("/todos/:todoid/", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			todoid, _ := strconv.Atoi(c.Param("todoid"))
			user, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			todo, _ := models.GetTodos(todoid, user.ID)
			items, _ := models.GetItems(todo.ID)
			return json.NewEncoder(c.Response()).Encode(items)
		}
	})

	// e.POST("/todos", func(c echo.Context) error {
	// 	sess, err := checksession(c)
	// 	if err != nil {
	// 		return c.Redirect(http.StatusFound, "/")
	// 	} else {
	// 		content := r.PostFormValue("content")
	// 		if err := models.CreateItem(content, todoid); err != nil {
	// 			log.Println(err)
	// 		}
	// 		todoid, _ := strconv.Atoi(c.Param("todoid"))
	// 		user, err := sess.GetUserBySession()
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		todo, _ := models.GetTodo(todoid, user.ID)
	// 		elements, _ := models.GetElements(todo.ID)
	// 		return c.File("static/todos.html")
	// 	}
	// })

	e.GET("/elements/:todoid", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			todoid, _ := strconv.Atoi(c.Param("todoid"))
			user, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			todo, _ := models.GetTodos(todoid, user.ID)
			elements, _ := models.GetElements(todo.ID)
			return json.NewEncoder(c.Response()).Encode(elements)
		}
	})

	e.POST("/elements/:todoid", func(c echo.Context) error {
		sess, err := checksession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			todoid, _ := strconv.Atoi(c.Param("todoid"))
			user, err := sess.GetUserBySession()
			if err != nil {
				log.Println(err)
			}
			todo, _ := models.GetTodos(todoid, user.ID)

			u := new([]models.Postele)
			if err := c.Bind(u); err != nil {
				return err
			}
			bytes, err := json.Marshal(u)
			models.UpdateElements(string(bytes), todo.ID)
			return json.NewEncoder(c.Response()).Encode(u)
		}
	})

	e.Logger.Fatal(e.Start(":" + config.Port))
}

func signupData(user_id int) error {
	todo := new(models.Todo)
	todo.Content = "いつもヘルプへようこそ！"
	todo.UserID = user_id
	models.AddTodo(user_id, todo.Content)
	gettodo, _ := models.GetTodo(todo.UserID)

	element := new(models.Element)
	element.Content = "[{\"Value\":\"こんにちわ！\",\"Check\":0},{\"Value\":\"いつもをぜひ使ってください\",\"Check\":1}]"
	element.TodoID = gettodo.ID
	models.CreateElement(element.Content, element.TodoID)
	return nil
}

func writeCookie(c echo.Context, name, value string) error {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
	return nil
}

func checksession(c echo.Context) (sess models.Session, err error) {
	cookie, err := c.Cookie("chech")
	if err == nil {
		sess = models.Session{UUID: cookie.Value}
		if ok, _ := sess.CheckSession(); !ok {
			err = errors.New("Invalid session")
		}
	}
	return
}
