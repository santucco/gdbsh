

/*2:*/


//line gdbsh.w:26

//line license:1

// This file is part of GDBSh toolset
// Author Alexander Sychev
//
// Copyright (c) 2015, 2016, 2018, 2020, 2023 Alexander Sychev. All rights reserved.
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
//line gdbsh.w:28

package main

import(


/*3:*/


//line gdbsh.w:54

"os"
"os/exec"
"io"



/*:3*/



/*9:*/


//line gdbsh.w:127

"fmt"
"bufio"
"strings"



/*:9*/



/*13:*/


//line gdbsh.w:187

"unicode"
"strconv"
"github.com/santucco/gdbsh/common"



/*:13*/



/*29:*/


//line gdbsh.w:511

"os/signal"
"syscall"



/*:29*/



/*40:*/


//line gdbsh.w:709

"github.com/golang/glog"
"flag"



/*:40*/



/*43:*/


//line gdbsh.w:728

"sync"



/*:43*/



/*54:*/


//line gdbsh.w:800

"github.com/chzyer/readline"



/*:54*/



/*59:*/


//line gdbsh.w:860

"sort"



/*:59*/


//line gdbsh.w:32

)

type(


/*10:*/


//line gdbsh.w:133

request struct{
pid int
out io.WriteCloser
cmd string
}



/*:10*/



/*20:*/


//line gdbsh.w:347

Cmd interface{
Start()error
Wait()error
}



/*:20*/



/*32:*/


//line gdbsh.w:547

internal struct{
cmd string
gdbin io.ReadCloser
gdbout io.WriteCloser
Stdin io.ReadCloser
Stdout io.WriteCloser
wait chan bool
}



/*:32*/


//line gdbsh.w:36

)

var(


/*4:*/


//line gdbsh.w:60

gdbin io.WriteCloser
gdbout io.ReadCloser
gdberr io.ReadCloser
cmd*exec.Cmd



/*:4*/



/*11:*/


//line gdbsh.w:141

togdbch= make(chan interface{})
fromgdbch= make(chan string)
ackch= make(chan bool)



/*:11*/



/*18:*/


//line gdbsh.w:315

cmds= map[string]string{


/*61:*/


//line gdbsh.w:901

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




/*:61*/


//line gdbsh.w:317



/*52:*/


//line gdbsh.w:780

"args":"",



/*:52*/


//line gdbsh.w:318

}



/*:18*/



/*41:*/


//line gdbsh.w:714

debug glog.Level= 0



/*:41*/



/*44:*/


//line gdbsh.w:732

ready= make(chan bool,1)
once sync.Once



/*:44*/



/*48:*/


//line gdbsh.w:762

next= make(chan bool,1)



/*:48*/



/*55:*/


//line gdbsh.w:804

pc[]readline.PrefixCompleterInterface
rl*readline.Instance



/*:55*/


//line gdbsh.w:40

)

func main(){


/*53:*/


//line gdbsh.w:784

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



/*:53*/


//line gdbsh.w:44



/*42:*/


//line gdbsh.w:718

{
flag.CommandLine= flag.NewFlagSet(os.Args[0],flag.ContinueOnError)
flag.Parse()
glog.V(debug).Infoln("main")
defer glog.V(debug).Infoln("main is done")
defer glog.Flush()
}



/*:42*/


//line gdbsh.w:45



/*30:*/


//line gdbsh.w:516

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
signal.Ignore(s)
togdbch<-"-exec-interrupt"
default:
glog.V(debug).Infof("signal %#v",s)
}
}
}()

}



/*:30*/


//line gdbsh.w:46



/*6:*/


//line gdbsh.w:79

