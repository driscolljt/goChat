package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"trace"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

var host = flag.String("host", ":8080", "The host of the application.")

func main() {
	flag.Parse() // parse the flags

	// setup gomniauth
	gomniauth.SetSecurityKey("NkaiAb60RlVeoPGfMztQsuPyZigOvwYizFfIzIsI9ZLOMyVWhTzOp99gP4jzDwSY")
	gomniauth.WithProviders(
		facebook.New("192781844767181", "40d969731e2cc15c1160eee627f80ba7", "http://localhost:8080/auth/callback/facebook"),
		github.New("e2848b5671d849a66940", "e9d4935f3bbfabc1f74afe96c1c60140a8298e06", "http://localhost:8080/auth/callback/github"),
		google.New("161248265003-sngv184c6o5sep8lg6cnb21t7sp801gg.apps.googleusercontent.com", "sektL0KKCRjHnaNzttNL4L2b", "http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	// root
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start web server
	log.Println("Starting web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
