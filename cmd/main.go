package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gabesullice/simplcheck/lib/checker"
)

const (
	page   = `<!DOCTYPE html><html><body><ol>{{ range . }}<li>{{ template "status" . }}</li>{{ end }}</ol></body></html>`
	status = `<span style="color: {{ if eq .State "passing" }}green{{ else if eq .State "unknown" }}orange{{ else }}red{{ end }};">{{ .URL }} has been {{ .State }} for the past {{ .Times }} checks`
)

type Reporter interface {
	Report() []checker.Report
}

type ConfigFileLoader interface {
	LoadConfigFile(*os.File) error
}

type Runner interface {
	Run()
}

func main() {
	addr := flag.String("address", ":80", "The network address on which to listen.")
	conf := flag.String("conf", "-", "The network address on which to listen.")

	flag.Parse()

	StartServer(addr, conf)
}

func StartServer(addr *string, conf *string) {
	log.Printf("Starting simplcheck\n")

	checker := checker.NewChecker(checker.DefaultClient)

	ConfigureChecker(checker, *conf)
	StartChecker(checker)

	http.HandleFunc("/", StatusPage(checker))
	log.Printf("Attempting to listen on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func StartChecker(runner Runner) {
	go func() {
		defer panic("The checker has crashed!")
		runner.Run()
	}()
	log.Println("Started checker.")
}

func ConfigureChecker(checker ConfigFileLoader, conf string) {
	var file *os.File
	var err error
	if conf == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(conf)
		if err != nil {
			log.Fatalln("Unable to load configuration file", err)
		}
	}
	if err := checker.LoadConfigFile(file); err != nil {
		log.Fatalln("Unable to parse configuration", err)
	}
}

func StatusPage(reporter Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		reports := reporter.Report()
		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.New("page").Parse(page))
		tmpl = template.Must(tmpl.New("status").Parse(status))
		ensure(tmpl.ExecuteTemplate(w, "page", reports))
	}
}

func ensure(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
