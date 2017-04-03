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
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {          // 只有第一次conn才会渲染模板，之后直接用就是了，有点聪明啊！
		t.templ = template.Must(template.ParseFiles(filepath.Join("/home/pd/gowork/src/github.com/chatroom/chat/templates",
			t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	// softcode server listener
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)  // 能够在terminal中打印出来，提示作用
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