{


/*5:*/


//line gdbsh.w:67

var args[]string
args= append(args,"--interpreter=mi")
for i,v:=range os.Args{
if i==0{
continue
}
fmt.Fprintf(os.Stderr,"%s\n",v)
args= append(args,v);
}



/*:5*/


//line gdbsh.w:81

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


//line gdbsh.w:47



/*12:*/


//line gdbsh.w:147

{


/*16:*/


//line gdbsh.w:276

go func(){
gdbr:=bufio.NewReader(gdbout)
for s,err:=gdbr.ReadString('\n');err==nil;s,err= gdbr.ReadString('\n'){
glog.V(debug).Infof("'%s'",s)
fromgdbch<-s
}
close(fromgdbch)
}()



/*:16*/


//line gdbsh.w:149



/*17:*/


//line gdbsh.w:288

go func(){
prev:=""


/*47:*/


//line gdbsh.w:754

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:47*/


//line gdbsh.w:291



/*56:*/


//line gdbsh.w:809

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



/*:56*/


//line gdbsh.w:292

for{
s,err:=rl.Readline()


/*45:*/


//line gdbsh.w:737

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:45*/


//line gdbsh.w:295

if err!=nil{// io.EOF
break
}
glog.V(debug).Infof("entered text: '%s'",s)
if len(s)==0{
s= prev
}
var stdout io.WriteCloser= os.Stdout


/*19:*/


//line gdbsh.w:322

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


//line gdbsh.w:354

{
var cmd Cmd
var togdb io.ReadCloser
var fromgdb io.WriteCloser


/*22:*/


//line gdbsh.w:365

{
n:=strings.TrimSpace(c)
if i:=strings.IndexFunc(n,unicode.IsSpace);i!=-1{
n= n[:i]
}
if _,ok:=cmds[n];!ok{


/*23:*/


//line gdbsh.w:379

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


//line gdbsh.w:372

}else{


/*36:*/


//line gdbsh.w:614

var ci internal
ci.wait= make(chan bool)
ci.cmd= c
cmd= &ci




/*:36*/


//line gdbsh.w:374

}
}



/*:22*/


//line gdbsh.w:359



/*27:*/


//line gdbsh.w:466

{
switch c:=cmd.(type){
case*exec.Cmd:


/*26:*/


//line gdbsh.w:441

{


/*25:*/


//line gdbsh.w:421

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


//line gdbsh.w:443

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


//line gdbsh.w:470

case*internal:


/*37:*/


//line gdbsh.w:623

{
c.gdbin,fromgdb= io.Pipe()
togdb,c.gdbout= io.Pipe()


/*25:*/


//line gdbsh.w:421

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


//line gdbsh.w:627

}



/*:37*/


//line gdbsh.w:472

}

if err:=cmd.Start();err!=nil{
glog.Errorf("can't start process: %s\n",err)
break
}


/*28:*/


//line gdbsh.w:502

{
for _,p:=range toclose{
p.Close()
}
toclose= nil
}



/*:28*/


//line gdbsh.w:479


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


/*51:*/


//line gdbsh.w:775

<-next
glog.V(debug).Infof("an execution of a next command is allowed")



/*:51*/


//line gdbsh.w:491

togdbch<-request{pid:pid,out:fromgdb,cmd:s}
}
glog.V(debug).Infof("end of input for pid %d",pid)
togdb.Close()
fromgdb.Close()
}()
cnv= append(cnv,cmd)
}



/*:27*/


//line gdbsh.w:360

}



/*:21*/


//line gdbsh.w:332

}
for _,cmd:=range cnv{
if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("waiting for process %s with pid %d is finished",v.Path,v.Process.Pid)
}


/*7:*/


//line gdbsh.w:113

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:338

if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("process %s with pid %d has finished",v.Path,v.Process.Pid)
}
}
}



/*:19*/


//line gdbsh.w:304

prev= s
rl.SetPrompt("gdbsh$ ")


/*47:*/


//line gdbsh.w:754

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:47*/


//line gdbsh.w:307

}
glog.V(debug).Infof("on exit")
togdbch<-"-gdb-exit"
}()




/*:17*/


//line gdbsh.w:150

rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")
devnull,_:=os.Open(os.DevNull)
var file io.WriteCloser= os.Stdout


/*49:*/


//line gdbsh.w:766

next<-true



/*:49*/


//line gdbsh.w:154

loop:for true{
select{
case s,ok:=<-fromgdbch:
if!ok{
break loop
}
glog.V(debug).Infof("from gdb: '%s'",s)


/*14:*/


//line gdbsh.w:198

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


/*47:*/


//line gdbsh.w:754

{
glog.V(debug).Infoln("an attempt to lock of input")
<-ready
glog.V(debug).Infoln("input is locked")
}



/*:47*/


//line gdbsh.w:214

}
glog.V(debug).Infof("writing to process %d': %s'",p,s)
if n,err:=io.WriteString(file,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s",s,n,err)
}
file= os.Stdout
glog.Flush()


/*50:*/


//line gdbsh.w:770

glog.V(debug).Infof("allow an execution of a next command")
next<-true



/*:50*/


//line gdbsh.w:222

continue
}
if strings.HasPrefix(s,"*stopped"){


/*45:*/


//line gdbsh.w:737

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:45*/


//line gdbsh.w:226

}


/*15:*/


//line gdbsh.w:238

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


/*46:*/


//line gdbsh.w:745

