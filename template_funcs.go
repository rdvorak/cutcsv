package cutcsv

import (
	"os"
	"strings"
	"text/template"
)

//CSVTemplateFuncMap ......
var CSVTemplateFuncMap = template.FuncMap{
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
