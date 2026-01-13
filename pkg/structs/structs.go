package structs

import (
	"time"
)

type YAMLFields struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type YAMLBruteforce struct {
	From      string   `yaml:"from"`
	File      string   `yaml:"file"`
	List      []string `yaml:"list"`
	Field     string   `yaml:"field"`
	Threads   int      `yaml:"threads"`
	NoVerbose bool     `yaml:"no_verbose"`
	Output    string   `yaml:"output"`
	Debug     bool     `yaml:"debug"`
	Plugin 	  any 	   `yaml:"plugin"`
	Throttle  time.Duration `yaml:"throttle" default:"0s"`
}

type YAMLSite struct {
	Host      string `yaml:"host"`
	Method    string `yaml:"method"`
	IgnoreTLS bool   `yaml:"ignore_tls"`
}

type YAMLCrawl struct {
	Name   string `yaml:"name"`
	Url    string `yaml:"url"`
	Search string `yaml:"search"`
}

type YAMLConfig struct {
	Import  any			   `yaml:"import"`
	F       []YAMLFields   `yaml:"fields"`
	B       YAMLBruteforce `yaml:"bruteforce"`
	OF      YAMLOn_fail    `yaml:"on_fail"`
	OP      YAMLOn_pass    `yaml:"on_pass"`
	H       []YAMLHeaders  `yaml:"headers"`
	S       YAMLSite       `yaml:"site"`
	C       YAMLCrawl      `yaml:"crawl"`
	P       YAMLProxy      `yaml:"proxy"`
	E       YAMLEmail      `yaml:"email"`
}

type YAMLProxy struct {
	Addr string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
	VerifyUrl string `yaml:"verify_url"`
}

type YAMLOn_fail struct {
	Message    string `yaml:"message"`
	StatusCode int    `yaml:"status_code"`
}

type YAMLOn_pass struct {
	Message    string `yaml:"message"`
	StatusCode int    `yaml:"status_code"`
}

type YAMLHeaders struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type YAMLEmail struct {
	Server YAMLEmailServer `yaml:"server"`
	Mail   YAMLEmailMail   `yaml:"mail"`
}

type YAMLEmailServer struct {
	Host    string `yaml:"host"`
	Port    int	   `yaml:"port"`
	Timeout string `yaml:"timeout"`

	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type YAMLEmailMail struct {
	Recipients interface{} `yaml:"recipients"`
	Name       string      `yaml:"name"`
	Subject    string      `yaml:"subject"`
	Message    string      `yaml:"message"`
}
