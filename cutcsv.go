package cutcsv

import (
	"bytes"
	"encoding/csv"
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

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

//MatchFile ...
type MatchFile struct {
	MatchFile string            `yaml:",omitempty"`
	Delimiter string            `yaml:",omitempty" short:"d" long:"delimiter" default:","`
	Fields    string            `yaml:",omitempty" short:"f" long:"fields" default:"A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z"`
	Comment   string            `yaml:",omitempty"`
	Trim      string            `yaml:",omitempty" default:"L"`
	Time      string            `yaml:",omitempty"`
	Skip      int               `yaml:",omitempty" short:"s" long:"skip" default:"0"`
	Template  map[string]string `yaml:",omitempty"`
	Field     map[string]string `yaml:",omitempty" long:"field"`
}

type InputOptions struct {
	ConfigFile string `long:"conf"`
	config     []MatchFile
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
func (c *InputOptions) ReadConfig(file string) {

	src, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(src, &c.config)
	if err != nil {
		log.Fatalf("error: %v", err)
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
	tmpl := template.New("master").Funcs(CSVTemplateFuncMap)
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
func (r *ReaderCSV) Query(w *WriterCSV, outFields ...string) {
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
			log.Fatal(err)
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

		for _, v := range outFields {
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

	input := "STAT_ISSUE_CPNI20160920NJOB49_1608.DAT"
	var opt InputOptions
	opt.ReadConfig("edgar_csv.yaml")
	//log.Printf("%+v", cfg)
	fi, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	r := NewReaderCSV(fi, path.Base(input), opt.config)
	w := NewWriterCSV(os.Stdout)
	w.limit = 10
	//log.Printf("%+v", r)
	r.Query(w, "SQNU", "RECTYPE", "TACN", "TDNR", "DAIS", "TYPDOC", "AGTN", "FAR_CZK")
	_ = tablewriter.NewWriter(os.Stdout)
}
