package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

var TemplateFuncMap = template.FuncMap{
	"Add":       func(x, y float64) float64 { return (x + y) },
	"Contains":  func(s, substr string) bool { return strings.Contains(s, substr) },
	"Div":       func(x, y float64) float64 { return (x / y) },
	"GetEnv":    func(name string) string { return os.Getenv(name) },
	"HasPrefix": func(s, prefix string) bool { return strings.HasPrefix(s, prefix) },
	"HasSuffix": func(s, suffix string) bool { return strings.HasSuffix(s, suffix) },
	"Join":      func(a []string, sep string) string { return strings.Join(a, sep) },
	"Mul":       func(x, y float64) float64 { return x * y },
	"Replace":   func(s, old, new string, n int) string { return strings.Replace(s, old, new, n) },
	"Split":     func(s, sep string) []string { return strings.Split(s, sep) },
	"Sub":       func(a, b float64) float64 { return a - b },
	"Title":     func(s string) string { return strings.Title(s) },
	"TrimSpace": func(s string) string { return strings.TrimSpace(s) },
	"ToLower":   func(s string) string { return strings.ToLower(s) },
	"ToUpper":   func(s string) string { return strings.ToUpper(s) },
}

//MatchFile ...
type MatchFile struct {
	MatchFile string            `yaml:",omitempty"`
	Delimiter string            `yaml:",omitempty" short:"d" long:"delimiter" default:","`
	Fields    string            `yaml:",omitempty" short:"i" long:"inputFields" default:"A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z"`
	Comment   string            `yaml:",omitempty"`
	Trim      string            `yaml:",omitempty" default:"L"`
	Time      string            `yaml:",omitempty"`
	Skip      int               `yaml:",omitempty" short:"s" long:"skip" default:"0"`
	Template  map[string]string `yaml:",omitempty"`
	Field     map[string]string `yaml:",omitempty" short:"a" long:"field"`
}

//Options ...
type Options struct {
	ConfigFile  string `short:"c" long:"conf"`
	config      []MatchFile
	ConfigInput MatchFile
	Output      OutputOptions
}

//OutputOptions ...
type OutputOptions struct {
	Template string `short:"t" long:"outputTemplate"`
	Fields   string `short:"o" long:"output"`
}

//ReaderCSV ...
type ReaderCSV struct {
	input     string
	filePath  string
	trim      string
	delimiter rune
	comment   string
	skip      int
	fieldMap  map[string]field
	valueMap  map[string]interface{}
	record    []string
	line      int
	recordSet [][]string
	reader    *csv.Reader
}

type field struct {
	fieldType  string
	fieldIndex int
	template   *template.Template
}

//WriterCSV ...
type WriterCSV struct {
	writer *csv.Writer
	limit  int
	line   int
}

//ReadConfig ...
func (c *Options) ReadConfig() {

	src, err := ioutil.ReadFile(c.ConfigFile)
	if err != nil {
		log.Fatalf("error at ReadFile %s: %v", c.ConfigFile, err)
	}
	err = yaml.Unmarshal(src, &c.config)
	if err != nil {
		log.Fatalf("error at yaml.Unmarshal: %v", err)
	}
}

func nvlint(p1, p2 int) int {
	if p1 != 0 {
		return p1
	}
	return p2
}
func nvl(p1, p2 string) string {
	if p1 != "" {
		return p1
	}
	return p2
}

// NewWriterCSV ---
func NewWriterCSV(w io.Writer) *WriterCSV {
	return &WriterCSV{
		writer: csv.NewWriter(os.Stdout),
		limit:  -1,
	}
}

