\def\ver{0.3}
\def\sname{Common}
\def\stitle{\titlefont \ttitlefont{\sname} - пакет для упрощения разработки команд \ttitlefont{GDBSh}}
\input header

@** Введение.

Пакет \.{common} содержит повторно используемый код для \.{GDBSh}

@** Реализация.

@c
@i license

package common

import (
	@<Импортируемые пакеты@>
)@#

type (
	@<Типы@>
)@#

var (
	@<Глобальные переменные@>
)@#

@
@<Типы@>=
Value interface{}
Result struct {
	Name string
	Val Value
}
Tuple []Result
ResultList  []Result
ValueList  []Value

@* ParseString.
@<Импортируемые пакеты@>=
"strings"

@
@c
// |ParseString| parses GDB result from |s| and produces |[]Result|, a rest of string and verdict everything is ok  
func ParseResult(s string) (ResultList, string, bool) {
	var l ResultList
	for true {
		i:=strings.Index(s, "=")
		if i==-1 {
			return l, s, false
		}
		var r Result
		r.Name = s[0:i]
		glog.V(debug).Infof("Name: '%s'", r.Name)
		var ok bool
		r.Val, s, ok = parseValue(s[i+1:])
		if !ok {
			break
		}
		glog.V(debug).Infof("Value: '%#v', Rest: '%s'", r.Val, s)
		l=append(l, r)
		if len(s)<1 || s[0]!=',' {
			glog.V(debug).Infof("RESULT: %#v", l)
			return l, s, len(l)!=0
		}
		s=s[1:]
	}
	glog.V(debug).Infof("RESULT: %#v", l)
	return l, s, false
}

func parseValue(s string) (interface{}, string, bool) {
	if len(s) == 0  {
		return nil, s, false
	}
	if s[0]=='"' { // CONST
		escaped:=false
		f:=func(r rune) bool {
			if !escaped && r=='"' {
				return true
			}
			if r=='\\' {
				escaped=!escaped
				return false
			}
			escaped=false
			return false
		}
		i:=strings.IndexFunc(s[1:], f)
		if i==-1 {
			return nil, s, false
		}
		s=s[1:]
		glog.V(debug).Infof("CONST: '%s'\n", s[:i])
		return s[:i], s[i+1:], true
	} else if s[0] == '{' { // TUPLE
		var t Tuple
		s=s[1:]
		if len(s)>=1 && s[0]=='}' {
			return t, s[1:], true
		}
		for r, ts, ok:=ParseResult(s); ok; r, ts, ok=ParseResult(s) {
			if len(ts)==0 {
				return nil, s, false
			} else if ts[0]=='}' {
				t=append(t, r...)
				s=ts[1:]
				break
			} else if ts[0]==',' {
				t=append(t, r...)
				s=ts[1:]
				continue
			}
			return nil, s, false
		}
		glog.V(debug).Infof("TUPLE: %#v, Rest: %s\n", t, s)
		return t, s, len(t)!=0
	} else if s[0] == '[' { // LIST
		var vl ValueList
		s=s[1:]
		if len(s)>=1 && s[0]==']' {
			return vl, s[1:], true
		}
		for v, ts, ok := parseValue(s); ok; v, ts, ok = parseValue(s) {
			if len(ts) == 0 {
				return nil, s, false
			} else if ts[0] == ']' {
				vl = append(vl, v)
				s=ts[1:]
				break
			} else if ts[0] == ',' {
				vl = append(vl, v)
				s=ts[1:]
				continue
			}
			return nil, s, false
		}		
		glog.V(debug).Infof("LIST(Values): %#v, Rest: '%s'\n", vl, s)
		if len(vl)!=0 {
			return vl, s, len(vl)!=0
		}
		var rl ResultList
		s=s[1:]
		for r, ts, ok:=ParseResult(s[1:]); ok; r, ts, ok=ParseResult(s[1:]) {
			if len(ts)==0 {
				return nil, s, false
			} else if ts[0]==']' {
				rl=append(rl, r...)
				s=ts[1:]
				break
			} else if ts[0]==',' {
				rl=append(rl, r...)
				s=ts
				continue
			}
			return nil, s, false
		}
		glog.V(debug).Infof("LIST(Results): %#v, Rest: '%s'\n", rl, s)
		return rl, s, true
	}
	return nil, s, false
}

