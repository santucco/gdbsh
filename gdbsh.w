\def\ver{0.31}
\def\sname{GDBSh}
\def\stitle{\titlefont \ttitlefont{\sname} - командная оболочка для \ttitlefont{GDB}}
\input header
@** Введение.

Отладчик \.{GDB} не имеет некоторых команд, нужных в повседневном использовании, например поиска адресов по памяти процесса. Да, есть команда \.{find},
позволяющая искать в регионе памяти, но процесс занимает не сплошной кусок памяти, a множество отдельных секций, сканировать в каждой из которых нужно отдельно.
В принципе, \.{GDB} можно расширить с помощью скриптов на \.{Python}, но эти скриптовые возможности ограничены.
Так возникла идея использовать механизм \.{GDB/MI} для запуска в \.{GDB} внешних команд.
Впоследствии захотелось выстраивать команды в конвейер, так получился \.{GDBSh}

@** Реализация.

\.{GDBSh} запускает \.{GDB} в режиме интерпретатора, передает на выполнение команды от порождаемых процессов и возвращает в процессы результат выполнения команд.
Каждый дочерний процесс получает, кроме  стандартных дескрипторов, еще два файловых дескриптора для взаимодействия с \.{GDB} через механизм \.{GDB/MI}.
\.{GDBSh} передает команды в \.{GDB} строго поочередно, после завершения выполнения предыдущей команды - это обусловлено невозможностью отличить вывод \.{GDB} для разных команд.

@c
@i license

package main

import (
	@<Импортируемые пакеты@>
)@#

type (
	@<Типы@>
)@#

var (
	@<Глобальные переменные@>
)@#

func main() {
	@<Проверить аргументы командной строки, вывести информацию о программе, если необходимо@>
	@<Подготовить трассировку@>
	@<Инициализация сигнальных обработчиков@>
	@<Запустить \.{GDB}@>
	@<Читать команды из  |stdin|, посылать их в \.{GDB}, обрабатывать результаты@>
	@<Ждать завершения процесса@>
	@<Проверить возвращаемый результат@>
}

@
@<Импортируемые пакеты@>=
"os"
"os/exec"
"io"

@
@<Глобальные переменные@>=
gdbin	io.WriteCloser
gdbout	io.ReadCloser
gdberr	io.ReadCloser
cmd		*exec.Cmd

@ Аргументы командной строки дополняются опцией для вызова интерпретатора \.{GDB}
@<Подготовить аргументы командной строки...@>=
var args []string
args=append(args, os.Args...)
args[0]="--interpreter=mi"

@
@<Запустить \.{GDB}@>=
{
	@<Подготовить аргументы командной строки для запуска \.{GDB}@>
	if cmd=exec.Command("gdb", args...); cmd==nil {
		glog.Errorf("can't create command to run gdb\n")
		return
	}

	var err error
	if gdbin, err=cmd.StdinPipe(); err!=nil {
		glog.Errorf("can't create pipe: %v\n", err)
		return
	}
	defer gdbin.Close()

	if gdbout, err=cmd.StdoutPipe(); err!=nil {
		glog.Errorf("can't create pipe: %v\n", err)
		return
	}
	defer gdbout.Close()

	if gdberr, err=cmd.StderrPipe(); err!=nil {
		glog.Errorf("can't create pipe: %v\n", err)
		return
	}
	defer gdberr.Close()

	if err=cmd.Start(); err!=nil {
		glog.Errorf("can't start gdb: %v\n", err)
		return
	}
}

@
@<Ждать завершения процесса@>=
{
	cmd.Wait()
}

@
@<Проверить возвращаемый результат@>=
{
	if !cmd.ProcessState.Success() {
		fmt.Fprintf(os.Stderr, "\n%s has finished with an error: %s\n", cmd.Path, cmd.ProcessState)
	}
}

@
@<Импортируемые пакеты@>=
"fmt"
"bufio"
"strings"

@
@<Типы@>=
request struct {
	pid	int
	out	io.WriteCloser
	cmd string
}

@
@<Глобальные переменные@>=
togdbch 	=make(chan interface{})
fromgdbch 	=make(chan string)
ackch		=make(chan bool)

