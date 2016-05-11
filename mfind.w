\def\ver{0.31}
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
	debug glog.Level=1
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
"flag"

@
@<Глобальные переменные@>=
options	string
values	[]string
help	bool
size	string
num		uint

@
@<Обработать аргументы командной строки@>=
{
	flag.BoolVar(&help, "help", false, "print the help")
	flag.StringVar(&size, "size", "", "search query size: b (bytes), h (halfwords - two bytes), w (words - four bytes), g (giant words - eight bytes)")
	flag.UintVar(&num, "number", 0, "maximum number of finds (default all)")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			"mfind 0.31, GDB extention command for using from GDBSh\n",
			"Copyright (C) 2015, 2016 Alexander Sychev\n",
			"Search memory for the sequences of bytes\n",
			"Usage:\n\tmfind [options] <sequence1> [<sequence2>...]\n",
			"Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if len(size)!=0 {
		if len(size)>1 {
			fmt.Fprint(os.Stderr, "wrong search query size: %s", size)
			flag.Usage()
			return
		}
		switch size[0] {
			case 'b', 'h', 'w', 'g':
				options+="/"+size
			default:
				fmt.Fprint(os.Stderr, "wrong search query size: %s", size)
				flag.Usage()
				return
		}
	}
	if num!=0 {
		options+=fmt.Sprintf(" /%d", num)
	}
	values=flag.Args()

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
	if len(values)!=0 {
		for _, val:=range values {
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
