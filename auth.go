package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"encoding/base64"
	"golang.org/x/oauth2"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		// 未認証
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// 何らかの別のエラーが発生
		panic(err.Error())
	} else {
		// 成功、ラップられたハンドラを呼ぶ
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	res, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer res.Body.Close()
	if err != nil {
		log.Printf("Getting user info failed with '%s'\n", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Parse user info failed with '%s'\n", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	log.Printf("Content: %s\n", contents)
	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: base64.StdEncoding.EncodeToString(contents),
		Path: "/",
	})
	http.Redirect(w, r, "/chat", http.StatusTemporaryRedirect)
}