@* Sections.
@<Импортируемые пакеты@>=
"io"
"fmt"
"bufio"

@
@c
// |Sections| retrieve information about sections placed in memory of GDB's target 
// and returns a list of a start and an end addresses, separated by the comma or error, if any. 
func Sections(gdbin io.Reader, gdbout io.Writer) ([]string, error) {
	var sections []string
	if _, err:=io.WriteString(gdbout, "info target\n"); err!=nil {
		return sections, err
	}
	gdbr:=bufio.NewReader(gdbin)
	for s, err:=gdbr.ReadString('\n'); err==nil; s, err=gdbr.ReadString('\n') {
		glog.V(debug).Infof(s)
		if strings.HasPrefix(s, "^") {
			break
		}
		if !strings.HasPrefix(s, "~") {
			continue
		}
		glog.V(debug).Infof(s)
		s=s[4:len(s)-4]
		f:=strings.Fields(s)
		if len(f)>=2 && strings.HasPrefix(f[0], "0x") {
			sections=append(sections, f[0]+","+f[2])
		}
	}
	return sections, nil
}

@
@<Глобальные переменные@>=
rp	=strings.NewReplacer("\\n", "", "\\t", "", "\\\"", "\"")

@* FindAddress.
@c
// |FindAddress| find |address| in |section| with specified |options| and returns
func FindAddress(gdbin io.Reader, gdbout io.Writer, options string, section string, address string) ([]string, error){
	var addrs []string
	if _, err:=fmt.Fprintf(gdbout, "find %s %s,%s\n", options, section, address); err!=nil {
		return addrs, nil
	}
	gdbr:=bufio.NewReader(gdbin)
	for s, err:=gdbr.ReadString('\n'); err==nil; s, err=gdbr.ReadString('\n') {
		glog.V(debug).Infof(s)
		if strings.HasPrefix(s, "^") {
			break
		}
		if !strings.HasPrefix(s, "~") {
			continue	
		}
		if strings.Contains(s, "found") {
			glog.V(debug).Infof("ignore '%s'", s)
			continue
		}
		s=rp.Replace(s)
		if len(s)<= 4 {
			continue
		}
		addrs=append(addrs, strings.Fields(s[2:len(s)-2])[0])
	}
	return addrs, nil
}

@* Vtables.
@c
// |Vtables| obtains addresses of virtual tables of |instance| and returns list of addresses or an error if any
func Vtables(gdbin io.Reader, gdbout io.Writer, instance string) ([]string, error) {
	var vtables []string
	if _, err:=fmt.Fprintf(gdbout, "info vtbl %s\n", instance); err!=nil {
		return vtables, nil
	}
	gdbr:=bufio.NewReader(gdbin)
	for s, err:=gdbr.ReadString('\n'); err==nil; s, err=gdbr.ReadString('\n') {
		glog.V(debug).Infof(s)
		if strings.HasPrefix(s, "^") {
			break
		}
		if !strings.HasPrefix(s, "~\"vtable for") {
			continue	
		}
		s=rp.Replace(s)
		b:=strings.LastIndex(s, "0x")
		e:=strings.LastIndex(s, ")")
		if b==-1 || e==-1 {
			continue
		}
		vtables=append(vtables, s[b:e])
	}
	return vtables, nil
}

@* Result.
@c
// |String| returns formatted result
func (this *Result) String() string {
	return this.StringWIndent(0)
}

// |StringWIndent| returns formatted result is indented with |indent|
func (this *Result) StringWIndent(indent int) string {
	var s string
	switch r:=this.Val.(type) {
		case string: 
			s=fmt.Sprintf("\"%s\"", r)
		case Tuple: 
			s=r.StringWIndent(indent)
		case ResultList:
			s=r.StringWIndent(indent)
		case ValueList:
			s=r.StringWIndent(indent)
	}
	return fmt.Sprintf("%s=%s", this.Name, s)	

}

@* ResultList.
@c
// |String| returns formatted result
func (this *ResultList) String() string {
	return this.StringWIndent(0)
}

