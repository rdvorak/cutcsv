package csvio

import (
	"encoding/csv"
	"io"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"
)

//WriterCSV ...
type WriterCSV struct {
	csvwriter  *csv.Writer
	writer     io.Writer
	limit      int
	line       int
	fields     []string
	filter     map[string]string
	template   *template.Template
	withHeader bool
	headerText string
}

// NewWriterCSV ---
func NewWriterCSV(w io.Writer, input string, cfg []FileOptions, opt OutputOptions) *WriterCSV {
	wCSV := &WriterCSV{
		csvwriter:  csv.NewWriter(w),
		writer:     w,
		limit:      -1,
		withHeader: false,
	}
	for _, c := range cfg {
		if matched, _ := regexp.MatchString(c.MatchFile, input); !matched {
			continue
		}
		if c.Output.Delimiter != "" {
			wCSV.csvwriter.Comma, _ = utf8.DecodeRuneInString(c.Output.Delimiter)
		}

		wCSV.withHeader = c.Output.WithHeader
		wCSV.headerText = c.Output.HeaderText
		wCSV.filter = c.Output.Filter
		wCSV.csvwriter.UseCRLF = c.Output.UseCRLF
		wCSV.fields = []string{}
		for _, v := range strings.Split(c.Output.Fields[opt.FieldSet], ",") {
			wCSV.fields = append(wCSV.fields, strings.TrimSpace(v))
		}
	}
	if opt.Limit > -1 {
		wCSV.limit = opt.Limit
	}
	if opt.WithHeader {
		wCSV.withHeader = opt.WithHeader
	}
	if opt.HeaderText != "" {
		wCSV.headerText = opt.HeaderText
	}
	wCSV.csvwriter.UseCRLF = opt.UseCRLF
	if len(opt.Filter) > 0 {
		wCSV.filter = opt.Filter
	}
	if opt.DefaultFields != "" {
		wCSV.fields = []string{}
		for _, v := range strings.Split(opt.DefaultFields, ",") {
			wCSV.fields = append(wCSV.fields, strings.TrimSpace(v))
		}
	}
	if opt.Template != "" {
		wCSV.template = template.Must(template.New("output").Parse(opt.Template))
		wCSV.fields = []string{}
		if wCSV.headerText == "" {
			wCSV.headerText = opt.Template
		}
	}
	return wCSV
}
