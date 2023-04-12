package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"itodo/app/models"
	"itodo/config"
	"net/http"
)

func session(writer http.ResponseWriter, request *http.Request) (sess models.Session, err error) {
	cookie, err := request.Cookie("chech")
	if err == nil {
		sess = models.Session{UUID: cookie.Value}
		if ok, _ := sess.CheckSession(); !ok {
			err = errors.New("Invalid session")
		}
	}
	return
}

func generateHTML(writer http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(writer, "layout", data)
}

func StartMainServer() error {
	http.HandleFunc("/", top)
	http.HandleFunc("/socialauth", auth)
	http.HandleFunc("/auth/google/callback/", callback)

	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/authenticate", authenticate)
	http.HandleFunc("/todos", index)
	http.HandleFunc("/todos/", todoId)
	http.HandleFunc("/todos/new", todoNew)
	http.HandleFunc("/todos/add", todoAdd)
	http.HandleFunc("/items/", itemEdit)

	return http.ListenAndServe(":"+config.Config.Port, nil)
}
