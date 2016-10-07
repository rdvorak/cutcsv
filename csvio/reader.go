package csvio

import (
	"encoding/csv"
	"io"
	"log"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

//ReaderCSV ...
type ReaderCSV struct {
	reader    *csv.Reader
	input     string
	fileNum   int
	filePath  string
	trim      string
	delimiter rune
	comment   string
	skip      int
	fieldMap  map[string]field
	valueMap  map[string]interface{}
	record    []string
	line      int
}

type field struct {
	fieldType  string
	fieldIndex int
	template   *template.Template
}

// NewReaderCSV ---
func NewReaderCSV(reader io.Reader, i int, input string, cfg []FileOptions) *ReaderCSV {
	rCSV := &ReaderCSV{
		delimiter: ',',
		skip:      0,
		fieldMap:  make(map[string]field),
		valueMap:  make(map[string]interface{}),
		fileNum:   i,
	}
	tmpl := template.New("master").Funcs(GetTemplateFuncMap())

	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1

	for _, c := range cfg {
		if matched, _ := regexp.MatchString(c.MatchFile, input); !matched {
			continue
		}
		if c.Input.Codepage == "8859_2" {

			dec := charmap.ISO8859_2.NewDecoder().Reader(reader)
			r = csv.NewReader(dec)
			r.FieldsPerRecord = -1
		}
		if c.Input.Delimiter != "" {
			rCSV.delimiter, _ = utf8.DecodeRuneInString(c.Input.Delimiter)
			r.Comma = rCSV.delimiter
		}
		if c.Input.Trim != "" {
			if strings.Contains(c.Input.Trim, "L") {
				r.TrimLeadingSpace = true
			}
			rCSV.trim = c.Input.Trim
		}
		if c.Input.Comment != "" {
			rCSV.comment = c.Input.Comment
		}
		if c.Input.Skip > 0 {
			rCSV.skip = c.Input.Skip
		}
		if c.Input.HeaderLine > 0 && rCSV.line == 0 {
			for i := 1; true; i++ {
				rcIn, err := r.Read()
				if err == io.EOF {
					break
				}
				rCSV.line++
				if err != nil {
					log.Fatalf("error reading input file : %v", err)
				}

				if i == c.Input.HeaderLine {
					c.Input.Fields = strings.Join(rcIn, ",")

					break
				}
			}

		}
		rCSV.fieldMap = make(map[string]field)
		for i, v := range strings.Split(c.Input.Fields, ",") {
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
		for k, v := range c.Input.Template {
			_ = template.Must(tmpl.New(k).Parse(v))
		}
		for k, v := range c.Input.Field {
			t := template.Must(tmpl.New(k).Parse(v))
			if f, ok := rCSV.fieldMap[k]; ok == false {
				rCSV.fieldMap[k] = field{template: t, fieldIndex: -1}
			} else {
				f.template = t
				rCSV.fieldMap[k] = f
			}
		}
	}

	rCSV.reader = r

	return rCSV
}