@ Запускаем параллельные обработки ввода/вывода от \.{GDB} и организуем синхронное выполнение команд
@<Читать команды из  |stdin|, посылать их в \.{GDB}, обрабатывать результаты@>=
{
	@<Запустить параллельную обработку вывода из \.{GDB}@>
	@<Запустить параллельную обработку ввода из |stdin|@>
	rp:=strings.NewReplacer("\\n", "\n", "\\t", "\t", "\\\"", "\"")
	devnull,_:=os.Open(os.DevNull)
	var file io.WriteCloser=os.Stdout
	@<Подготовить синхронизацию выполнения команд@>
	loop: for true {
		select {
			case s, ok:=<-fromgdbch:
				if !ok {
					break loop
				}
				glog.V(debug).Infof("from gdb: '%s'", s)
				@<Обработка и отправка |s| в |file|@>
 			case v, ok:=<-togdbch:
				if !ok {
					break loop
				}
				switch r:=v.(type) {
					case request:
						glog.V(debug).Infof("to gdb from %d: '%s'", r.pid, r.cmd)
						file=r.out
						c:=strings.TrimSpace(r.cmd)
						if strings.HasPrefix(c, "-") {
							c=fmt.Sprintf("%d%s\n", r.pid, c)
						} else {
							c=fmt.Sprintf("%d-interpreter-exec console \"%s\"\n", r.pid, c)
						}
						io.WriteString(gdbin, c)
					case string:
						glog.V(debug).Infof("to gdb: '%s'", r)
						io.WriteString(gdbin, r+"\n")
				}
		}
	}
}

@
@<Импортируемые пакеты@>=
"unicode"
"strconv"
"bitbucket.org/santucco/gdbsh/common"

@ Сначала нужно проверить, нет ли в начале полученной строки идентификатора процесса, который сигнализирует о окончании выполнения команды.
Если идентификатор есть, отправляем результат в соответствующий процесс, переключаем текущий вывод на |stdout| и разрешаем выполнение следующей команды.
Если полученная строка содержит |"^running"|, то запрещаем ввод команд до появления строки |"*stopped"|.
Если идентификатора нет, строка выводится в соответствующий процесс или в |stdout|, причем в последнем случае она проходит предварительную обработку.
Приглашение |"(gdb)"|, получаемое от \.{GDB}, не выводится, поскольку за приглашение отвечае пакет |readline|. При получении первого приглашения мы разрешаем
ввод.
@<Обработка и отправка |s| в |file|@>=
{
	if len(s)==0 {
		continue
	}
	i:=0
	var r rune
	for i, r=range s {
		if !unicode.IsDigit(r) {
			break
		}
	}

	if p, err:=strconv.Atoi(s[:i]); err==nil {
		s=s[i:]
		if strings.HasPrefix(s, "^running") {
			@<Захватить ввод@>
		}
		glog.V(debug).Infof("writing to process %d': %s'", p, s)
		if n, err:=io.WriteString(file, s); err!=nil || n!=len(s) {
			glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s", s, n, err)
		}
		file=os.Stdout
		glog.Flush()
		@<Разрешение выполнения следующей команды@>
		continue
	}
	if strings.HasPrefix(s, "*stopped") {
		@<Разрешить ввод@>
	}
	@<Обработать строки для вывода в |os.Stdout|@>
	glog.V(debug).Infof("sending: '%s'", s)
	if n, err:=io.WriteString(file, s); err!=nil || n!=len(s) {
		glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s", s, n, err)
		file=devnull
	}

}

@
@<Обработать строки для вывода в |os.Stdout|@>=
{
	if file==os.Stdout{
		glog.V(debug).Infof("preprocessing for stdout: '%s'", s)
		switch s[0]{
			case '~', '&':
				s=s[2:len(s)-2]
			case '^':
				if strings.HasPrefix(s, "^error") {
					s=s[6:]
					if len(s)==0 || s[0]!=',' {
						continue
					}
					v, _, ok:=common.ParseResult(s[1:])
					glog.Errorf("%#v\n", v)
					if ok && len(v)!=0 && v[0].Name=="msg" {
						s=fmt.Sprintf("%s\n", v[0].Val.(string))
					}
				} else if strings.HasPrefix(s, "^done") {
					continue
				} else {
					continue
				}
				case '(':
					if strings.HasPrefix(s, "(gdb)") {
						@<Однократно выполнить операции инициализации при загруженном \.{GDB}@>
					}
					continue
			case '*':
				continue
			case '=':
				continue
		}
		s=rp.Replace(s)
	}
}

