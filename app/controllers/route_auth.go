package controllers

import (
	"log"
	"net/http"
)

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		generateHTML(w, nil, "layout", "public_navbar", "signup")
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, "/", 302)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "public_navbar", "login")
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)

	}

	http.Redirect(w, r, "/todos", 302)
}
