package csvio

import (
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/mgo.v2"
)

type mgoDB struct {
	db  *mgo.Database
	col map[string]*mgo.Collection
}

var m mgoDB

//MgoLookup ...
func MgoLookup(db, col, id string) map[string]interface{} {
	if m.db == nil {
		session, err := mgo.Dial("localhost")
		if err != nil {
			panic(err)
		}
		session.SetMode(mgo.Monotonic, true)
		m.db = session.DB(db)
		m.col = make(map[string]*mgo.Collection)
	}
	if _, ok := m.col[col]; !ok {
		m.col[col] = m.db.C(col)
	}
	result := make(map[string]string)
	err := m.col[col].FindId(id).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

//GetTemplateFuncMap ...
func GetTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
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
		"Float":     func(s string) (float64, error) { return strconv.ParseFloat(s, 64) },
		"MLookup":   func(db, col, id string) map[string]interface{} { return MgoLookup(db, col, id) },
	}
}
