// deaffink is one file compiled to binary that writes out a file called
// "configuration.go" in the directory it is invoked.
//
// Configure anything you require with any number of functions you add.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type options struct {
	PName, CName, Letter string
	IsPointer, Document  bool
}

func mkFlags(o *options) {
	flag.StringVar(&o.PName, "package", o.PName, "Set the package name.")
	flag.StringVar(&o.CName, "configurable", o.CName, "Set the configurable item name.")
	flag.BoolVar(&o.IsPointer, "isPointer", o.IsPointer, "Configurable is a pointer.")
	flag.BoolVar(&o.Document, "document", o.Document, "Generate documentation strings on public functions, variables, and structs.")
}

var (
	path string
	O    *options
	T    *template.Template
	B    *bytes.Buffer
	tErr error
	t    string = `package {{.PName}}
{{if .Document}}/*
This file provides configuration functionality for {{.CName}} of {{.PName}}.

ConfigFn: A function taking {{.CName}} and returning an error.

Config: An interface that provides order & ConfigFn functionality.

Configuration: An interface that aggregates multiple Config.

An example use:
	package {{.PName}}

	type {{.CName}} struct {
		Configuration
	}

	func New(conf ...Config) ({{.CName}}, error) {
		{{.Letter}} := (instance of {{.CName}})
		c := newConfiguration({{.Letter}}, conf...)
		{{.Letter}}.Configuration = c
		err := {{.Letter}}.Configure()
		if err != nil {
			return nil, err
		}
		return {{.Letter}}, nil
	}
*/{{end}}

{{if .Document}}// A function taking {{.CName}} and returning an error.{{end}}
type ConfigFn func({{.CName}}) error

{{if .Document}}// An interface providing Order & Configure functions.{{end}}
type Config interface {
	Order() int
	Configure({{.CName}}) error
}

type config struct {
	order int
	fn    ConfigFn
}

{{if .Document}}// Returns a default Config with order of 50 and the provided ConfigFn.{{end}}
func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

{{if .Document}}// Returns a Config with the provided order and ConfigFn.{{end}}
func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

{{if .Document}}// Returns an integer used for ordering.{{end}}
func (c config) Order() int {
	return c.order
}

{{if .Document}}// Provided a {{.CName}} runs any defined functionality, returning any error.{{end}}
func (c config) Configure({{.Letter}} {{.CName}}) error {
	return c.fn({{.Letter}})
}

type configList []Config

{{if .Document}}// Len for sort.Sort.{{end}}
func (c configList) Len() int {
	return len(c)
}

{{if .Document}}// Swap for sort.Sort.{{end}}
func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

{{if .Document}}// Less for sort.Sort.{{end}}
func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

{{if .Document}}// An interface providing facility for multiple configuration options.{{end}}
type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	{{.Letter}}          {{.CName}}
	configured bool
	list       configList
}

func newConfiguration({{.Letter}} {{.CName}}, conf ...Config) *configuration {
	c := &configuration{
		{{.Letter}}:    {{.Letter}},
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

{{if .Document}}// Adds any number of Config to the Configuration.{{end}}
func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func configure({{.Letter}} {{.CName}}, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure({{.Letter}})
		if err != nil {
			return err
		}
	}
	return nil
}

{{if .Document}}// Runs all configuration for this Configuration, return any encountered error immediately.{{end}}
func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.{{.Letter}}, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

{{if .Document}}// Returns a boolean indicating if Configuration has run Configure.{{end}}
func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{0, builtinConfigurationExample},
}

func builtinConfigurationExample({{.Letter}} {{.CName}}) error {
	return nil
}

{{if .Document}}// An externally available function to provide to a new instance of {{.CName}}.{{end}}
func ExternalConfigurationExample(anything ...interface{}) Config {
	return NewConfig(
		100,
		func({{.Letter}} {{.CName}}) error {
			return nil
		},
	)
}
`
)

func init() {
	wd, _ := os.Getwd()
	path = filepath.Join(wd, "configuration.go")
	O = &options{"main", "Item", "i", false, false}
	mkFlags(O)
	T, tErr = template.New("deaffink").Parse(t)
	B = new(bytes.Buffer)
}

func errOut(e string, p ...interface{}) {
	log.Printf(e, p...)
	os.Exit(-1)
}

func main() {
	// err out if template error
	if tErr != nil {
		errOut("template error: %s", tErr)
	}

	//parse flags
	flag.Parse()

	//is pointer & set letter
	switch {
	case O.IsPointer:
		O.CName = fmt.Sprintf("*%s", O.CName)
		O.Letter = strings.ToLower(O.CName[1:2])
	default:
		O.Letter = strings.ToLower(O.CName[0:1])
	}

	//exec template
	if xErr := T.Execute(B, O); xErr != nil {
		errOut("template execute error: %s", xErr)
	}

	//format & write out go code
	src, fErr := format.Source(B.Bytes())
	if fErr != nil {
		errOut("source format error: %s", fErr)
	}

	// write out file
	if wErr := ioutil.WriteFile(path, src, 0644); wErr != nil {
		errOut("file write error: %s", wErr)
	}

	os.Exit(0)
}
