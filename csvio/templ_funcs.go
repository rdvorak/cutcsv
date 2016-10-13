package csvio

import (
	"log"
	"math"
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

//SsDiv ....
//analogy to SsMul with the division instead of multiplication
//Example:
//SsDiv "014.15.36E" 0, 3, 1, 4, 6, 60, 7, 9, 3600
//converts geo coordinates to degrees
func SsDiv(s string, posmap []float64) float64 {
	var sum, num, mul float64
	var from, to int
	if len(posmap) > 1 {
		for i, pos := range posmap {
			switch math.Mod(float64(i), 3) {
			case 0:
				if i > 2 {
					sum += num / mul
				}
				from = int(pos)
				num = 1
			case 1:
				to = int(pos)
				num, _ = strconv.ParseFloat(s[from:to], 64)
				mul = 1
			case 2:
				mul = pos
			}
		}
		return sum + (num / mul)
	}
	return 0
}

//SsMul ....
// first argument is string
// remaining arguments are repeating blocks of 3 arguments where first 2 arguments defines substring (slice) of the first argument and the third si the multuplicator, which is optional in the last block
// the result of the block is multiplivation of number from the string and the multuplicator
// the result of the fucntions is sum of block results
// Example:
// SSumMul("01H.30M.20S", 0, 2, 3600, 4, 6, 60, 6, 8)
// calculates number of seconds: 1*3600 + 30*60 + 20 from the given string
func SsMul(s string, posmap []float64) float64 {
	var sum, num, mul float64
	var from, to int
	if len(posmap) > 1 {
		for i, pos := range posmap {
			switch math.Mod(float64(i), 3) {
			case 0:
				if i > 2 {
					sum += num * mul
				}
				from = int(pos)
				num = 1
			case 1:
				to = int(pos)
				num, _ = strconv.ParseFloat(s[from:to], 64)
				mul = 1
			case 2:
				mul = pos
			}
		}
		return sum + (num * mul)
	}
	return 0
}

//SString ....
func SString(s string, posmap []int) string {

	var substr string
	var from, to int
	for i, pos := range posmap {
		if math.Mod(float64(i), 2) == 0 {
			if i > 1 {
				substr += s[from:to]
			}
			from = pos
			to = len(s)
		} else {
			to = pos
		}
	}
	substr += s[from:to]
	return substr
}

//Decode ....
func Decode(s string, maping []string) string {
	var in, out string
	for i, inout := range maping {
		if math.Mod(float64(i), 2) == 0 {
			in = inout
			out = inout
		} else {
			if s == in {
				return inout
			}
			out = ""
		}
	}
	return out
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
		"Trim":      func(s string) string { return strings.TrimSpace(s) },
		"ToLower":   func(s string) string { return strings.ToLower(s) },
		"ToUpper":   func(s string) string { return strings.ToUpper(s) },
		"Float":     func(s string) (float64, error) { return strconv.ParseFloat(s, 64) },
		"HMGet":     func(url, id string) map[string]string { return MgoLookup(url, id) },
		"HGet": func(url, id, key string) string {
			s := MgoLookup(url, id)
			if _, ok := s[key]; ok {
				return s[key]
			}
			return ""
		},
		"SString": func(s string, posmap ...int) string { return SString(s, posmap) },
		"SsMul":   func(s string, posmap ...float64) float64 { return SsMul(s, posmap) },
		"SsDiv":   func(s string, posmap ...float64) float64 { return SsDiv(s, posmap) },
	}
}