// NewReaderCSV ---
func NewReaderCSV(reader io.Reader, input string, cfg []MatchFile) *ReaderCSV {
	rCSV := &ReaderCSV{
		delimiter: ',',
		skip:      0,
		fieldMap:  make(map[string]field),
		valueMap:  make(map[string]interface{}),
	}
	tmpl := template.New("master").Funcs(TemplateFuncMap)
	for _, c := range cfg {
		if matched, _ := regexp.MatchString(c.MatchFile, input); !matched {
			continue
		}
		if c.Delimiter != "" {
			rCSV.delimiter, _ = utf8.DecodeRuneInString(c.Delimiter)
		}
		if c.Trim != "" {
			rCSV.trim = c.Trim
		}
		if c.Comment != "" {
			rCSV.comment = c.Comment
		}
		if c.Skip > 0 {
			rCSV.skip = c.Skip
		}
		for i, v := range strings.Split(c.Fields, ",") {
			s := strings.Split(strings.TrimSpace(v), ".")
			f, ok := rCSV.fieldMap[s[0]]
			if ok == false {
				f = field{fieldIndex: i}
			} else {
				f.fieldIndex = i
			}
			if len(s) > 1 {
				f.fieldType = s[1]
			} else {
				f.fieldType = "string"
			}
			rCSV.fieldMap[s[0]] = f
			rCSV.valueMap[s[0]] = ""
		}
		for k, v := range c.Template {
			_ = template.Must(tmpl.New(k).Parse(v))
		}
		for k, v := range c.Field {
			t := template.Must(tmpl.New(k).Parse(v))
			if f, ok := rCSV.fieldMap[k]; ok == false {
				rCSV.fieldMap[k] = field{template: t, fieldIndex: -1}
			} else {
				f.template = t
				rCSV.fieldMap[k] = f
			}
		}
	}
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1
	if strings.Contains(rCSV.trim, "L") {
		r.TrimLeadingSpace = true
	}
	r.Comma = rCSV.delimiter
	rCSV.reader = r

	return rCSV
}

func atoi(a string) int {
	i, _ := strconv.Atoi(a)
	return i
}

//Query ...
func (r *ReaderCSV) Query(w *WriterCSV, outFields string) {
	r.line = 0
	w.line = 0
	for {
		if w.limit > -1 && w.line >= w.limit {
			break
		}
		var rcOut []string
		rcIn, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading input file %s: %v", r.input, err)
		}
		r.line++
		if strings.HasPrefix(rcIn[0], r.comment) || r.line <= r.skip {
			continue
		}

		for k := range r.valueMap {
			if i := r.fieldMap[k].fieldIndex; i >= 0 && i < len(rcIn) {
				if strings.Contains(r.trim, "R") { // L nam resi csv.reader
					rcIn[i] = strings.TrimRight(rcIn[i], " ")
				}
				switch r.fieldMap[k].fieldType {
				case "Int()":
					if s, err := strconv.Atoi(rcIn[i]); err == nil {
						r.valueMap[k] = s
					}
				case "Float()":
					if s, err := strconv.ParseFloat(rcIn[i], 64); err == nil {
						r.valueMap[k] = s
					}
				default:
					r.valueMap[k] = rcIn[i]
				}
			}
		}

		for _, v := range strings.Split(outFields, ",") {
			v = strings.TrimSpace(v)
			if tmpl := r.fieldMap[v].template; tmpl != nil {
				var buf bytes.Buffer
				err := tmpl.Execute(&buf, r.valueMap)
				if err != nil {
					log.Println("executing template:", err)
				}
				rcOut = append(rcOut, buf.String())
			} else if i := r.fieldMap[v].fieldIndex; i >= 0 && i < len(rcIn) {
				rcOut = append(rcOut, rcIn[i])

			}
		}
		//rsetOut = append(rsetOut, rcOut)
		w.line++
		if err := w.writer.Write(rcOut); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	// Write any buffered data to the underlying writer (standard output).
	w.writer.Flush()

	if err := w.writer.Error(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var options Options
	args, err := flags.ParseArgs(&options, os.Args)
	if err != nil {
		panic(err)
	}
	if options.ConfigFile == "" {
		options.config = append(options.config, options.ConfigInput)
	} else {
		options.ReadConfig()
	}
	fmt.Println(options)
	fmt.Println(args)
	fmt.Println("")
	for _, file := range args[1:] {
		fh, err := os.Open(file)
		if err != nil {
			log.Fatalf("error openning file : %v", err)
		}
		defer fh.Close()

		r := NewReaderCSV(fh, path.Base(file), options.config)
		fmt.Println(r)
		w := NewWriterCSV(os.Stdout)
		w.limit = 10
		//log.Printf("%+v", r)
		r.Query(w, options.Output.Fields)
		_ = tablewriter.NewWriter(os.Stdout)
	}
}
