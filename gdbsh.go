

/*2:*/


//line gdbsh.w:18

//line license:1

// This file is part of GDBSh toolset
// Author Alexander Sychev
//
// Copyright (c) 2015, 2016, 2018, 2020 Alexander Sychev. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * The name of author may not be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//line gdbsh.w:20

package main

import(


/*3:*/


//line gdbsh.w:46

"os"
"os/exec"
"io"



/*:3*/



/*9:*/


//line gdbsh.w:113

"fmt"
"bufio"
"strings"



/*:9*/



/*13:*/


//line gdbsh.w:173

"unicode"
"strconv"
"github.com/santucco/gdbsh/common"



/*:13*/



/*29:*/


//line gdbsh.w:497

"os/signal"
"syscall"



/*:29*/



/*39:*/


//line gdbsh.w:680

"github.com/golang/glog"
"flag"



/*:39*/



/*42:*/


//line gdbsh.w:698

"sync"



/*:42*/



/*53:*/


//line gdbsh.w:770

"github.com/chzyer/readline"



/*:53*/



/*58:*/


//line gdbsh.w:830

"sort"



/*:58*/


//line gdbsh.w:24

)

type(


/*10:*/


//line gdbsh.w:119

request struct{
pid int
out io.WriteCloser
cmd string
}



/*:10*/



/*20:*/


//line gdbsh.w:333

Cmd interface{
Start()error
Wait()error
}



/*:20*/



/*32:*/


//line gdbsh.w:532

internal struct{
cmd string
gdbin io.ReadCloser
gdbout io.WriteCloser
Stdin io.ReadCloser
Stdout io.WriteCloser
wait chan bool
}



/*:32*/


//line gdbsh.w:28

)

var(


/*4:*/


//line gdbsh.w:52

gdbin io.WriteCloser
gdbout io.ReadCloser
gdberr io.ReadCloser
cmd*exec.Cmd



/*:4*/



/*11:*/


//line gdbsh.w:127

togdbch= make(chan interface{})
fromgdbch= make(chan string)
ackch= make(chan bool)



/*:11*/



/*18:*/


//line gdbsh.w:301

cmds= map[string]string{


/*60:*/


//line gdbsh.w:869

"help":"",
"b":"",
"c":"",
"d":"",
"f":"",
"i":"",
"l":"",
"n":"",
"p":"",
"q":"",
"r":"",
"u":"",
"x":"",




/*:60*/


//line gdbsh.w:303



/*51:*/


//line gdbsh.w:750

"args":"",



/*:51*/


//line gdbsh.w:304

}



/*:18*/



/*40:*/


//line gdbsh.w:685

debug glog.Level= 0



/*:40*/



/*43:*/


//line gdbsh.w:702

ready= make(chan bool,1)
once sync.Once



/*:43*/



/*47:*/


//line gdbsh.w:732

next= make(chan bool,1)



/*:47*/



/*54:*/


//line gdbsh.w:774

pc[]readline.PrefixCompleterInterface
rl*readline.Instance



/*:54*/


//line gdbsh.w:32

)