once.Do(func(){
go func(){


/*58:*/


//line gdbsh.w:848
{
var o[][]string


/*60:*/


//line gdbsh.w:866

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
for _,v:=range strings.Split(s[:i],","){
ss= append(ss,strings.Trim(v," "))
}
}
}
sort.Strings(ss)
for _,v:=range ss{
o= append(o,strings.Fields(v))
}
}()


/*19:*/


//line gdbsh.w:322

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


//line gdbsh.w:354

{
var cmd Cmd
var togdb io.ReadCloser
var fromgdb io.WriteCloser


/*22:*/


//line gdbsh.w:365

{
n:=strings.TrimSpace(c)
if i:=strings.IndexFunc(n,unicode.IsSpace);i!=-1{
n= n[:i]
}
if _,ok:=cmds[n];!ok{


/*23:*/


//line gdbsh.w:379

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


//line gdbsh.w:372

}else{


/*36:*/


//line gdbsh.w:614

var ci internal
ci.wait= make(chan bool)
ci.cmd= c
cmd= &ci




/*:36*/


//line gdbsh.w:374

}
}



/*:22*/


//line gdbsh.w:359



/*27:*/


//line gdbsh.w:466

{
switch c:=cmd.(type){
case*exec.Cmd:


/*26:*/


//line gdbsh.w:441

{


/*25:*/


//line gdbsh.w:421

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


//line gdbsh.w:443

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


//line gdbsh.w:470

case*internal:


/*37:*/


//line gdbsh.w:623

{
c.gdbin,fromgdb= io.Pipe()
togdb,c.gdbout= io.Pipe()


/*25:*/


//line gdbsh.w:421

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


//line gdbsh.w:627

}



/*:37*/


//line gdbsh.w:472

}

if err:=cmd.Start();err!=nil{
glog.Errorf("can't start process: %s\n",err)
break
}


/*28:*/


//line gdbsh.w:502

{
for _,p:=range toclose{
p.Close()
}
toclose= nil
}



/*:28*/


//line gdbsh.w:479


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


/*51:*/


//line gdbsh.w:775

<-next
glog.V(debug).Infof("an execution of a next command is allowed")



/*:51*/


//line gdbsh.w:491

togdbch<-request{pid:pid,out:fromgdb,cmd:s}
}
glog.V(debug).Infof("end of input for pid %d",pid)
togdb.Close()
fromgdb.Close()
}()
cnv= append(cnv,cmd)
}



/*:27*/


//line gdbsh.w:360

}



/*:21*/


//line gdbsh.w:332

}
for _,cmd:=range cnv{
if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("waiting for process %s with pid %d is finished",v.Path,v.Process.Pid)
}


/*7:*/


//line gdbsh.w:113

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:338

if v,ok:=cmd.(*exec.Cmd);ok{
glog.V(debug).Infof("process %s with pid %d has finished",v.Path,v.Process.Pid)
}
}
}



/*:19*/


//line gdbsh.w:896

<-ready
}



/*:60*/


//line gdbsh.w:850

for _,v:=range o{
cmds[v[0]]= ""
}
pc= makePcItems(o,0)
pc= append(pc,readline.PcItem("args",pc...))
pc= append(pc,readline.PcItem("help",pc...))
}



/*:58*/


//line gdbsh.w:748



/*45:*/


//line gdbsh.w:737

{
glog.V(debug).Infoln("an attempt to allow of input")
ready<-true
glog.V(debug).Infoln("input is allowed")
}



/*:45*/


//line gdbsh.w:749

}()
})



/*:46*/


//line gdbsh.w:263

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


//line gdbsh.w:228

glog.V(debug).Infof("sending: '%s'",s)
if n,err:=io.WriteString(file,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to output, %d bytes has been written: %s",s,n,err)
file= devnull
}

}



/*:14*/


//line gdbsh.w:162

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


//line gdbsh.w:48



/*7:*/


//line gdbsh.w:113

{
cmd.Wait()
}



/*:7*/


//line gdbsh.w:49



/*8:*/


//line gdbsh.w:119

{
if!cmd.ProcessState.Success(){
fmt.Fprintf(os.Stderr,"\n%s has finished with an error: %s\n",cmd.Path,cmd.ProcessState)
}
}



/*:8*/


//line gdbsh.w:50

}



/*:2*/



/*24:*/


//line gdbsh.w:394

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


//line gdbsh.w:558

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


/*34:*/


//line gdbsh.w:594

s= strings.TrimSpace(s)
args:=FieldsFunc(s,unicode.IsSpace)
cmd:=this.cmd
for i:=len(args);i> 0;i--{
cmd= strings.Replace(cmd,"$"+strconv.Itoa(i),args[i-1],-1)
}
cmd= strings.Replace(cmd,"$0",s,-1)




/*:34*/


//line gdbsh.w:578



/*38:*/


//line gdbsh.w:632

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


/*39:*/


//line gdbsh.w:694

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



/*:39*/


//line gdbsh.w:653

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



/*:38*/


//line gdbsh.w:579

}
}else{
cmd:=this.cmd


/*38:*/


//line gdbsh.w:632

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


/*39:*/


//line gdbsh.w:694

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



/*:39*/


//line gdbsh.w:653

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



/*:38*/


//line gdbsh.w:583

}
return
}()
return nil
}



/*:33*/



/*35:*/


//line gdbsh.w:605

func(this*internal)Wait()error{
glog.V(debug).Infof("waiting for internal command %#v is finished",this.cmd)
<-this.wait
glog.V(debug).Infof("internal command %#v has finished",this.cmd)
return nil
}



/*:35*/



/*57:*/


//line gdbsh.w:825

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



/*:57*/


