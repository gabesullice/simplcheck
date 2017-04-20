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
	state string
	times uint
	err   error
}

type Report struct {
	URL   string
	State string
	Times uint
	Err   error
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
		for url := range c.statuses {
			go c.Check(url)
		}
		time.Sleep(dur)
	}
}

func (c Checker) Report() []Report {
	var reports []Report
	for uri, status := range c.statuses {
		reports = append(
			reports,
			Report{uri, status.state, status.times, status.err},
		)
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
		new.times = old.times + 1
	} else {
		new.times = 1
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