func main(){


/*52:*/


//line gdbsh.w:754

{
if len(os.Args)> 1&&strings.TrimSpace(os.Args[1])=="-h"{
fmt.Fprint(os.Stdout,"GDBSh 0.31, a shell for GDB\n",
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



/*:52*/


//line gdbsh.w:36



/*41:*/


//line gdbsh.w:689

{
flag.Parse()
glog.V(debug).Infoln("main")
defer glog.V(debug).Infoln("main is done")
defer glog.Flush()
}



/*:41*/


//line gdbsh.w:37



/*30:*/


//line gdbsh.w:502

{
sigch:=make(chan os.Signal,10)
defer signal.Stop(sigch)
signal.Notify(sigch)
go func(){
for true{
s,ok:=<-sigch
if!ok{
fmt.Fprintf(os.Stderr,"exit from handler\n")
return
}
switch s{
case syscall.SIGPIPE:
glog.V(debug).Infof("signal SIGPIPE(%#v)",s)
signal.Ignore(s)
case os.Interrupt:
glog.V(debug).Infof("signal SIGINT(%#v)",s)
togdbch<-"-exec-interrupt"
default:
glog.V(debug).Infof("signal %#v",s)
}
}
}()

}



/*:30*/


//line gdbsh.w:38



/*6:*/


//line gdbsh.w:65

{


/*5:*/


//line gdbsh.w:59

var args[]string
args= append(args,os.Args...)
args[0]= "--interpreter=mi"



/*:5*/


//line gdbsh.w:67

if cmd= exec.Command("gdb",args...);cmd==nil{
glog.Errorf("can't create command to run gdb\n")
return
}

var err error
if gdbin,err= cmd.StdinPipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
return
}
defer gdbin.Close()

if gdbout,err= cmd.StdoutPipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
return
}
defer gdbout.Close()

if gdberr,err= cmd.StderrPipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
return
}
defer gdberr.Close()

if err= cmd.Start();err!=nil{
glog.Errorf("can't start gdb: %v\n",err)
return
}
}



/*:6*/


//line gdbsh.w:39



/*12:*/


//line gdbsh.w:133

{


/*16:*/


//line gdbsh.w:262

go func(){
gdbr:=bufio.NewReader(gdbout)
for s,err:=gdbr.ReadString('\n');err==nil;s,err= gdbr.ReadString('\n'){
glog.V(debug).Infof("'%s'",s)
fromgdbch<-s
}
close(fromgdbch)
}()



/*:16*/


//line gdbsh.w:135



/*17:*/


//line gdbsh.w:274

go func(){
prev:=""


/*46:*/


//line gdbsh.w:724

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:46*/


//line gdbsh.w:277



/*55:*/


//line gdbsh.w:779

var err error
rl,err= readline.NewEx(&readline.Config{
Prompt:"gdbsh$ ",
AutoComplete:readline.NewPrefixCompleter(pc...),
InterruptPrompt:"interrupt",
EOFPrompt:"quit"})

if err!=nil{
panic(err)
}
defer rl.Close()



/*:55*/


//line gdbsh.w:278

for{
s,err:=rl.Readline()


/*44:*/


//line gdbsh.w:707

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:44*/


//line gdbsh.w:281

if err!=nil{// io.EOF
break
}
glog.V(debug).Infof("entered text: '%s'",s)
if len(s)==0{
s= prev
}
var stdout io.WriteCloser= os.Stdout


/*19:*/


//line gdbsh.w:308

{
f:=func(r rune)bool{return r=='|'}
cl:=FieldsFunc(strings.TrimSpace(s),f)
glog.V(debug).Infof("commands: %#v",cl)
var cnv[]Cmd
var toclose[]io.Closer
for i:=len(cl)-1;i>=0;i--{
first:=i==0
c:=cl[i]


/*21:*/


//line gdbsh.w:340

{
var cmd Cmd
var togdb io.ReadCloser
var fromgdb io.WriteCloser


/*22:*/


//line gdbsh.w:351

{
n:=strings.TrimSpace(c)
if i:=strings.IndexFunc(n,unicode.IsSpace);i!=-1{
n= n[:i]
}
if _,ok:=cmds[n];!ok{


/*23:*/


//line gdbsh.w:365

{
var ar[]string
ar= append(ar,"sh","-c",c)
glog.V(debug).Infof("command arguments: %#v",ar)
c:=exec.Command("/usr/bin/env",ar...)
if c==nil{
glog.Errorf("can't create command to run %s\n",n)
break
}
cmd= c
}




/*:23*/


//line gdbsh.w:358

}else{


/*35:*/


//line gdbsh.w:585

var ci internal
ci.wait= make(chan bool)
ci.cmd= c
cmd= &ci




/*:35*/


//line gdbsh.w:360

}
}



/*:22*/


//line gdbsh.w:345



/*27:*/


//line gdbsh.w:452

{
switch c:=cmd.(type){
case*exec.Cmd:


/*26:*/


//line gdbsh.w:427

{


/*25:*/


//line gdbsh.w:407

{
c.Stdout= stdout
if!first{
if out,in,err:=os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}else{
c.Stdin= out
stdout= in
}
}else{
c.Stdin= os.Stdin
}
}



/*:25*/


//line gdbsh.w:429

if _,ok:=c.Stdout.(*io.PipeWriter);!ok&&c.Stdout!=os.Stdout{
toclose= append(toclose,c.Stdout.(io.Closer))
}
if c.Stdin!=os.Stdin{
toclose= append(toclose,c.Stdin.(io.Closer))
}
c.Stderr= os.Stderr
var err error
var r,w*os.File
if r,fromgdb,err= os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}
if togdb,w,err= os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}
c.ExtraFiles= append(c.ExtraFiles,r,w)
toclose= append(toclose,r,w)
}



/*:26*/


//line gdbsh.w:456

case*internal:


/*36:*/


//line gdbsh.w:594

{
c.gdbin,fromgdb= io.Pipe()
togdb,c.gdbout= io.Pipe()


/*25:*/


//line gdbsh.w:407

{
c.Stdout= stdout
if!first{
if out,in,err:=os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}else{
c.Stdin= out
stdout= in
}
}else{
c.Stdin= os.Stdin
}
}



/*:25*/


//line gdbsh.w:598

}



/*:36*/


//line gdbsh.w:458

}

if err:=cmd.Start();err!=nil{
glog.Errorf("can't start process: %s\n",err)
break
}


/*28:*/


//line gdbsh.w:488

{
for _,p:=range toclose{
p.Close()
}
toclose= nil
}



/*:28*/


//line gdbsh.w:465


go func(){
var pid int
if v,ok:=cmd.(*exec.Cmd);ok{
pid= v.Process.Pid
}else{
pid= os.Getpid()
}
bufr:=bufio.NewReader(togdb)
for s,err:=bufr.ReadString('\n');err==nil||len(s)!=0;s,err= bufr.ReadString('\n'){
glog.V(debug).Infof("%s has been recived from pid %d",s,pid)


/*50:*/


//line gdbsh.w:745

<-next
glog.V(debug).Infof("an execution of a next command is allowed")



/*:50*/


//line gdbsh.w:477

togdbch<-request{pid:pid,out:fromgdb,cmd:s}
}
glog.V(debug).Infof("end of input for pid %d",pid)
togdb.Close()
fromgdb.Close()
}()
cnv= append(cnv,cmd)
}



/*:27*/


//line gdbsh.w:346

}



/*:21*/


//line gdbsh.w:318

}
for _,cmd:=range cnv{
if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("waiting for process %s with pid %d is finished",v.Path,v.Process.Pid)
}


/*7:*/


//line gdbsh.w:99

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:324

if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("process %s with pid %d has finished",v.Path,v.Process.Pid)
}
}
}



