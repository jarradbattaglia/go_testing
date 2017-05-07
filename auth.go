package main

import "net/http"
import "strings"
import "fmt"
import "github.com/stretchr/gomniauth"
import "github.com/stretchr/objx"

type authHandler struct {
    next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    _, err := r.Cookie("auth")
    if err == http.ErrNoCookie {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusTemporaryRedirect)
        return
    }
    h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
    return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    segs := strings.Split(r.URL.Path, "/")
    action := segs[2]
    provider := segs[3]
    switch action {
        case "login":
        provider, err := gomniauth.Provider(provider)
        if err != nil {
            http.Error(w, fmt.Sprintf("error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
            return
        }
        loginURL, err := provider.GetBeginAuthURL(nil, nil)
        if err != nil {
            http.Error(w, fmt.Sprintf("error when trying to GetBeginAuthURL for %s: %s", provider, err), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", loginURL)
        w.WriteHeader(http.StatusTemporaryRedirect)
        case "callback":
        provider, err := gomniauth.Provider(provider)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
            return
        }
        creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
        if err != nil {
            http.Error(w, fmt.Sprintf("Unable to get objx for provider %s: %s", provider, err), http.StatusInternalServerError)
            return
        }
        user, err := provider.GetUser(creds)
        if err != nil {
            http.Error(w, fmt.Sprintf("UNable to get user from provider %s: %s", provider, err), http.StatusInternalServerError)
            return
        }
        authCookieValue := objx.New(map[string]interface{}{
            "name": user.Name(),
        }).MustBase64()
        http.SetCookie(w, &http.Cookie{
            Name: "auth",
            Value: authCookieValue,
            Path: "/",
        })
        w.Header().Set("Location", "/chat")
        w.WriteHeader(http.StatusTemporaryRedirect)
        default:
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "Auth action %s is not supported", action)
    }

}