// |StringWIndent| returns formatted result is indented with |indent|
func (this *ResultList) StringWIndent(indent int) string {
	s:=""
	if len(*this)>1 {
		for i:=0;i<indent;i++ {
			s+="\t"
		}
	} 
	p:=""
	for i, v:=range *this {
		s+=p+v.StringWIndent(indent)
		if i==0 {
			p="\n"
			for i:=0;i<indent;i++ {
				p+="\t"
			}
		}
	}
	return s
}

@* Tuple.
@c
// |String| returns formatted result
func (this *Tuple) String() string {
	return this.StringWIndent(0)
}

// |StringWIndent| returns formatted result is indented with |indent|
func (this *Tuple) StringWIndent(indent int) string {
	s:="{"
	ind:=indent
	if len(*this)>1 {
		s="{\n"
		ind++
		for i:=0;i<ind;i++ {
			s+="\t"
		}
	} 
	p:=""
	for i, v:=range *this {
		s+=p+v.StringWIndent(ind)
		if i==0 {
			p=",\n"
			for i:=0;i<ind;i++ {
				p+="\t"
			}
		}
	}
	if len(*this)>1 {
		s+="\n"
		for i:=0;i<indent;i++ {
			s+="\t"
		}
	}
	s+="}"
	return s
}

@* ValueList.
@c
// |String| returns formatted result
func (this *ValueList) String() string {
	return this.StringWIndent(0)
}

// |StringWIndent| returns formatted result is indented with |indent|
func (this *ValueList) StringWIndent(indent int) string {
	s:="["
	ind:=indent
	if len(*this)>1 {
		s="[\n"
		ind++
		for i:=0;i<ind;i++ {
			s+="\t"
		}
	}
	p:=""
	for i, v:=range *this {
		switch r:=v.(type) {
			case string: 
				s+=fmt.Sprintf("%s\"%s\"", p, r)
			case Tuple: 
				s+=p+r.StringWIndent(ind)
			case ResultList:
				s+=p+r.StringWIndent(ind)
			case ValueList:
				s+=p+r.StringWIndent(ind)
		}
		if i==0 {
			p=",\n"
			for i:=0;i<ind;i++ {
				p+="\t"
			}
		}
	}
	if len(*this)>1 {
		s+="\n"
		for i:=0;i<indent;i++ {
			s+="\t"
		}
	}
	s+="]"
	return s
}

@* RunCommand.
@<Импортируемые пакеты@>=
"errors"

@
@<Глобальные переменные@>=
UnknownError	= errors.New("Unknown error")

@
@c
// |RunCommand| runs |cmd| and returns output, results or an error if any
func RunCommand(gdbin io.Reader, gdbout io.Writer, cmd string) ([]string, ResultList, error) {
	if _, err:=io.WriteString(gdbout, cmd+"\n"); err!=nil {
		return nil, nil, err
	}
	var out []string
	var res ResultList
	var err error
	gdbr:=bufio.NewReader(gdbin)
	for s, e:=gdbr.ReadString('\n'); e==nil; s, e=gdbr.ReadString('\n') {
		glog.V(debug).Infof(s)
		if strings.HasPrefix(s, "^error") {
			s=s[6:]
			if len(s)==0 || s[0]!=',' {
				break
			}
			v, _, ok:=ParseResult(s[1:])
			if ok && len(v)!=0 && v[0].Name=="msg" {
				err = errors.New(v[0].Val.(string))
			} else {
				err = UnknownError
			}
			break
		} else if strings.HasPrefix(s, "^done") {
			s=s[5:]
			if len(s)==0 || s[0]!=',' {
				break
			}
			res, _, _=ParseResult(s[1:])
			break
		} else if !strings.HasPrefix(s, "~") {
			continue
		}
		out=append(out, s[2:len(s)-2])
	}
	return out, res, err
}

@* |Get| functions.
@c
// |Get| return |Result| by name and a success of the operation
func (this *ResultList) Get(n string) (Result, bool) {
	for _, v:=range *this {
		if v.Name==n {
			return v, true
		}
	}
	return Result{}, false
}

// |Get| return |Result| by name and a success of the operation
func (this *Tuple) Get(n string) (Result, bool) {
	for _, v:=range *this {
		if v.Name==n {
			return v, true
		}
	}
	return Result{}, false
}

@
@<Импортируемые пакеты@>=
"github.com/golang/glog"

@
@<Глобальные переменные@>=
debug glog.Level=0


@** Индекс.
