// deaffink is one file compiled to binary that writes out a file called
// "configuration.go" in the directory it is invoked.
//
// Configure anything you require with any number of functions you create.
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
	IsPointer            bool
}

func mkFlags(o *options) {
	flag.StringVar(&o.PName, "package", o.PName, "Set the package name.")
	flag.StringVar(&o.CName, "configurable", o.CName, "Set the configurable item name.")
	flag.BoolVar(&o.IsPointer, "isPointer", o.IsPointer, "Configurable is a pointer.")
}

var (
	path string
	O    *options
	T    *template.Template
	B    *bytes.Buffer
	tErr error
	t    string = `package {{.PName}}

type ConfigFn func({{.CName}}) error

type Config interface {
	Order() int
	Configure({{.CName}}) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure({{.Letter}} {{.CName}}) error {
	return c.fn({{.Letter}})
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

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

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
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

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.{{.Letter}}, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{0, example},
}

func example({{.Letter}} {{.CName}}) error {
	return nil
}
`
)

func init() {
	wd, _ := os.Getwd()
	path = filepath.Join(wd, "configuration.go")
	O = &options{"main", "Item", "i", false}
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
