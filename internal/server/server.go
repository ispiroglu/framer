package server

import (
	"embed"
	"fmt"
	"frame/internal/frame"
	"html/template"
	"net/http"
	"strconv"
)

//go:embed templates
var templates embed.FS

var currentCount int
var totalCount int

type dataS struct {
	CurrentCount, TotalCount int
}

func Start() {

	fmt.Printf("templates: %v\n", templates)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /process", dimensionsHandler)
	mux.HandleFunc("GET /process", processGetHandler)
	mux.HandleFunc("GET /get-ok-count", getOkCountHandler)
	mux.HandleFunc("/", pageHandler)
	http.ListenAndServe(":8080", mux)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		panic(err)
	}

	t.Execute(w, nil)
}

func processGetHandler(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFS(templates, "templates/process.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, nil)
}

func getOkCountHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
        <div class="message">
            %d / %d fotoğraf işlendi.
            <p>İyi sergiler YTÜFOK</p>
        </div>
    `, currentCount, totalCount)
}

func dimensionsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	h, _ := strconv.Atoi(r.FormValue("height"))
	wid, _ := strconv.Atoi(r.FormValue("width"))
	p, _ := strconv.Atoi(r.FormValue("padding"))

	fmt.Printf("\"s\": %v\n", "s")

	tChan := make(chan int)
	okChan := make(chan struct{})

	go frame.Process(h, wid, p, tChan, okChan)
	go func() {
		for {
			select {
			case <-okChan:
				currentCount++
			case t := <-tChan:
				totalCount = t
			}
		}
	}()

	tmpl, err := template.ParseFS(templates, "templates/process.html")
	if err != nil {
		panic(err)
	}

	data := dataS{
		CurrentCount: currentCount,
		TotalCount:   totalCount,
	}

	tmpl.Execute(w, data)
}
