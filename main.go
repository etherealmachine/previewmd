package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"code.google.com/p/go.exp/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday"
)

var (
	mdFile = flag.String("f", "", "Markdown file to preview")
	port   = flag.Int("port", 8080, "Port to listen on")

	fileChanged chan bool
	tmpl        *template.Template
)

var tmplText = `<!doctype html>
<html>
	<head>
		<script>
			var s = new WebSocket('ws://localhost:{{ .Port }}/ws');
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
	<body>{{ .MarkdownText }}</body>
</html>
`

func init() {
	fileChanged = make(chan bool, 0)
	tmpl = template.Must(template.New("index").Parse(tmplText))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadFile(*mdFile)
	if err != nil {
		fmt.Fprintf(w, "error reading file: %v", err)
		return
	}
	mdText := string(blackfriday.MarkdownCommon(fileBytes))
	err = tmpl.Execute(w, struct {
		Port         int
		MarkdownText string
	}{
		Port:         *port,
		MarkdownText: mdText,
	})
	if err != nil {
		fmt.Fprintf(w, "error writing HTTP response: %v", err)
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := (&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}).Upgrade(w, r, nil)
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

func watchFile(f string, c chan bool) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("error watching %q: %v", f, err)
	}
	if err := w.Watch(f); err != nil {
		log.Printf("error watching %q: %v", f, err)
	}
	for _ = range w.Event {
		c <- true
	}
}

func main() {
	flag.Parse()
	if *mdFile == "" {
		fmt.Println("Markdown file must be specified (-f)")
		flag.Usage()
		os.Exit(1)
	}
	if _, err := ioutil.ReadFile(*mdFile); err != nil {
		log.Fatalf("error reading %q: %v", *mdFile, err)
	}
	go watchFile(*mdFile, fileChanged)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", websocketHandler)
	fmt.Printf("Preview file at http://localhost:%d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil))
}
