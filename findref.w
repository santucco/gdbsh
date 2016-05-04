\def\ver{0.3}
\def\sname{FindRef}
\def\stitle{\titlefont \ttitlefont{\sname} - команда \ttitlefont{GDBSh} для поиска ссылок на объекты}
\input header

@** Введение.

Команда \.{FindRef} создана как команда расширения функционала \.{GDB} для через \.{GDBSh}.
Она ищет объекты, предположительно содержащие в своих членах-данных указатель на указанный экземпляр класса с виртуальной таблицей. 


@** Реализация.

@c
@i license

package main

import (
	@<Импортируемые пакеты@>
	"github.com/golang/glog"
	"os"
	"bufio"
)@#

var (
	@<Глобальные переменные@>
	debug glog.Level=0
)@#

func main() {
	defer glog.Flush()
	glog.V(debug).Infoln("main")
	defer glog.V(debug).Infoln("main is done")
	@<Обработать аргументы командной строки@>
	gdbin:=os.NewFile(uintptr(3), "input")
	gdbout:=os.NewFile(uintptr(4), "output")
	defer gdbin.Close()
	defer gdbout.Close()
	@<Получение секций@>
	@<Искать указатели на виртуальные таблицы во всех секциях для всех переданных или введенных экземпляров@>
}

@
@<Импортируемые пакеты@>=
"fmt"
"strings"

@
@<Глобальные переменные@>=
instances	[]string

@
@<Обработать аргументы командной строки@>=
{
	if len(os.Args)==2 && strings.TrimSpace(os.Args[1])=="-h" {
		fmt.Fprint(os.Stderr,
			"findref 0.3, GDB extention command for using from GDBSh\n",
			"Copyright (C) 2015, 2016 Alexander Sychev\n",
			"Usage:\n\tfindref <instance1> [<instance2>...]\n",
			"Search for instances of objects potentially have a reference to the specified instances\n")
		return
	}
	if len(os.Args)>1 {
		instances=os.Args[1:]
	}
}

@
@<Импортируемые пакеты@>=
"bitbucket.org/santucco/gdbsh/common"

@
@<Глобальные переменные@>=
sections	[]string

@
@<Получение секций@>=
{
	var err error
	sections, err=common.Sections(gdbin, gdbout)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "can't get sections from GDB: %s\n", err)
		return
	}
}


@
@<Искать указатели на виртуальные таблицы во всех секциях для всех переданных или введенных экземпляров@>=
{
	glog.V(debug).Infof("instances: %#v", instances)
	if len(instances)!=0 {
		for _, val:=range instances {
			@<Искать указатели на виртуальные таблицы для |val| во всех секциях@>
		}
	} else {
		stdr:=bufio.NewReader(os.Stdin)
		for val, err:=stdr.ReadString('\n'); err==nil; val, err=stdr.ReadString('\n') {
			val=strings.TrimSpace(val)
			@<Искать указатели на виртуальные таблицы для |val| во всех секциях@>
		}
	}
	
}

@
@<Импортируемые пакеты@>=
"io"

@
@<Глобальные переменные@>=
vtables	[]string

@
@<Искать указатели на виртуальные таблицы для |val| во всех секциях@>=
{
	vtables, err:=common.Vtables(gdbin, gdbout, val)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "can't get vtables for %s from GDB: %s\n", val, err)
		return
	}
	glog.V(debug).Infof("vtables: %#v", vtables)
	rl:=make(map[string][]string)
	@<Искать адреса во всех секциях, добавлять |rl|@>
	@<Посмотреть ближайщее окружение найденных адресов@>

}

@
@<Искать адреса во всех секциях, добавлять |rl|@>=
{
	for _, v:=range vtables {
		for _, a:=range sections {
			glog.V(debug).Infof("searching for %#v in  %#v", v, a)
			al, err:=common.FindAddress(gdbin, gdbout, "", a, v)
			if err!=nil {
				fmt.Fprintf(os.Stderr, "can't find address %s in section: %s\n", v, a, err)
				return
			}
			rl[v]=append(rl[v], al...)
		}
	}
	glog.V(debug).Infof("addresses: %#v", rl)
}

