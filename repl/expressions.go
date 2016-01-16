package zygo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	//"github.com/shurcooL/go-goon"
)

type Sexp interface {
	SexpString() string
}

type SexpPair struct {
	Head Sexp
	Tail Sexp
}
type SexpInt int
type SexpBool bool
type SexpFloat float64
type SexpChar rune
type SexpStr string
type SexpRaw []byte
type SexpReflect reflect.Value
type SexpError struct {
	error
}
type SexpSentinel int

const (
	SexpNull SexpSentinel = iota
	SexpEnd
	SexpMarker
)

func (sent SexpSentinel) SexpString() string {
	if sent == SexpNull {
		return "()"
	}
	if sent == SexpEnd {
		return "End"
	}
	if sent == SexpMarker {
		return "Marker"
	}

	return ""
}

func Cons(a Sexp, b Sexp) SexpPair {
	return SexpPair{a, b}
}

//func (pair SexpPair) Head() Sexp {
//	return pair.Head
//}

//func (pair SexpPair) Tail() Sexp {
//	return pair.Tail
//}

func (pair SexpPair) SexpString() string {
	str := "("

	for {
		switch pair.Tail.(type) {
		case SexpPair:
			str += pair.Head.SexpString() + " "
			pair = pair.Tail.(SexpPair)
			continue
		}
		break
	}

	str += pair.Head.SexpString()

	if pair.Tail == SexpNull {
		str += ")"
	} else {
		str += " \\ " + pair.Tail.SexpString() + ")"
	}

	return str
}

type SexpArray []Sexp

func (e SexpError) SexpString() string {
	return e.error.Error()
}

type EmbedPath struct {
	ChildName     string
	ChildFieldNum int
}

func GetEmbedPath(e []EmbedPath) string {
	r := ""
	last := len(e) - 1
	for i, s := range e {
		r += s.ChildName
		if i < last {
			r += ":"
		}
	}
	return r
}

type HashFieldDet struct {
	FieldNum     int
	FieldType    reflect.Type
	StructField  reflect.StructField
	FieldName    string
	FieldJsonTag string
	EmbedPath    []EmbedPath // we are embedded if len(EmbedPath) > 0
}
type SexpHash struct {
	TypeName         *string
	Map              map[int][]SexpPair
	KeyOrder         *[]Sexp // must user pointer here, else hset! will fail to update.
	GoStructFactory  *RegistryEntry
	NumKeys          *int
	GoMethods        *[]reflect.Method
	GoFields         *[]reflect.StructField
	GoMethSx         *SexpArray
	GoFieldSx        *SexpArray
	GoType           *reflect.Type
	NumMethod        *int
	GoShadowStruct   *interface{}
	GoShadowStructVa *reflect.Value

	// json tag name -> pointers to example values, as factories for SexpToGoStructs()
	JsonTagMap *map[string]*HashFieldDet
	DetOrder   *[]*HashFieldDet
}

func (h *SexpHash) SetGoStructFactory(factory RegistryEntry) {
	(*h.GoStructFactory) = factory
}

var SexpIntSize = reflect.TypeOf(SexpInt(0)).Bits()
var SexpFloatSize = reflect.TypeOf(SexpFloat(0.0)).Bits()

func (r SexpReflect) SexpString() string {
	return fmt.Sprintf("%#v", r)
}

func (arr SexpArray) SexpString() string {
	if len(arr) == 0 {
		return "[]"
	}

	str := "[" + arr[0].SexpString()
	for _, sexp := range arr[1:] {
		str += " " + sexp.SexpString()
	}
	str += "]"
	return str
}

func (hash SexpHash) SexpString() string {
	if *hash.TypeName != "hash" {
		return NamedHashSexpString(hash)
	}
	str := "{"
	for _, arr := range hash.Map {
		for _, pair := range arr {
			str += pair.Head.SexpString() + " "
			str += pair.Tail.SexpString() + " "
		}
	}
	if len(str) > 1 {
		return str[:len(str)-1] + "}"
	}
	return str + "}"
}

func NamedHashSexpString(hash SexpHash) string {
	str := " (" + *hash.TypeName + " "

	for _, key := range *hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		if err == nil {
			switch s := key.(type) {
			case SexpStr:
				str += string(s) + ":"
			case SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString() + ":"
			}

			str += val.SexpString() + " "
		} else {
			panic(err)
		}
	}
	if len(hash.Map) > 0 {
		return str[:len(str)-1] + ")"
	}
	return str + ")"
}

func (b SexpBool) SexpString() string {
	if b {
		return "true"
	}
	return "false"
}

func (i SexpInt) SexpString() string {
	return strconv.Itoa(int(i))
}

func (f SexpFloat) SexpString() string {
	return strconv.FormatFloat(float64(f), 'g', 5, SexpFloatSize)
}

func (c SexpChar) SexpString() string {
	return "#" + strings.Trim(strconv.QuoteRune(rune(c)), "'")
}

func (s SexpStr) SexpString() string {
	return strconv.Quote(string(s))
}

func (r SexpRaw) SexpString() string {
	return fmt.Sprintf("%#v", []byte(r))
}

type SexpSymbol struct {
	name   string
	number int
}

func (sym SexpSymbol) SexpString() string {
	return sym.name
}

func (sym SexpSymbol) Name() string {
	return sym.name
}

func (sym SexpSymbol) Number() int {
	return sym.number
}

type SexpFunction struct {
	name       string
	user       bool
	nargs      int
	varargs    bool
	fun        GlispFunction
	userfun    GlispUserFunction
	closeScope *Stack
	orig       Sexp
}

func (sf SexpFunction) SexpString() string {
	if sf.orig == nil {
		return "fn [" + sf.name + "]"
	}
	return sf.orig.SexpString()
}

func IsTruthy(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpBool:
		return bool(e)
	case SexpInt:
		return e != 0
	case SexpChar:
		return e != 0
	case SexpSentinel:
		return e != SexpNull
	}
	return true
}

type SexpStackmark struct {
	sym SexpSymbol
}

func (mark SexpStackmark) SexpString() string {
	return "stackmark " + mark.sym.name
}