@
@<Запустить параллельную обработку вывода из \.{GDB}@>=
go func(){
	gdbr:=bufio.NewReader(gdbout)
	for s, err:=gdbr.ReadString('\n'); err==nil; s, err=gdbr.ReadString('\n') {
		glog.V(debug).Infof("'%s'", s)
		fromgdbch<-s
	}
	close(fromgdbch)
}()

@ Пакет |readline| осуществляет поддержку получения ввода с клавиатуры, историю и автозавершение команд.
Если введенная команда пустая, используется предыдущая команда, хранящаяся в |prev|
@<Запустить параллельную обработку ввода из |stdin|@>=
go func (){
	prev:=""
	@<Захватить ввод@>
	@<Создать экземпляр |readline| @>
	for	{
		s, err:=rl.Readline()
		@<Разрешить ввод@>
        if err!=nil { // io.EOF
                break
        }
		glog.V(debug).Infof("entered text: '%s'", s)
		if len(s)==0 {
			s=prev
		}
		var stdout io.WriteCloser = os.Stdout
		@<Запуск команд с выводом в |stdout|@>
		prev=s
		rl.SetPrompt("gdbsh$ ")
		@<Захватить ввод@>
	}
	glog.V(debug).Infof("on exit")
	togdbch<-"-gdb-exit"
} ()


@
@<Глобальные переменные@>=
cmds=map[string]string {
	@<Зарезервированные команды \.{GDB}@>
	@<Дополнительные встроенные команды@>
}

@ Запускаем конвейер команд, начиная с последней, затем ждем окончания всех команд и отправляем запросы на удаление из списка процессоров
@<Запуск команд с выводом в |stdout|@>=
{
	f:=func(r rune) bool { return r == '|' }
	cl:=FieldsFunc(strings.TrimSpace(s), f)
	glog.V(debug).Infof("commands: %#v", cl)
	var cnv []Cmd
	var toclose []io.Closer
	for i:=len(cl)-1; i>=0; i-- {
		first:=i==0
		c:=cl[i]
		@<Запустить команду |c| на выполнение и поместить ее в |cnv|@>
	}
	for _, cmd:=range cnv {
		if v, ok:=cmd.(*exec.Cmd); ok {
			glog.V(debug).Infof("waiting for process %s with pid %d is finished", v.Path, v.Process.Pid)
		}
		@<Ждать завершения процесса@>
		if v, ok:=cmd.(*exec.Cmd); ok {
			glog.V(debug).Infof("process %s with pid %d has finished", v.Path, v.Process.Pid)
		}
	}
}

@ Так как команды могу т быть как внутренними, так и внешними, создадим унифицированный интерфейс |Cmd|, подогнанный к функциям |exec.Cmd|.
Таким образом можно будет запускать и ожидаться окончания выполнения команд оинаковым образом
@<Типы@>=
Cmd interface {
	Start() error
	Wait() error
}

@
@<Запустить команду |c| на выполнение и поместить ее в |cnv|@>=
{
	var cmd Cmd
	var togdb io.ReadCloser
	var fromgdb io.WriteCloser
	@<Определить запускаемую команду и создать |cmd|@>
	@<Запустить |cmd| и добавить в список команд@>
}

@ Надо разделить команду на поля, взять первое и поискать его в |cmds|. Если команда не найдется, надо поискать путь до нее в |\$PATH|.
Если команда найдется с непустым путем или не найдется, ее нужно запустить как внешнюю команду.
@<Определить запускаемую команду и создать |cmd|@>=
{
	n:=strings.TrimSpace(c)
	if i:=strings.IndexFunc(n, unicode.IsSpace); i!=-1 {
		n=n[:i]
	}
	if _, ok:=cmds[n]; !ok {
		@<Создать внешний процесс@>
	} else {
		@<Создать внутреннюю команду@>
	}
}

