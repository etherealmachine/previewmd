package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"code.google.com/p/go.exp/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday"
)

var mdFile = flag.String("f", "", "Markdown file to preview")

var tmpl = `<!doctype html>
<html>
	<head>
		<script>
			var s = new WebSocket('ws://localhost:8080/ws');
			s.onopen = function(e) {
				console.log(e)
			};
			s.onclose = function() {
				console.log('closed');
			};
			s.onerror = function(e) {
				console.log(e);
			};
			s.onmessage = function(e) {
				window.location = window.location;
			};
		</script>
	</head>
	<body>{{ . }}</body>
</html>
`

var fileChanged chan bool

func init() {
	fileChanged = make(chan bool, 0)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		for {
			if _, _, err := c.NextReader(); err != nil {
				c.Close()
				break
			}
		}
	}()
	for _ = range fileChanged {
		c.WriteMessage(websocket.TextMessage, []byte("file changed"))
	}
}

func main() {
	flag.Parse()
	t := template.Must(template.New("index").Parse(tmpl))
	if *mdFile == "" {
		log.Fatal("Markdown file must be specified (--f)")
	}
	if _, err := ioutil.ReadFile(*mdFile); err != nil {
		log.Fatalf("error reading %q: %v", *mdFile, err)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("error watching %q: %v", *mdFile, err)
	}
	if err := w.Watch(*mdFile); err != nil {
		log.Printf("error watching %q: %v", *mdFile, err)
	}
	go func() {
		for _ = range w.Event {
			fileChanged <- true
		}
	}()
	http.HandleFunc("/ws", websocketHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if mdBytes, err := ioutil.ReadFile(*mdFile); err != nil {
			fmt.Fprintf(w, "error reading file: %v", err)
		} else if err := t.Execute(w, string(blackfriday.MarkdownCommon(mdBytes))); err != nil {
			fmt.Fprintf(w, "error writing HTTP response: %v", err)
		}
	})
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
