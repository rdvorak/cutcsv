package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"gopkg.in/yaml.v2"
	"encoding/csv"
	"path"
	"regexp"
	"os"
	"io"
	"unicode/utf8"
	"strings"
	"github.com/olekukonko/tablewriter"
	)
type Field_cfgT struct{
		Default  map[string]string `yaml:",omitempty"`
		Match	 map[string]string `yaml:",omitempty"`
		Printf   map[string][]string `yaml:",omitempty"`
		Type   map[string]string `yaml:",omitempty"`
		Case   map[string]map[string]string `yaml:",omitempty"`
	}	
type FileGrpT struct{
	Match, Delimiter, Comment string `yaml:",omitempty"`
    Trim, Time string `yaml:",omitempty"`
	Skip int `yaml:",omitempty"`
	Field Field_cfgT
	File []struct{
		Match, Delimiter, Comment string `yaml:",omitempty"`
		Trim, Time string `yaml:",omitempty"`
		Skip int `yaml:",omitempty"`
		Fields []string `yaml:",omitempty"`
		Field Field_cfgT
		Pipe []struct{
			Line []map[string][]string `yaml:",omitempty"`
		}
	}
}


var cfg []FileGrpT

type FieldT struct{
				index int
				ftype string
				printf []string
				defaul string
				value string
				regex *regexp.Regexp
				submatch []string
				cases map[string]string
		}


type ReaderCSV struct{
	FileName string
	FilePath string
	Delimiter string
	Comment string
	Skip int
	Fields map[string]FieldT
	Match map[string]FieldT
	Type map[string]FieldT
	Default map[string]FieldT
	Printf map[string]FieldT
	Case map[string]FieldT
	Record []string
	RecordSet [][]string
	Reader *csv.Reader
}

func read_conf_file(fp string) {
 cfg_src, _ := ioutil.ReadFile(fp)
 err := yaml.Unmarshal(cfg_src, &cfg)
 if err != nil {
   log.Fatalf("error: %v", err)
 }
}

func nvlint (p1, p2 int) int {
	if p1 != 0 {
		return p1
	} else {
		return p2
	}
}
func nvl (p1, p2 string) string {
	if p1 != "" {
		return p1
	} else {
		return p2
	}
}
func NewReaderCSV(fi io.Reader, fp string) *ReaderCSV{
 fn := path.Base(fp)
 for j, c := range cfg {
	if matched, _ := regexp.MatchString(c.Match, fn); !matched  {
		continue
	}
	 for _, file := range cfg[j].File {
		if matched, _ := regexp.MatchString(file.Match, fn); !matched  {
			continue
		}
		fm := make(map[string]FieldT)
		mm := make(map[string]FieldT)
		pm := make(map[string]FieldT)
		dm := make(map[string]FieldT)
		tm := make(map[string]FieldT)
		cm := make(map[string]FieldT)
		for i, f := range file.Fields {
			fm[string(f)] = FieldT{index: i, ftype: "string"}
		}
		// definice poli mame na urovni skupiny a prednostne souboru
		for _, def := range []Field_cfgT{c.Field, file.Field} {

			// v fm mame vsechna pole, duplicitne pak pro Default,..
			for k, v := range def.Default {
				i, ok := fm[k]
				if ok {
					i.defaul = v
					fm[k] = i
				} else {
					fm[k] = FieldT{index: len(fm) , ftype: "string", defaul: v}
				}
				dm[k] = fm[k]
			}

			for k, v := range def.Match {
				i, ok := fm[k]
				if ok {
					i.regex = regexp.MustCompile(v)
					fm[k] = i
				} else {
					fm[k] = FieldT{index: len(fm) , ftype: "string", regex: regexp.MustCompile(v)}
				}
				mm[k] = fm[k]
			}

			for k, v := range def.Printf {
				i, ok := fm[k]
				if ok {
					i.printf = v
					fm[k] = i
				} else {
					fm[k] = FieldT{index: len(fm) , ftype: "string", printf: v}
				}
				pm[k] = fm[k]
			}
				for k, v := range def.Type {
				i, ok := fm[k]
				if ok {
					i.ftype = v
					fm[k] = i
				} else {
					fm[k] = FieldT{index: len(fm) , ftype: v}
				}
				tm[k] = fm[k]
			}

			for k, v := range def.Case {
				i, ok := fm[k]
				if ok {
					i.cases = v
					fm[k] = i
				} else {
					fm[k] = FieldT{index: len(fm) , ftype: "string", cases: v}
				}
				cm[k] = fm[k]
			}

		}
		r := csv.NewReader(fi)
		ru, _ := utf8.DecodeRuneInString( nvl(c.Delimiter, file.Delimiter ))
		r.Comma = ru
		r.FieldsPerRecord = -1
		r.TrimLeadingSpace = true

		return &ReaderCSV{
			FileName: fn,
			FilePath: fp,
			Comment:  nvl(c.Comment, file.Comment ),
			Skip: nvlint(c.Skip, file.Skip),
			Fields: fm,
			Match: mm,
			Default: dm,
			Printf: pm,
			Type: tm,
			Case: cm,
			Reader: r,

		}
	 }
 }
 return nil
}