@
@<Создать внешний процесс@>=
{
	var ar []string
	c=strings.Replace(c, "$", "\\$", -1)
	ar=append(ar, "sh", "-c", c)
	glog.V(debug).Infof("command arguments: %#v", ar)
	c:=exec.Command("/usr/bin/env", ar...)
	if c==nil {
		glog.Errorf("can't create command to run %s\n", n)
		break
	}
	cmd=c
}


@ Определим расширенную функцию разбиения на поля с учетом экранирования символов и неделимых строковых аргументов
@c
func FieldsFunc(s string, f func(rune) bool) []string {
	openeds:=false
	openedd:=false
	escaped:=false
	ff:=func(r rune) bool {
		if !openeds && !openedd && !escaped && f(r) {
			return true
		}
		if r=='\\' {
			escaped=!escaped
			return false
		}
		if r=='\'' && !escaped {
			openeds=!openeds
		}

		if r=='"' && !escaped {
			openedd=!openedd
		}
		escaped=false
		return false
	}
	return strings.FieldsFunc(s, ff)
}

@ Инициализируем |c.Stdout| предыдущим значением |stdout|. Если это не первая команда конвейера, то создаем канал для |c.Stdin|, по которому команда будет получать данные. Второй конец нового канала сохраняется в |stdout|
@<Заполнить |c.Stdin| и |c.Stdout| и сохранить в |stdout| второй конец канала@>=
{
	c.Stdout=stdout
	if !first {
		if out, in, err:=os.Pipe(); err!=nil {
			glog.Errorf("can't create pipe: %v\n", err)
			break
		} else {
			c.Stdin=out
			stdout=in
		}
	} else {
		c.Stdin=os.Stdin
	}
}

@ Если запускается конвейер, создаем для запускаемого процесса канал для связи с процессом - источником данных.
Также создаем два канала для взаимодействия процесса с \.{GDB} через \.{GDB/MI}.
Так как после запуска процесса дублирующие дескрипторы каналов должны быть закрыты, добавляем каналы в |toclose|
Канал |stdout| может быть каналом в памяти для обмена данными с внутренней командой, в этом случае его не нужно закрывать
@<Заполнить у |c| стандартные дескрипторы и два дополнительных для взаимодействия с \.{GDB}@>=
{
	@<Заполнить |c.Stdin| и |c.Stdout| и сохранить в |stdout| второй конец канала@>
	if _, ok:=c.Stdout.(*io.PipeWriter); !ok && c.Stdout!=os.Stdout {
		toclose=append(toclose, c.Stdout.(io.Closer))
	}
	if c.Stdin!=os.Stdin {
		toclose=append(toclose, c.Stdin.(io.Closer))
	}
	c.Stderr=os.Stderr
	var err error
	var r, w *os.File
	if r, fromgdb, err=os.Pipe(); err!=nil {
		glog.Errorf("can't create pipe: %v\n", err)
		break
	}
	if togdb, w, err=os.Pipe(); err!=nil {
		glog.Errorf("can't create pipe: %v\n", err)
		break
	}
	c.ExtraFiles = append(c.ExtraFiles, r, w)
	toclose=append(toclose, r, w)
}

@ Стартуем процесс или внутреннюю команду и параллельное считывание команд для \.{GDB}
@<Запустить |cmd| и добавить в список команд@>=
{
	switch c:=cmd.(type){
	case *exec.Cmd:
		@<Заполнить у |c| стандартные дескрипторы и два дополнительных для взаимодействия с \.{GDB}@>
	case *internal:
		@<Заполняем дескрипторы внутренней команды@>
	}

	if err:=cmd.Start(); err!=nil {
		glog.Errorf("can't start process: %s\n", err)
		break
	}
	@<Закрыть переданные дескрипторы@>

	go func() {
		var pid int
		if v, ok:=cmd.(*exec.Cmd); ok {
			pid=v.Process.Pid
		} else {
			pid=os.Getpid()
		}
		bufr:=bufio.NewReader(togdb)
		for s, err:=bufr.ReadString('\n'); err==nil || len(s)!=0; s, err=bufr.ReadString('\n') {
			glog.V(debug).Infof("%s has been recived from pid %d", s, pid)
			@<Ожидать разрешения выполнения следующей команды@>
			togdbch<-request{pid: pid, out: fromgdb, cmd: s}
		}
		glog.V(debug).Infof("end of input for pid %d", pid)
		togdb.Close()
		fromgdb.Close()
	} ()
	cnv=append(cnv, cmd)
}

