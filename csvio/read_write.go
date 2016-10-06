package csvio

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
)

//ReadWriteCSV ...
func ReadWriteCSV(r *ReaderCSV, w *WriterCSV) {
	w.line = 0
	if w.withHeader && w.headerText == "" {
		if err := w.csvwriter.Write(w.fields); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	} else if w.withHeader && w.headerText != "" {
		io.WriteString(w.writer, "\""+w.headerText+"\"\n")
	}
	for {
		if w.limit > -1 && w.line >= w.limit {
			break
		}
		var rcOut []string
		rcIn, err := r.reader.Read()
		//log.Debugln("read line:", r.line)
		if err == io.EOF {
			break
		}
		r.line++
		if err != nil {
			log.Fatalf("error reading input file %s: %v", r.input, err)
		}
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

			//vystupni pole
			for _, v := range w.fields {
				//pole existuje
				if _, ok := r.fieldMap[v]; ok {
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
				} else {
					//prazdna hodnota, pokud pole neexistuje
					rcOut = append(rcOut, "")
				}
			}
		}
		//rsetOut = append(rsetOut, rcOut)
		//log.Debugln("write line:", w.line)
		for k, v := range w.filter {
			if r.valueMap[k] != v {
				rcOut = []string{}
			}
		}
		if len(rcOut) == 0 {
			continue
		}
		w.line++
		if err := w.csvwriter.Write(rcOut); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	// Write any buffered data to the underlying writer (standard output).
	w.csvwriter.Flush()

	if err := w.csvwriter.Error(); err != nil {
		log.Fatal(err)
	}
}
