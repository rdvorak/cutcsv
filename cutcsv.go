package main

import (
	"cutcsv/csvio"
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

var log = logrus.New()

var config []csvio.FileOptions

//CommandOptions ...
type CommandOptions struct {
	ConfigFile string              `short:"c" long:"conf"`
	Input      csvio.InputOptions  `group:"input" namespace:"input"`
	Output     csvio.OutputOptions `group:"output" namespace:"output"`
}

//ReadConfig ...
func (c *CommandOptions) ReadConfig() {

	src, err := ioutil.ReadFile(c.ConfigFile)
	if err != nil {
		log.Fatalf("error at ReadFile %s: %v", c.ConfigFile, err)
	}
	err = yaml.Unmarshal(src, &config)
	if err != nil {
		log.Fatalf("error at yaml.Unmarshal: %v", err)
	}
}

func main() {
	log.Level = logrus.DebugLevel
	var options CommandOptions
	args, err := flags.ParseArgs(&options, os.Args)
	if err != nil {
		panic(err)
	}
	if options.ConfigFile == "" {
		config = append(config, csvio.FileOptions{Input: options.Input, Output: options.Output})
	} else {
		options.ReadConfig()
	}
	//log.Debug(options)
	//the rest  are input files/dirs
	for _, file := range args[1:] {
		fh, err := os.Open(file)
		if err != nil {
			log.Fatalf("error openning file : %v", err)
		}
		defer fh.Close()

		r := csvio.NewReaderCSV(fh, path.Base(file), config)
		//log.Debugf("%+v", r)
		// in case of WriterCSV command line options have precedance over config file options
		w := csvio.NewWriterCSV(os.Stdout, path.Base(file), config, options.Output)

		csvio.ReadWriteCSV(r, w)
		_ = tablewriter.NewWriter(os.Stdout)
	}
}
