package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gabesullice/simplcheck/lib/checker"
)

type Reporter interface {
	Report() []string
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
		for _, report := range reports {
			fmt.Fprintf(w, "%s\n", report)
		}
	}
}
