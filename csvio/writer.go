package csvio

import (
	"encoding/csv"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
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
		writer:     w,
		limit:      -1,
		withHeader: false,
	}
	if opt.Table {
		wt := tabwriter.NewWriter(w, 0, 0, 1, '.', tabwriter.AlignRight|tabwriter.Debug)
		wCSV.csvwriter = csv.NewWriter(wt)
		wCSV.writer = wt
		wCSV.csvwriter.Comma = 9 // tab
	} else {
		wCSV.csvwriter = csv.NewWriter(w)
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
	tmpl := template.New("master").Funcs(GetTemplateFuncMap())
	if opt.Template != "" {
		wCSV.template = template.Must(tmpl.New("output").Parse(opt.Template))
		wCSV.fields = []string{}
		if wCSV.headerText == "" {
			wCSV.headerText = opt.Template
		}
	}
	return wCSV
}

func (w *WriterCSV) write(from chan map[string]interface{}) {
	vals := <-from
	_ = vals

}