@
@<Закрыть переданные дескрипторы@>=
{
	for _, p:=range toclose {
		p.Close()
	}
	toclose=nil
}

@
@<Импортируемые пакеты@>=
"os/signal"
"syscall"

@
@<Инициализация сигнальных обработчиков@>=
{
	sigch:=make(chan os.Signal, 10)
	defer signal.Stop(sigch)
	signal.Notify(sigch)
	go func() {
		for true {
			s, ok:=<-sigch
			if !ok {
				fmt.Fprintf(os.Stderr, "exit from handler\n")
				return
			}
			switch s {
				case syscall.SIGPIPE:
					glog.V(debug).Infof("signal SIGPIPE(%#v)", s)
					signal.Ignore(s)
				case os.Interrupt:
					glog.V(debug).Infof("signal SIGINT(%#v)", s)
					togdbch<-"-exec-interrupt"
				default:
					glog.V(debug).Infof("signal %#v", s)
			}
		}
	} ()

}

@ Если в качестве внутренней команды используется |args|, то читаем их |stdin| аргументы для идущей за |args| внутренней команды и запускаем внутреннюю команду с каждым считанным из |stdin| набором аргументов

@
@<Типы@>=
internal struct {
	cmd		string
	gdbin	io.ReadCloser
	gdbout	io.WriteCloser
	Stdin	io.ReadCloser
	Stdout	io.WriteCloser
	wait	chan bool
}

@ Метод |Start| для внутренней команды.
@c
func (this *internal)Start() error {
	go func () {
		defer this.gdbin.Close()
		defer this.gdbout.Close()
		if this.Stdout!=os.Stdout {
			defer this.Stdout.Close()
		}
		defer close(this.wait)
		defer func() {glog.V(debug).Infof("command %#v has done", this.cmd)}()
		c:=strings.TrimSpace(this.cmd)
		if i:=strings.IndexFunc(c, unicode.IsSpace); i!=-1 {
			c=c[:i]
		}

		if c=="args" {
			this.cmd=this.cmd[4:]
			stdr:=bufio.NewReader(this.Stdin)
			for s, err:=stdr.ReadString('\n'); err==nil; s, err=stdr.ReadString('\n') {
				cmd:=this.cmd + " " + strings.TrimSpace(s)
				@<Отправить команду |cmd| и обработать результат@>
			}
		} else {
			cmd:=this.cmd
			@<Отправить команду |cmd| и обработать результат@>
		}
		return
	} ()
	return nil
}

@ Метод |Stop| для внутренней команды.
@c
func (this *internal)Wait() error {
	glog.V(debug).Infof("waiting for internal command %#v is finished", this.cmd)
	<-this.wait
	glog.V(debug).Infof("internal command %#v has finished", this.cmd)
	return nil
}

@ Заполням поля внутренней команды.
@<Создать внутреннюю команду@>=
	var ci internal
	ci.wait=make(chan bool)
	ci.cmd=c
	cmd=&ci


@ Для |togdb| и |fromgdb| используем каналы в памяти.
|stdin| и |stdout| могут наполняться из внешних процессов, поэтому для них используются обычные каналы
@<Заполняем дескрипторы внутренней команды@>=
{
	c.gdbin, fromgdb=io.Pipe()
	togdb, c.gdbout=io.Pipe()
	@<Заполнить |c.Stdin| и |c.Stdout| и сохранить в |stdout| второй конец канала@>
}

