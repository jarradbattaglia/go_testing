package main

import (
    "net/http";
    "log";
    "text/template";
    "path/filepath";
    "sync";
    "flag";
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
    once sync.Once
    filename string
    templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r  *http.Request) { 
  t.once.Do(func() {
    t.templ =  template.Must(template.ParseFiles(filepath.Join("templates",
      t.filename)))
  })
  data := map[string]interface{}{
      "Host": r.Host,
  }
  if authCookie, err := r.Cookie("auth"); err == nil {
      data["UserData"] = objx.MustFromBase64(authCookie.Value)
  }
  t.templ.Execute(w, data) 
}

func main() {
    var addr = flag.String("addr", ":8080", "The port of the application")
    flag.Parse()
    gomniauth.SetSecurityKey("AUTHKEY")
    gomniauth.WithProviders(google.New("", "", "http://localhost:8080/auth/callback/google"))
    r := newRoom()
    http.Handle("/", MustAuth(&templateHandler{filename: "chat.html"}))
    http.Handle("/login", &templateHandler{filename: "login.html"})
    http.HandleFunc("/auth/", loginHandler)
    http.Handle("/room", r)
    go r.run()
    log.Print("Starting server on address: ", *addr)
    if err := http.ListenAndServe(*addr, nil); err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}