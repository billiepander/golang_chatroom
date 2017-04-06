package main

import (
	"net/http"
	"log"
	"sync"
	"html/template"
	"path/filepath"
	"flag"
	"os"
	"github.com/chatroom/trace"
	"strings"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func loginHandler(w http.ResponseWriter, r *http.Request)  {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil{
			http.Error(w, fmt.Sprintf("Error when trying to get provider%s: %s",provider, err), http.StatusBadRequest)
			return
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)   //to start the authorization process.
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to GetBeginAuthURLfor %s:%s", provider, err), http. StatusInternalServerError)
			return
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		fmt.Println("creds:", creds)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to complete auth for%s: %s", provider, err), http.StatusInternalServerError)
			return
		}
		user, userErr := provider.GetUser(creds)
		if userErr != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get user from provider"), http.StatusInternalServerError)
			return
		}
		authCookieValue := objx.New(map[string]interface{}{   // creates a new Map containing the map[string]interface{}
			"name": user.Name(),
			"avatar_url": user.AvatarURL(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: authCookieValue,
			Path: "/"})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {          // 只有第一次conn才会渲染模板，之后直接用就是了，有点聪明啊！
		t.templ = template.Must(template.ParseFiles(filepath.Join("/home/pd/gowork/src/github.com/chatroom/chat/templates",
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
	// softcode server listener
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags

	gomniauth.SetSecurityKey("yLiCQYG7CAflDavqGH461IO0MHp7TEbpg6TwHBWdJzNwYod1i5ZTbrIF5bEoO3Pd") // NOTE: DO NOT COPY THIS - MAKE YOR OWN!
	gomniauth.WithProviders(
		github.New("a788546dede1f5506ccb", "d8bef97ce4947631d83ef3314849767348afb2bb", "http://localhost:8080/auth/callback/github"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {   // 此时设置logout，主要是为了remove cookie
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: "",
			Path: "/",
			MaxAge: -1,    // -1就是立即删除cookie
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	go r.run()

	log.Println("Starting web server on", *addr)  // 能够在terminal中打印出来，提示作用
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
