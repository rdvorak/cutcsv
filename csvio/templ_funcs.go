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
	sess map[string]*mgo.Session
	db   map[string]*mgo.Database
	col  map[string]*mgo.Collection
}

var m = mgoDB{
	make(map[string]*mgo.Session),
	make(map[string]*mgo.Database),
	make(map[string]*mgo.Collection),
}

//MgoLookup ...
func MgoLookup(url, id string) map[string]string {
	u := strings.Split(url, "/")
	var srv, db, col, dbkey, colkey string
	if len(u) == 2 {
		u = []string{"localhost", u[0], u[1]}
	} else if len(u) != 3 {
		log.Fatalf("failed to parse mgo url: %s", url)
	}
	srv, db, col, dbkey, colkey = u[0], u[1], u[2], u[0]+u[1], u[0]+u[1]+u[2]
	if _, ok := m.sess[srv]; !ok {
		session, err := mgo.Dial(srv)
		if err != nil {
			panic(err)
		}
		session.SetMode(mgo.Monotonic, true)
		m.sess[srv] = session
	}
	if _, ok := m.db[dbkey]; !ok {
		m.db[dbkey] = m.sess[srv].DB(db)
	}
	if _, ok := m.col[colkey]; !ok {
		m.col[colkey] = m.db[dbkey].C(col)
	}
	result := make(map[string]string)
	_ = m.col[colkey].FindId(id).One(&result)
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
		"HMGet":     func(url, id string) map[string]string { return MgoLookup(url, id) },
	}
}
