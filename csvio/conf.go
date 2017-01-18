package csvio

//FileOptions ...
type FileOptions struct {
	MatchFile string `yaml:",omitempty"`
	Input     InputOptions
	Output    OutputOptions
}

//InputOptions ...
type InputOptions struct {
	Delimiter  string            `yaml:",omitempty" short:"d" long:"Delimiter" default:","`
	Fields     string            `yaml:",omitempty" short:"i" long:"Fields" default:"A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z"`
	Comment    string            `yaml:",omitempty"`
	Codepage   string            `yaml:",omitempty" long:"Codepage"`
	Trim       string            `yaml:",omitempty" default:"L"`
	Time       string            `yaml:",omitempty"`
	Skip       int               `yaml:",omitempty" short:"s" long:"Skip" default:"0"`
	HeaderLine int               `yaml:",omitempty" long:"HeaderLine"`
	Template   map[string]string `yaml:",omitempty"`
	Field      map[string]string `yaml:",omitempty" short:"a" long:"Field"`
}

//OutputOptions ...
type OutputOptions struct {
	Template      string `yaml:",omitempty" short:"t" long:"Template"`
	Fields        map[string]string
	DefaultFields string            `yaml:",omitempty" short:"o" long:"Fields"`
	FieldSet      string            `yaml:",omitempty" short:"O" long:"fieldSet" default:"default"`
	Delimiter     string            `yaml:",omitempty" long:"Delimiter"`
	Limit         int               `yaml:",omitempty" short:"l" long:"Limit" default:"-1"`
	WithHeader    bool              `yaml:",omitempty" short:"H" long:"Header"`
	HeaderText    string            `yaml:",omitempty" long:"HeaderText"`
	Filter        map[string]string `yaml:",omitempty" short:"F" long:"Filter"`
	UseCRLF       bool              `yaml:",omitempty" long:"UseCRLF"`
	Codepage      string            `yaml:",omitempty" long:"Codepage"`
	Csv           bool              `yaml:",omitempty" long:"csv"`
	Table         bool              `yaml:",omitempty" long:"tab"`
}