/*:19*/


//line gdbsh.w:290

prev= s
rl.SetPrompt("gdbsh$ ")


/*46:*/


//line gdbsh.w:724

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:46*/


//line gdbsh.w:293

}
glog.V(debug).Infof("on exit")
togdbch<-"-gdb-exit"
}()




/*:17*/


//line gdbsh.w:136

rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")
devnull,_:=os.Open(os.DevNull)
var file io.WriteCloser= os.Stdout


/*48:*/


//line gdbsh.w:736

next<-true



/*:48*/


//line gdbsh.w:140

loop:for true{
select{
case s,ok:=<-fromgdbch:
if!ok{
break loop
}
glog.V(debug).Infof("from gdb: '%s'",s)


/*14:*/


//line gdbsh.w:184

{
if len(s)==0{
continue
}
i:=0
var r rune
for i,r= range s{
if!unicode.IsDigit(r){
break
}
}

if p,err:=strconv.Atoi(s[:i]);err==nil{
s= s[i:]
if strings.HasPrefix(s,"^running"){


/*46:*/


//line gdbsh.w:724

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:46*/


//line gdbsh.w:200

}
glog.V(debug).Infof("writing to process %d': %s'",p,s)
if n,err:=io.WriteString(file,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s",s,n,err)
}
file= os.Stdout
glog.Flush()


/*49:*/


//line gdbsh.w:740

glog.V(debug).Infof("allow an execution of a next command")
next<-true



/*:49*/


//line gdbsh.w:208

continue
}
if strings.HasPrefix(s,"*stopped"){


/*44:*/


//line gdbsh.w:707

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:44*/


//line gdbsh.w:212

}


/*15:*/


//line gdbsh.w:224

{
if file==os.Stdout{
glog.V(debug).Infof("preprocessing for stdout: '%s'",s)
switch s[0]{
case'~','&':
s= s[2:len(s)-2]
case'^':
if strings.HasPrefix(s,"^error"){
s= s[6:]
if len(s)==0||s[0]!=','{
continue
}
v,_,ok:=common.ParseResult(s[1:])
glog.Errorf("%#v\n",v)
if ok&&len(v)!=0&&v[0].Name=="msg"{
s= fmt.Sprintf("%s\n",v[0].Val.(string))
}
}else if strings.HasPrefix(s,"^done"){
continue
}else{
continue
}
case'(':
if strings.HasPrefix(s,"(gdb)"){


/*45:*/


//line gdbsh.w:715

once.Do(func(){
go func(){


/*57:*/


//line gdbsh.w:818
{
var o[][]string


/*59:*/


//line gdbsh.w:836

{
s:="help all"
var stdout io.WriteCloser
var gdbin io.ReadCloser
gdbin,stdout= io.Pipe()

ready:=make(chan bool)
go func(){
defer stdout.Close()
defer gdbin.Close()
defer close(ready)
gdbr:=bufio.NewReader(gdbin)
var ss[]string
for{
s,err:=gdbr.ReadString('\n')
if err!=nil{
break
}
if i:=strings.Index(s," --");i!=-1{
ss= append(ss,s[:i])
}
}
sort.Strings(ss)
for _,v:=range ss{
o= append(o,strings.Fields(v))
}
}()


/*19:*/


//line gdbsh.w:308

{
f:=func(r rune)bool{return r=='|'}
cl:=FieldsFunc(strings.TrimSpace(s),f)
glog.V(debug).Infof("commands: %#v",cl)
var cnv[]Cmd
var toclose[]io.Closer
for i:=len(cl)-1;i>=0;i--{
first:=i==0
c:=cl[i]


/*21:*/


//line gdbsh.w:340

{
var cmd Cmd
var togdb io.ReadCloser
var fromgdb io.WriteCloser


/*22:*/


//line gdbsh.w:351

{
n:=strings.TrimSpace(c)
if i:=strings.IndexFunc(n,unicode.IsSpace);i!=-1{
n= n[:i]
}
if _,ok:=cmds[n];!ok{


/*23:*/


//line gdbsh.w:365

{
var ar[]string
ar= append(ar,"sh","-c",c)
glog.V(debug).Infof("command arguments: %#v",ar)
c:=exec.Command("/usr/bin/env",ar...)
if c==nil{
glog.Errorf("can't create command to run %s\n",n)
break
}
cmd= c
}




/*:23*/


//line gdbsh.w:358

}else{


/*35:*/


//line gdbsh.w:585

var ci internal
ci.wait= make(chan bool)
ci.cmd= c
cmd= &ci




/*:35*/


//line gdbsh.w:360

}
}



/*:22*/


//line gdbsh.w:345



/*27:*/


//line gdbsh.w:452

{
switch c:=cmd.(type){
case*exec.Cmd:


/*26:*/


//line gdbsh.w:427

{


/*25:*/


//line gdbsh.w:407

{
c.Stdout= stdout
if!first{
if out,in,err:=os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}else{
c.Stdin= out
stdout= in
}
}else{
c.Stdin= os.Stdin
}
}



/*:25*/


//line gdbsh.w:429

if _,ok:=c.Stdout.(*io.PipeWriter);!ok&&c.Stdout!=os.Stdout{
toclose= append(toclose,c.Stdout.(io.Closer))
}
if c.Stdin!=os.Stdin{
toclose= append(toclose,c.Stdin.(io.Closer))
}
c.Stderr= os.Stderr
var err error
var r,w*os.File
if r,fromgdb,err= os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}
if togdb,w,err= os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}
c.ExtraFiles= append(c.ExtraFiles,r,w)
toclose= append(toclose,r,w)
}



/*:26*/


//line gdbsh.w:456

case*internal:


/*36:*/


//line gdbsh.w:594

{
c.gdbin,fromgdb= io.Pipe()
togdb,c.gdbout= io.Pipe()


/*25:*/


//line gdbsh.w:407

{
c.Stdout= stdout
if!first{
if out,in,err:=os.Pipe();err!=nil{
glog.Errorf("can't create pipe: %v\n",err)
break
}else{
c.Stdin= out
stdout= in
}
}else{
c.Stdin= os.Stdin
}
}



/*:25*/


//line gdbsh.w:598

}



/*:36*/


//line gdbsh.w:458

}

if err:=cmd.Start();err!=nil{
glog.Errorf("can't start process: %s\n",err)
break
}


/*28:*/


//line gdbsh.w:488

{
for _,p:=range toclose{
p.Close()
}
toclose= nil
}



/*:28*/


//line gdbsh.w:465


go func(){
var pid int
if v,ok:=cmd.(*exec.Cmd);ok{
pid= v.Process.Pid
}else{
pid= os.Getpid()
}
bufr:=bufio.NewReader(togdb)
for s,err:=bufr.ReadString('\n');err==nil||len(s)!=0;s,err= bufr.ReadString('\n'){
glog.V(debug).Infof("%s has been recived from pid %d",s,pid)


/*50:*/


//line gdbsh.w:745

<-next
glog.V(debug).Infof("an execution of a next command is allowed")



/*:50*/


//line gdbsh.w:477

togdbch<-request{pid:pid,out:fromgdb,cmd:s}
}
glog.V(debug).Infof("end of input for pid %d",pid)
togdb.Close()
fromgdb.Close()
}()
cnv= append(cnv,cmd)
}



/*:27*/


//line gdbsh.w:346

}



/*:21*/


//line gdbsh.w:318

}
for _,cmd:=range cnv{
if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("waiting for process %s with pid %d is finished",v.Path,v.Process.Pid)
}


/*7:*/


//line gdbsh.w:99

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:324

if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("process %s with pid %d has finished",v.Path,v.Process.Pid)
}
}
}



/*:19*/


//line gdbsh.w:864

<-ready
}



/*:59*/


//line gdbsh.w:820

for _,v:=range o{
cmds[v[0]]= ""
}
pc= makePcItems(o,0)
pc= append(pc,readline.PcItem("args",pc...))
pc= append(pc,readline.PcItem("help",pc...))
}



/*:57*/


//line gdbsh.w:718



/*44:*/


//line gdbsh.w:707

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:44*/


//line gdbsh.w:719

}()
})



/*:45*/


//line gdbsh.w:249

}
continue
case'*':
continue
case'=':
continue
}
s= rp.Replace(s)
}
}



