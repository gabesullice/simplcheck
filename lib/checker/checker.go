package checker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Status struct {
	state  string
	checks uint
	err    error
}

type Checker struct {
	statuses map[string]Status
	client   Getter
	config   Config
}

type Config struct {
	Settings     Settings
	Applications []string
}

type Settings struct {
	Interval string
}

type Getter interface {
	Get(string) (resp *http.Response, err error)
}

func NewChecker(opts ...func(c *Checker)) *Checker {
	checker := &Checker{client: &http.Client{}}
	for _, opt := range opts {
		opt(checker)
	}
	return checker
}

func DefaultClient(c *Checker) {
	c.client = &http.Client{}
}

func UseClient(g Getter) func(c *Checker) {
	return func(c *Checker) {
		c.client = g
	}
}

func (c *Checker) Run() {
	for {
		dur, err := time.ParseDuration(c.config.Settings.Interval)
		if err != nil {
			log.Fatalln("Unable to parse duration from interval configuration.", err)
		}
		time.Sleep(dur)
		for url := range c.statuses {
			c.Check(url)
		}
	}
}

func (c Checker) Report() []string {
	var reports []string
	for uri, status := range c.statuses {
		tmpl := "%s: %s for the past %d checks"
		reports = append(reports, fmt.Sprintf(tmpl, uri, status.state, status.checks))
	}
	return reports
}

func (c *Checker) Check(url string) (new Status, err error) {
	old, ok := c.statuses[url]
	if !ok {
		return new, fmt.Errorf("Cannot check %s. No associated configuration.", url)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		new.err = err
		log.Println(err)
		return new, err
	}

	if resp.StatusCode != http.StatusOK || err != nil {
		new.state = "failing"
	} else {
		new.state = "passing"
	}

	if old.state == new.state {
		new.checks = old.checks + 1
	} else {
		new.checks = 1
	}

	c.statuses[url] = new

	return new, nil
}

func (c *Checker) LoadConfigFile(file *os.File) error {
	var conf Config
	dec := json.NewDecoder(file)
	if err := dec.Decode(&conf); err != nil {
		return err
	}
	return c.LoadConfig(conf)
}

func (c *Checker) LoadConfig(conf Config) error {
	c.config = conf
	c.initializeStatuses()
	return nil
}

func (c *Checker) initializeStatuses() {
	statuses := map[string]Status{}
	for _, url := range c.config.Applications {
		statuses[url] = Status{state: "unknown"}
	}
	c.statuses = statuses
}