@ Отправляем команду на выполнение в \.{GDB}, читаем вывод до появления результата с префиксом |\^|.
Печатаем обработанные результаты и возможные ошибки, остальной вывод игнорируется
@<Отправить команду |cmd| и обработать результат@>=
{
		glog.V(debug).Infof("internal command: %#v", cmd)
		if _, err:=io.WriteString(this.gdbout, cmd+"\n"); err!=nil {
			fmt.Fprintf(os.Stderr, "can't start gdb command '%s': %s\n", cmd, err)
			return
		}
		gdbr:=bufio.NewReader(this.gdbin)
		rp:=strings.NewReplacer("\\n", "\n", "\\t", "\t", "\\\"", "\"")
		quit:=false

		for s, err:=gdbr.ReadString('\n'); err==nil; s, err=gdbr.ReadString('\n') {
			glog.V(debug).Infof("sending: '%s'", s)

			if len(s) == 0 {
				continue
			}
			print:=true
			switch s[0]{
				case '~':
					s=s[2:len(s)-2]
					@<Если есть приглашение для вводе |">"|, отправить в |gdbin| ввод с терминала@>
				case '^':
					quit=true
					print=false
					if strings.HasPrefix(s, "^error") {
						s=s[6:]
						if len(s)==0 || s[0]!=',' {
							break
						}
						v, _, ok:=common.ParseResult(s[1:])
						if ok && len(v)!=0 && v[0].Name=="msg" {
							s=fmt.Sprintf("%s\n", v[0].Val.(string))
							print=true
						}
					} else if strings.HasPrefix(s, "^done") {
						s=s[5:]
						if len(s)==0 || s[0]!=',' {
							break
						}
						if v, _, ok:=common.ParseResult(s[1:]);ok {
							s=v.String()+"\n"
							print=true
						}
					}
				default:
					continue
			}
			if print {
				s=rp.Replace(s)
				if n, err:=io.WriteString(this.Stdout, s); err!=nil || n!=len(s) {
					glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s", s, n, err)
					return
				}
			}
			if quit {
				break
			}
		}
}

@ Команда \.{GDB} |commands| ждет ввода команд с терминала, будем читать ввод через |readline| и отправлять напрямую в |gdbin|
@<Если есть приглашение для вводе |">"|, отправить в |gdbin| ввод с терминала@>=
{
	if s1:=strings.TrimSpace(s); s1== ">" {
		rl.SetPrompt(s)
		s, err:=rl.Readline()
		if err!=nil {
			continue
		}
		glog.V(debug).Infof("entered text inside of internal command: '%s'", s)
		io.WriteString(gdbin, s+"\n")
		continue
	}
}



@
@<Импортируемые пакеты@>=
"github.com/golang/glog"

@
@<Глобальные переменные@>=
debug glog.Level=1

@
@<Подготовить трассировку@>=
{
	glog.V(debug).Infoln("main")
	defer glog.V(debug).Infoln("main is done")
	defer glog.Flush()
}

@
@<Импортируемые пакеты@>=
"sync"

@
@<Глобальные переменные@>=
ready	=make(chan bool, 1)
once sync.Once

@
@<Разрешить ввод@>=
{
	glog.V(debug).Infoln("an attempt to allow of input")
	ready<-true
	glog.V(debug).Infoln("input is allowed")
}

@
@<Однократно выполнить операции инициализации при загруженном \.{GDB}@>=
once.Do(func() {
	go func() {
		@<Заполнить автозаполнение и список зарезервированных команд@>
		@<Разрешить ввод@>
	}()
})

@
@<Захватить ввод@>=
{
	glog.V(debug).Infoln("an attempt to lock of input")
	<-ready
	glog.V(debug).Infoln("input is locked")
}

@
@<Глобальные переменные@>=
next	=make(chan bool, 1)

@
@<Подготовить синхронизацию выполнения команд@>=
next<-true

@
@<Разрешение выполнения следующей команды@>=
glog.V(debug).Infof("allow an execution of a next command")
next<-true

@
@<Ожидать разрешения выполнения следующей команды@>=
<-next
glog.V(debug).Infof("an execution of a next command is allowed")

@
@<Дополнительные встроенные команды@>=
"args": "",

@
@<Проверить аргументы командной строки, вывести информацию о программе, если необходимо@>=
{
	if len(os.Args)>1 && strings.TrimSpace(os.Args[1])=="-h" {
		fmt.Fprint(os.Stdout, "GDBSh 0.31, a shell for GDB\n",
			"Copyright (C) 2015, 2016 Alexander Sychev\n",
			"Usage:\n",
			"\tgdbsh <GDB options>\n",
			"GDBSh allows to use pipelines from GDB and external programs.\n",
			"A command 'args' allows to build and execute command lines with GDB commands from standart input.\n",
			"Special external programs can send GDB commands to GDB and obtain results.\n",
			"Two descriptors (3,4) are dedicated for an every external program for such purposes.\n")
		return
	}
}