/*:15*/


//line gdbsh.w:214

glog.V(debug).Infof("sending: '%s'",s)
if n,err:=io.WriteString(file,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s",s,n,err)
file= devnull
}

}



/*:14*/


//line gdbsh.w:148

case v,ok:=<-togdbch:
if!ok{
break loop
}
switch r:=v.(type){
case request:
glog.V(debug).Infof("to gdb from %d: '%s'",r.pid,r.cmd)
file= r.out
c:=strings.TrimSpace(r.cmd)
if strings.HasPrefix(c,"-"){
c= fmt.Sprintf("%d%s\n",r.pid,c)
}else{
c= fmt.Sprintf("%d-interpreter-exec console \"%s\"\n",r.pid,c)
}
io.WriteString(gdbin,c)
case string:
glog.V(debug).Infof("to gdb: '%s'",r)
io.WriteString(gdbin,r+"\n")
}
}
}
}



/*:12*/


//line gdbsh.w:40



/*7:*/


//line gdbsh.w:99

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:41



/*8:*/


//line gdbsh.w:105

{
if!cmd.ProcessState.Success(){
fmt.Fprintf(os.Stderr,"\n%s has finished with an error: %s\n",cmd.Path,cmd.ProcessState)
}
}



/*:8*/


//line gdbsh.w:42

}