func atoi( a string) int {
	i, _ := strconv.Atoi(a)
	return i
}
func (r *ReaderCSV) Query(fields ...string) [][]string {
	indexes := make([]int, len(fields))
	for i, f := range fields {
		indexes[i] = r.Fields[f].index 
	}
	var rcs [][]string
	for _, row := range r.RecordSet {
		rc := make([]string, len(indexes))
		for i, index := range indexes {
			rc[i] = row[index]
		}
		rcs = append(rcs, rc)
	}
	return rcs
	
}
func (r *ReaderCSV) read() {
i := 0
rx := regexp.MustCompile(`\[(\w+)\](?:\[(\d+)\])*`)
      for {
        rc, err := r.Reader.Read()
        if err == io.EOF  {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
		i++
			
        if rc[0] == r.Comment || i <= r.Skip {
            continue
        }

		//doplnime pridana vlastni pole
		if l := len(r.Fields) - len(rc); l > 0 {
			rc = append(rc, make([]string, l)...)
		}

		for _, v := range r.Default {
			if rc[v.index] == "" {
				rc[v.index] = v.defaul
			}
		}

		for k, v := range r.Match {
			v.submatch = v.regex.FindStringSubmatch(rc[v.index])
			r.Match[k] = v
		}

		for _, v := range r.Printf {
			for _, y := range v.printf[1:] {
			
			x :=  rx.FindStringSubmatch(y) // vraci pole 3 lementu 
			//varianty %[SMOD] nebo %[SMOD][0]
			if len(x) == 3 && x[2] != "" {
				q := strings.Replace(y, x[0], r.Match[x[1]].submatch[atoi(x[2])], -1)
				rc[v.index] = fmt.Sprintf(v.printf[0], q)
			} else if len(x) == 3 && x[2] == "" {
				q := strings.Replace(y, x[0], rc[r.Fields[x[1]].index], -1)
				rc[v.index] = fmt.Sprintf(v.printf[0], q)
			}
			}
			
		}

		for _, v := range r.Case {
			rc[v.index] = v.cases[rc[v.index]]
			//fmt.Println(v.index)
		}

		r.Record = rc
		r.RecordSet = append(r.RecordSet , rc)
		//fmt.Println(rc)

	}
}
func main() {
	fp := "STAT_ISSUE_FOP20160818NJOB49_1607.DAT"
	read_conf_file("edgar_csv.yml")
	fi, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer  fi.Close()

	r := NewReaderCSV(fi,fp) 
	r.read()
	table := tablewriter.NewWriter(os.Stdout)
	table.AppendBulk(r.Query("RECTYPE", "FLAG", "TACN", "TDNR", "DAIS", "TYPDOC", "AGTN", "PNR", "PNRA"))
	table.Render()
	
}	
