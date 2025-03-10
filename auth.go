package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"crypto/md5"
	"strings"
	"fmt"
	"io"
	"golang.org/x/oauth2"
	"github.com/stretchr/objx"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	userInfo map[string]string
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

func (u chatUser) AvatarURL() string {
	if url, ok := u.userInfo["avatar_url"]; ok {
		return url
	}
	return ""
}

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
	user := objx.MustFromJSON(string(contents))
	userInfo := map[string]string{
		"name": user.Get("name").Str(),
		"email": user.Get("email").Str(),
		"avatar_url": user.Get("picture").Str(),
	}
	chatUser := &chatUser{userInfo: userInfo}
	m := md5.New()
	io.WriteString(m, strings.ToLower(userInfo["email"]))
	chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
	avatarURL, err := avatars.GetAvatarURL(chatUser)
	if err != nil {
		log.Fatalln("GetAvatarURLに失敗しました", "-", err)
	}
	authCookieValue := objx.New(map[string]interface{}{
		"userid": chatUser.uniqueID,
		"name": userInfo["name"],
		"avatar_url": avatarURL,
	}).MustBase64()
	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: authCookieValue,
		Path: "/",
	})
	http.Redirect(w, r, "/chat", http.StatusTemporaryRedirect)
}