/*:2*/



/*24:*/


//line gdbsh.w:380

func FieldsFunc(s string,f func(rune)bool)[]string{
openeds:=false
openedd:=false
escaped:=false
ff:=func(r rune)bool{
if!openeds&&!openedd&&!escaped&&f(r){
return true
}
if r=='\\'{
escaped= !escaped
return false
}
if r=='\''&&!escaped{
openeds= !openeds
}

if r=='"'&&!escaped{
openedd= !openedd
}
escaped= false
return false
}
return strings.FieldsFunc(s,ff)
}



/*:24*/



/*33:*/


//line gdbsh.w:543

func(this*internal)Start()error{
go func(){
defer this.gdbin.Close()
defer this.gdbout.Close()
if this.Stdout!=os.Stdout{
defer this.Stdout.Close()
}
defer close(this.wait)
defer func(){glog.V(debug).Infof("command %#v has done",this.cmd)}()
this.cmd= strings.TrimLeftFunc(this.cmd,unicode.IsSpace)
var c string
if i:=strings.IndexFunc(this.cmd,unicode.IsSpace);i!=-1{
c= this.cmd[:i]
}

if c=="args"{
this.cmd= this.cmd[len(c):]
stdr:=bufio.NewReader(this.Stdin)
for s,err:=stdr.ReadString('\n');err==nil;s,err= stdr.ReadString('\n'){
cmd:=this.cmd+" "+strings.TrimSpace(s)


/*37:*/


//line gdbsh.w:603

{
glog.V(debug).Infof("internal command: %#v",cmd)
if _,err:=io.WriteString(this.gdbout,cmd+"\n");err!=nil{
fmt.Fprintf(os.Stderr,"can't start gdb command '%s': %s\n",cmd,err)
return
}
gdbr:=bufio.NewReader(this.gdbin)
rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")
quit:=false

for s,err:=gdbr.ReadString('\n');err==nil;s,err= gdbr.ReadString('\n'){
glog.V(debug).Infof("sending: '%s'",s)

if len(s)==0{
continue
}
print:=true
switch s[0]{
case'~':
s= s[2:len(s)-2]


/*38:*/


//line gdbsh.w:665

{
if s1:=strings.TrimSpace(s);s1==">"{
rl.SetPrompt(s)
s,err:=rl.Readline()
if err!=nil{
continue
}
glog.V(debug).Infof("entered text inside of internal command: '%s'",s)
io.WriteString(gdbin,s+"\n")
continue
}
}



/*:38*/


//line gdbsh.w:624

case'^':
quit= true
print= false
if strings.HasPrefix(s,"^error"){
s= s[6:]
if len(s)==0||s[0]!=','{
break
}
v,_,ok:=common.ParseResult(s[1:])
if ok&&len(v)!=0&&v[0].Name=="msg"{
s= fmt.Sprintf("%s\n",v[0].Val.(string))
print= true
}
}else if strings.HasPrefix(s,"^done"){
s= s[5:]
if len(s)==0||s[0]!=','{
break
}
if v,_,ok:=common.ParseResult(s[1:]);ok{
s= v.String()+"\n"
print= true
}
}
default:
continue
}
if print{
s= rp.Replace(s)
if n,err:=io.WriteString(this.Stdout,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",s,n,err)
return
}
}
if quit{
break
}
}
}



/*:37*/


//line gdbsh.w:564

}
}else{
cmd:=this.cmd


/*37:*/


//line gdbsh.w:603

{
glog.V(debug).Infof("internal command: %#v",cmd)
if _,err:=io.WriteString(this.gdbout,cmd+"\n");err!=nil{
fmt.Fprintf(os.Stderr,"can't start gdb command '%s': %s\n",cmd,err)
return
}
gdbr:=bufio.NewReader(this.gdbin)
rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")
quit:=false

for s,err:=gdbr.ReadString('\n');err==nil;s,err= gdbr.ReadString('\n'){
glog.V(debug).Infof("sending: '%s'",s)

if len(s)==0{
continue
}
print:=true
switch s[0]{
case'~':
s= s[2:len(s)-2]


/*38:*/


//line gdbsh.w:665

{
if s1:=strings.TrimSpace(s);s1==">"{
rl.SetPrompt(s)
s,err:=rl.Readline()
if err!=nil{
continue
}
glog.V(debug).Infof("entered text inside of internal command: '%s'",s)
io.WriteString(gdbin,s+"\n")
continue
}
}



/*:38*/


//line gdbsh.w:624

case'^':
quit= true
print= false
if strings.HasPrefix(s,"^error"){
s= s[6:]
if len(s)==0||s[0]!=','{
break
}
v,_,ok:=common.ParseResult(s[1:])
if ok&&len(v)!=0&&v[0].Name=="msg"{
s= fmt.Sprintf("%s\n",v[0].Val.(string))
print= true
}
}else if strings.HasPrefix(s,"^done"){
s= s[5:]
if len(s)==0||s[0]!=','{
break
}
if v,_,ok:=common.ParseResult(s[1:]);ok{
s= v.String()+"\n"
print= true
}
}
default:
continue
}
if print{
s= rp.Replace(s)
if n,err:=io.WriteString(this.Stdout,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",s,n,err)
return
}
}
if quit{
break
}
}
}



/*:37*/


//line gdbsh.w:568

}
return
}()
return nil
}



/*:33*/



/*34:*/


//line gdbsh.w:576

func(this*internal)Wait()error{
glog.V(debug).Infof("waiting for internal command %#v is finished",this.cmd)
<-this.wait
glog.V(debug).Infof("internal command %#v has finished",this.cmd)
return nil
}



/*:34*/



/*56:*/


//line gdbsh.w:795

func makePcItems(o[][]string,i int)(res[]readline.PrefixCompleterInterface){
loop:for len(o)> 0{
if len(o[0])<=i||o[0][0]=="help"{
o= o[1:]
continue
}
s:=o[0][i]
j:=1;
for;j<len(o);j++{
if len(o[j])> i&&o[j][i]!=s{
res= append(res,readline.PcItem(s,makePcItems(o[0:j],i+1)...))
o= o[j:]
continue loop
}
}
res= append(res,readline.PcItem(s,makePcItems(o[0:j],i+1)...))
o= o[j:]
}
return res
}



/*:56*/


