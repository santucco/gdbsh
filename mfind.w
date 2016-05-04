\def\ver{0.3}
\def\sname{MFind}
\def\stitle{\titlefont \ttitlefont{\sname} - команда \ttitlefont{GDBSh} для поиска по всей памяти процесса}
\input header

@** Введение.

Команда \.{mfind} создана как команда расширения функционала \.{GDB} через \.{GDBSh}.
Она ищет переданные значения по всей занимаемой процессом памяти. 


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
	defer os.Stdout.Close()
	@<Получение секций@>
	@<Искать адреса во всех секциях@>
}

@
@<Импортируемые пакеты@>=
"fmt"
"strings"

@
@<Глобальные переменные@>=
options	string
values	string

@
@<Обработать аргументы командной строки@>=
{
	if len(os.Args)==2 && strings.TrimSpace(os.Args[1])=="-h" {
		fmt.Fprint(os.Stderr,
			"mfind 0.3, GDB extention command for using from GDBSh\n",
			"Copyright (C) 2015, 2016 Alexander Sychev\n",
			"Usage:\n\tmfind [/SN] VAL1,VAL2 [VAL2_1,VAL2_2 ...]\n",
			"Search memory for the sequence of bytes specified by VAL1, VAL2, etc\n", 
			"\tS, search query size:\n",
			"\t'b'\n",
			"\t\tbytes\n",
			"\t'h'\n",
			"\t\thalfwords (two bytes)\n",
			"\t'w'\n",
			"\t\twords (four bytes)\n",
			"\t'g'\n",
			"\t\tgiant words (eight bytes)\n",
			"\tN, maximum number of finds (default all)\n")
		return
	}
	for i:=1; i<len(os.Args); i++ {
		if os.Args[i][0]=='/' {
			options+=os.Args[i]
		} else {
			values+=" "+os.Args[i]
		}
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
@<Искать адреса во всех секциях@>=
{
	glog.V(debug).Infof("%#v", sections)
	vl:=strings.Fields(values)
	if len(vl)!=0 {
		for _, val:=range vl {
			@<Искать |val| во всех секциях@>
		}
	} else {
		stdr:=bufio.NewReader(os.Stdin)
		for val, err:=stdr.ReadString('\n'); err==nil; val, err=stdr.ReadString('\n') {
			val=strings.TrimSpace(val)
			@<Искать |val| во всех секциях@>
		}	
	}
	
}

@
@<Импортируемые пакеты@>=
"io"

@
@<Искать |val| во всех секциях@>=
{
	v:=fmt.Sprintf("%s:\n", val)
	for _, a:=range sections {
		al, err:=common.FindAddress(gdbin, gdbout, options, a, val)
		if err!=nil {
			fmt.Fprintf(os.Stderr, "can't find address %s in section: %s\n", val, a, err)
			continue	
		}
		glog.V(debug).Infof("%s found in %#v", val, al)
		for _, s:=range al {
			v=fmt.Sprintf("%s\t%s\n", v, s)	
			if n, err:=io.WriteString(os.Stdout, v); err!=nil || n!=len(v) {
				glog.Warningf("can't write '%s' to stdout, %d bytes has been written: %s", v, n, err)
				return
			}
			v=""
		}
	}
}

@** Индекс.