@
@<Импортируемые пакеты@>=
"strconv"

@
@<Посмотреть ближайщее окружение найденных адресов@>=
{	
	cmds:=make(map[string]map[int64][]string)
	@<Для всех адресов просмотреть предыдущие адреса на наличие ссылок на виртуальные таблицы, создать команды распечатки содержимого по найденным адресам, преобразованным к реализации@>
	@<Выполнить сформированные команды@>
}


@
@<Для всех адресов просмотреть предыдущие адреса на наличие ссылок на виртуальные таблицы, создать команды распечатки содержимого по найденным адресам, преобразованным к реализации@>=
{
	rp:=strings.NewReplacer("\\n","","\\t","","\\\"","\"")
	d:=true
	for address, r:=range rl {
		for _, a:=range r {
			adr, err:=strconv.ParseInt(a, 0, 64)
			if err!=nil {
				continue
			}
			for i:=adr; i>adr-160; i-=8 {
				o, r, err:=common.RunCommand(gdbin, gdbout, fmt.Sprintf("-data-read-memory-bytes 0x%x 8", i))
				if err!=nil {
					continue
				}
				var a string
				if m, ok:=r.Get("memory"); !ok {
					continue
				} else if vl, ok:=m.Val.(common.ValueList); !ok || len(vl)==0 {
					continue
				} else if t, ok:=vl[0].(common.Tuple); !ok {
					continue
				} else if c, ok:=t.Get("contents"); !ok {
					continue
				} else if s, ok:=c.Val.(string); ok {
					for i:=len(s)-1; i>=0; i-=2 {
						a+=s[i-1:i+1]
					}
				}

				o, _, err=common.RunCommand(gdbin, gdbout, fmt.Sprintf("info symbol 0x%s", a))
				if err!=nil {
					continue
				}
				for _, s:=range o {
					s=rp.Replace(s)
					if strings.HasPrefix(s, "No symbol matches") || len(s)==0 {
						glog.V(debug).Info(s)
						continue
					}
					var p int
					if p=strings.Index(s, " in "); p==-1 {
						p=len(s)
					}
					sym:=s[0:p]
					if p:=strings.LastIndex(sym, "+"); p!=-1 {
						sym=sym[:p]
					}
					glog.V(debug).Info(sym)
					if d {
						if o, _, err=common.RunCommand(gdbin, gdbout, fmt.Sprintf("demangle %s", sym)); err==nil && len(o)!=0 {
							sym=o[0]
						} else if err!=nil {
							d=false
						}
					}
					if  strings.HasPrefix(sym, "vtable for ") {
						if _, ok:=cmds[address]; !ok {
							cmds[address]=make(map[int64][]string)
						}
						cmds[address][i]=append(cmds[address][i],fmt.Sprintf("p *(%s*)0x%x\n",sym[11:], i))
					}
				}
			}
		}
	}
}

@
@<Выполнить сформированные команды@>=
{
	rp:=strings.NewReplacer("\\n", "\n", "\\t", "\t", "\\\"", "\"")
	
	for a, v:=range cmds {
		for aa, c:=range v {
			o, _, err:=common.RunCommand(gdbin, gdbout, c[0])
			if err!=nil {
				continue
			}
			f:=false
			adr:=strings.Trim(a, "0x")
			for _, s:=range o {
				glog.V(debug).Infof("looking for %s in '%s'", adr, s)
				if i:=strings.Index(s, adr); i!=-1{
					glog.V(debug).Infof("%s has been found in '%s'", adr, s)
					f=true
					break
				}	
			}
			if !f {
				continue
			}
			p:=fmt.Sprintf("%s at 0x%x: ", a, aa)
			if n, err:=io.WriteString(os.Stdout, p); err!=nil || n!=len(p) {
				glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s", p, n, err)
				return
			}
			for _, s:=range o {
				glog.V(debug).Info(s)
				s=rp.Replace(s)
				if n, err:=io.WriteString(os.Stdout, s); err!=nil || n!=len(s) {
					glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s", s, n, err)
					return
				}
			}
		}
	}
}

@** Индекс.