@
@<Импортируемые пакеты@>=
"github.com/chzyer/readline"

@
@<Глобальные переменные@>=
pc []readline.PrefixCompleterInterface
rl *readline.Instance

@
@<Создать экземпляр |readline| @>=
	var err error
	rl, err=readline.NewEx(&readline.Config{
		Prompt:	"gdbsh$ ",
		AutoComplete: readline.NewPrefixCompleter(pc...),
		InterruptPrompt: "interrupt",
		EOFPrompt: "quit"})

	if err != nil {
		panic(err)
	}
	defer rl.Close()

@ С помощью рекурсивной функции |makePcItems| полученный список команд с подкомандами преобразуется в |readline.PrefixCompilerInterface|
Рекурсия нужна, чтобы сгруппировать подкоманды для одной команды в один |readline.PcItem|. Команда |"help"| игнорируется, поскольку она добавляется отдельно с
возможностью автозаполнения всеми остальными командами.
@c
func makePcItems (o [][]string, i int) (res []readline.PrefixCompleterInterface) {
	loop: for len(o)>0{
	 	if len(o[0])<=i || o[0][0]=="help" {
			o=o[1:]
			continue
		}
		s:=o[0][i]
		j:=1;
		for ; j<len(o); j++ {
			if len(o[j])>i && o[j][i]!=s {
				res=append(res, readline.PcItem(s, makePcItems(o[0:j], i+1)...))
				o=o[j:]
				continue loop
			}
		}
		res=append(res, readline.PcItem(s, makePcItems(o[0:j], i+1)...))
		o=o[j:]
	}
	return res
}

@ Создаем список для автозаполнения и лдополняем его командами |"help"| и |"args"| с автозаполнением всем перечнем команд.
@<Заполнить автозаполнение и список зарезервированных команд@>={
	var o [][]string
	@<Получить список команд@>
	for _, v:=range o {
		cmds[v[0]]=""
	}
	pc=makePcItems(o, 0)
	pc=append(pc, readline.PcItem("args", pc...))
	pc=append(pc, readline.PcItem("help", pc...))
}

@
@<Импортируемые пакеты@>=
"sort"

@ Получим скисок всех команд с помощью команды |"help all"|. Для этого запустим команду, а в качестве канала для вывода будем использовать канал в памяти,
из которого в отдельном потоке будет вычитываться весь вывод запущеной команды, фильтроваться и добавляться в массив строк. К сожалению, из-за ошибок в \.{GDB},
команды не всегда упорядочены, поэтому их приходится дополнительно отсортировать. Затем полученные команды разбиваюися на подкоманды для дальнейшей обработки.
@<Получить список команд@>=
{
	s:="help all"
	var stdout io.WriteCloser
	var gdbin io.ReadCloser
	gdbin, stdout=io.Pipe()

	ready:=make(chan bool)
	go func() {
		defer stdout.Close()
		defer gdbin.Close()
		defer close(ready)
		gdbr:=bufio.NewReader(gdbin)
		var ss []string
		for {
			s, err:=gdbr.ReadString('\n')
			if err!=nil {
				break
			}
			if i:=strings.Index(s, " --"); i!=-1 {
				ss=append(ss, s[:i])
			}
		}
		sort.Strings(ss)
		for _, v:=range ss {
			o=append(o, strings.Fields(v))
		}
	}()
	@<Запуск команд с выводом в |stdout|@>
	<-ready
}

@ Здесь определяются зарезервированные команды, которые должны иметь возможность запуститься при инициализации и которкие команжы, не описанные в \.{GDB}
@<Зарезервированные команды \.{GDB}@>=
"help": "",@#
"b": "",@#
"c": "",@#
"d": "",@#
"f": "",@#
"i": "",@#
"l": "",@#
"n": "",@#
"p": "",@#
"q": "",@#
"r": "",@#
"u": "",@#
"x": "",@#


@** Индекс.
