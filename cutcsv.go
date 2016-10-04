package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

var log = logrus.New()

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
	MatchFile  string            `yaml:",omitempty"`
	Delimiter  string            `yaml:",omitempty" short:"d" long:"Delimiter" default:","`
	Fields     string            `yaml:",omitempty" short:"i" long:"Fields" default:"A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z"`
	Comment    string            `yaml:",omitempty"`
	Trim       string            `yaml:",omitempty" default:"L"`
	Time       string            `yaml:",omitempty"`
	Skip       int               `yaml:",omitempty" short:"s" long:"SkipLines" default:"0"`
	HeaderLine int               `yaml:",omitempty" long:"HeaderLine"`
	Template   map[string]string `yaml:",omitempty"`
	Field      map[string]string `yaml:",omitempty" short:"a" long:"field"`
}

//Options ...
type Options struct {
	ConfigFile string `short:"c" long:"conf"`
	config     []MatchFile
	Input      MatchFile     `group:"input" namespace:"input"`
	Output     OutputOptions `group:"output" namespace:"output"`
}

//OutputOptions ...
type OutputOptions struct {
	Template   string `short:"t" long:"Template"`
	Fields     string `short:"o" long:"Fields"`
	Limit      int    `short:"l" long:"Limit"`
	WithHeader bool   `short:"H" long:"Header"`
	HeaderText string `long:"HeaderText"`
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
	reader    *csv.Reader
}

type field struct {
	fieldType  string
	fieldIndex int
	template   *template.Template
}

//WriterCSV ...
type WriterCSV struct {
	writer     *csv.Writer
	limit      int
	line       int
	fields     []string
	template   *template.Template
	withHeader bool
	headerText string
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
func NewWriterCSV(w io.Writer, opt OutputOptions) *WriterCSV {
	a := &WriterCSV{
		writer:     csv.NewWriter(w),
		limit:      opt.Limit,
		withHeader: opt.WithHeader,
		headerText: opt.HeaderText,
		fields:     strings.Split(strings.Replace(opt.Fields, " ", "", -1), ","),
	}
	if opt.Template != "" {
		a.template = template.Must(template.New("output").Parse(opt.Template))
		a.fields = []string{}
		if a.headerText == "" {
			a.headerText = opt.Template
		}
	}
	return a
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
func View(r *ReaderCSV, w *WriterCSV) {
	r.line = 0
	w.line = 0
	if w.withHeader && w.headerText == "" {
		if err := w.writer.Write(w.fields); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	for {
		if w.limit > -1 && w.line >= w.limit {
			break
		}
		var rcOut []string
		rcIn, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		//log.Debugln("read line:", r.line)
		if err != nil {
			log.Fatalf("error reading input file %s: %v", r.input, err)
		}
		r.line++
		if (r.comment != "" && strings.HasPrefix(rcIn[0], r.comment)) || r.line <= r.skip {
			//log.Debugln("skip line:", r.line)
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

		if w.template != nil {
			var buf bytes.Buffer
			err := w.template.Execute(&buf, r.valueMap)
			if err != nil {
				log.Println("executing template:", err)
			}
			rcOut = []string{buf.String()}

		} else {

			for _, v := range w.fields {
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
		}
		//rsetOut = append(rsetOut, rcOut)
		//log.Debugln("write line:", w.line)
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
	log.Level = logrus.DebugLevel
	var options Options
	args, err := flags.ParseArgs(&options, os.Args)
	if err != nil {
		panic(err)
	}
	if options.ConfigFile == "" {
		options.config = append(options.config, options.Input)
	} else {
		options.ReadConfig()
	}
	log.Debug(options)
	for _, file := range args[1:] {
		fh, err := os.Open(file)
		if err != nil {
			log.Fatalf("error openning file : %v", err)
		}
		defer fh.Close()

		r := NewReaderCSV(fh, path.Base(file), options.config)
		log.Debugf("%+v", r)
		w := NewWriterCSV(os.Stdout, options.Output)
		if w.withHeader && w.headerText != "" {
			io.WriteString(os.Stdout, "\""+w.headerText+"\"\n")
		}
		View(r, w)
		_ = tablewriter.NewWriter(os.Stdout)
	}
}
