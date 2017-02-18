package main

import (
	"log"
	"net/http"
	"text/template"
	"path/filepath"
	"sync"
	"flag"
	"os"
	"encoding/base64"
	"encoding/json"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/umatoma/chat/trace"
)

var (
	googleOauthConfig = new(oauth2.Config)
	oauthStateString = "random"
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func init() {
	err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

	googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:8080/auth/callback/google",
		ClientID:			os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:	os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:				[]string{
										"https://www.googleapis.com/auth/userinfo.profile",
										"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint:			google.Endpoint,
	}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	t.once.Do(func() {
		filenames := filepath.Join("templates", t.filename)
		t.templ = template.Must(template.ParseFiles(filenames))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		if decoded, err := base64.StdEncoding.DecodeString(authCookie.Value); err == nil {
			var userData map[string]interface{}
			json.Unmarshal(decoded, &userData)
			data["UserData"] = userData
		}
	}
	t.templ.Execute(w, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: "",
		Path: "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func main()  {
	var addr = flag.String("addr", ":8080", "アプリケーションのアドレス")
	flag.Parse()

	r := newRoom(UseGravatar)
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/auth/login/google", loginHandler)
	http.HandleFunc("/auth/callback/google", callbackHandler)
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	http.HandleFunc("/uploader", uploaderHandler)
	http.Handle("/room", r)

	go r.run()
	// Webサーバーを起動
	log.Println("Webサーバーを開始します。ポート:", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndService:", err)
	}
}
