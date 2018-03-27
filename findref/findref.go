

/*2:*/


//line findref.w:13

//line license:1

// This file is part of GDBSh toolset
// Author Alexander Sychev
//
// Copyright (c) 2015, 2016 Alexander Sychev. All rights reserved.
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
//line findref.w:15

package main

import(


/*3:*/


//line findref.w:44

"fmt"
"strings"
"flag"



/*:3*/



/*6:*/


//line findref.w:81

"bitbucket.org/santucco/gdbsh/common"



/*:6*/



/*10:*/


//line findref.w:119

"io"



/*:10*/



/*14:*/


//line findref.w:159

"strconv"



/*:14*/


//line findref.w:19

"github.com/golang/glog"
"os"
"bufio"
)

var(


/*4:*/


//line findref.w:50

instances[]string
offset uint
help bool



/*:4*/



/*7:*/


//line findref.w:85

sections[]string



/*:7*/



/*11:*/


//line findref.w:123

vtables[]string



/*:11*/


//line findref.w:26

debug glog.Level= 1
)

func main(){
defer glog.Flush()
glog.V(debug).Infoln("main")
defer glog.V(debug).Infoln("main is done")


/*5:*/


//line findref.w:56

{
flag.BoolVar(&help,"help",false,"print the help")
flag.UintVar(&offset,"offset",160,"a size of offset backward for analysing")
flag.Usage= func(){
fmt.Fprint(os.Stderr,
"findref 0.31, GDB extention command for using from GDBSh\n",
"Copyright (C) 2015, 2016 Alexander Sychev\n",
"Search for instances of virtual objects potentially have a reference to the specified instances\n",
"Usage:\n\tfindref [options] <instance1> [<instance2>...]\n",
"Options:\n")
flag.PrintDefaults()
}
flag.Parse()
if help{
flag.Usage()
return
}
glog.V(debug).Infof("args: %#v",flag.Args())
if len(flag.Args())> 0{
instances= flag.Args()
}
}



/*:5*/


//line findref.w:34

gdbin:=os.NewFile(uintptr(3),"input")
gdbout:=os.NewFile(uintptr(4),"output")
defer gdbin.Close()
defer gdbout.Close()


/*8:*/


//line findref.w:89

{
var err error
sections,err= common.Sections(gdbin,gdbout)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't get sections from GDB: %s\n",err)
return
}
}




/*:8*/


//line findref.w:39



/*9:*/


//line findref.w:101

{
glog.V(debug).Infof("instances: %#v",instances)
if len(instances)!=0{
for _,val:=range instances{


/*12:*/


//line findref.w:127

{
vtables,err:=common.Vtables(gdbin,gdbout,val)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't get vtables for %s from GDB: %s\n",val,err)
return
}
glog.V(debug).Infof("vtables: %#v",vtables)
rl:=make(map[string][]string)


/*13:*/


//line findref.w:142

{
for _,v:=range vtables{
for _,a:=range sections{
glog.V(debug).Infof("searching for %#v in  %#v",v,a)
al,err:=common.FindAddress(gdbin,gdbout,"",a,v)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't find address %s in section: %s\n",v,a,err)
return
}
rl[v]= append(rl[v],al...)
}
}
glog.V(debug).Infof("addresses: %#v",rl)
}



/*:13*/


//line findref.w:136



/*15:*/


//line findref.w:163

{
cmds:=make(map[string]map[int64][]string)


/*16:*/


//line findref.w:172

{
rp:=strings.NewReplacer("\\n","","\\t","","\\\"","\"")
d:=true
for address,r:=range rl{
for _,a:=range r{
adr,err:=strconv.ParseInt(a,0,64)
if err!=nil{
continue
}
for i:=adr;i> adr-int64(offset);i-= 8{
o,r,err:=common.RunCommand(gdbin,gdbout,fmt.Sprintf("-data-read-memory-bytes 0x%x 8",i))
if err!=nil{
continue
}
var a string
if m,ok:=r.Get("memory");!ok{
continue
}else if vl,ok:=m.Val.(common.ValueList);!ok||len(vl)==0{
continue
}else if t,ok:=vl[0].(common.Tuple);!ok{
continue
}else if c,ok:=t.Get("contents");!ok{
continue
}else if s,ok:=c.Val.(string);ok{
for i:=len(s)-1;i>=0;i-= 2{
a+= s[i-1:i+1]
}
}

o,_,err= common.RunCommand(gdbin,gdbout,fmt.Sprintf("info symbol 0x%s",a))
if err!=nil{
continue
}
for _,s:=range o{
s= rp.Replace(s)
if strings.HasPrefix(s,"No symbol matches")||len(s)==0{
glog.V(debug).Info(s)
continue
}
var p int
if p= strings.Index(s," in ");p==-1{
p= len(s)
}
sym:=s[0:p]
if p:=strings.LastIndex(sym,"+");p!=-1{
sym= sym[:p]
}
glog.V(debug).Info(sym)
if d{
if o,_,err= common.RunCommand(gdbin,gdbout,fmt.Sprintf("demangle %s",sym));err==nil&&len(o)!=0{
sym= o[0]
}else if err!=nil{
d= false
}
}
if strings.HasPrefix(sym,"vtable for "){
if _,ok:=cmds[address];!ok{
cmds[address]= make(map[int64][]string)
}
cmds[address][i]= append(cmds[address][i],fmt.Sprintf("p *(%s*)0x%x\n",sym[11:],i))
}
}
}
}
}
}



/*:16*/


//line findref.w:166



/*17:*/


//line findref.w:241

{
rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")

for a,v:=range cmds{
for aa,c:=range v{
o,_,err:=common.RunCommand(gdbin,gdbout,c[0])
if err!=nil{
continue
}
f:=false
adr:=strings.Trim(a,"0x")
for _,s:=range o{
glog.V(debug).Infof("looking for %s in '%s'",adr,s)
if i:=strings.Index(s,adr);i!=-1{
glog.V(debug).Infof("%s has been found in '%s'",adr,s)
f= true
break
}
}
if!f{
continue
}
p:=fmt.Sprintf("%s at 0x%x: ",a,aa)
if n,err:=io.WriteString(os.Stdout,p);err!=nil||n!=len(p){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",p,n,err)
return
}
for _,s:=range o{
glog.V(debug).Info(s)
s= rp.Replace(s)
if n,err:=io.WriteString(os.Stdout,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",s,n,err)
return
}
}
}
}
}



/*:17*/


//line findref.w:167

}




/*:15*/


//line findref.w:137


}



/*:12*/


//line findref.w:106

}
}else{
stdr:=bufio.NewReader(os.Stdin)
for val,err:=stdr.ReadString('\n');err==nil;val,err= stdr.ReadString('\n'){
val= strings.TrimSpace(val)


/*12:*/


//line findref.w:127

{
vtables,err:=common.Vtables(gdbin,gdbout,val)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't get vtables for %s from GDB: %s\n",val,err)
return
}
glog.V(debug).Infof("vtables: %#v",vtables)
rl:=make(map[string][]string)


/*13:*/


//line findref.w:142

{
for _,v:=range vtables{
for _,a:=range sections{
glog.V(debug).Infof("searching for %#v in  %#v",v,a)
al,err:=common.FindAddress(gdbin,gdbout,"",a,v)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't find address %s in section: %s\n",v,a,err)
return
}
rl[v]= append(rl[v],al...)
}
}
glog.V(debug).Infof("addresses: %#v",rl)
}



/*:13*/


//line findref.w:136



/*15:*/


//line findref.w:163

{
cmds:=make(map[string]map[int64][]string)


/*16:*/


//line findref.w:172

{
rp:=strings.NewReplacer("\\n","","\\t","","\\\"","\"")
d:=true
for address,r:=range rl{
for _,a:=range r{
adr,err:=strconv.ParseInt(a,0,64)
if err!=nil{
continue
}
for i:=adr;i> adr-int64(offset);i-= 8{
o,r,err:=common.RunCommand(gdbin,gdbout,fmt.Sprintf("-data-read-memory-bytes 0x%x 8",i))
if err!=nil{
continue
}
var a string
if m,ok:=r.Get("memory");!ok{
continue
}else if vl,ok:=m.Val.(common.ValueList);!ok||len(vl)==0{
continue
}else if t,ok:=vl[0].(common.Tuple);!ok{
continue
}else if c,ok:=t.Get("contents");!ok{
continue
}else if s,ok:=c.Val.(string);ok{
for i:=len(s)-1;i>=0;i-= 2{
a+= s[i-1:i+1]
}
}

o,_,err= common.RunCommand(gdbin,gdbout,fmt.Sprintf("info symbol 0x%s",a))
if err!=nil{
continue
}
for _,s:=range o{
s= rp.Replace(s)
if strings.HasPrefix(s,"No symbol matches")||len(s)==0{
glog.V(debug).Info(s)
continue
}
var p int
if p= strings.Index(s," in ");p==-1{
p= len(s)
}
sym:=s[0:p]
if p:=strings.LastIndex(sym,"+");p!=-1{
sym= sym[:p]
}
glog.V(debug).Info(sym)
if d{
if o,_,err= common.RunCommand(gdbin,gdbout,fmt.Sprintf("demangle %s",sym));err==nil&&len(o)!=0{
sym= o[0]
}else if err!=nil{
d= false
}
}
if strings.HasPrefix(sym,"vtable for "){
if _,ok:=cmds[address];!ok{
cmds[address]= make(map[int64][]string)
}
cmds[address][i]= append(cmds[address][i],fmt.Sprintf("p *(%s*)0x%x\n",sym[11:],i))
}
}
}
}
}
}



/*:16*/


//line findref.w:166



/*17:*/


//line findref.w:241

{
rp:=strings.NewReplacer("\\n","\n","\\t","\t","\\\"","\"")

for a,v:=range cmds{
for aa,c:=range v{
o,_,err:=common.RunCommand(gdbin,gdbout,c[0])
if err!=nil{
continue
}
f:=false
adr:=strings.Trim(a,"0x")
for _,s:=range o{
glog.V(debug).Infof("looking for %s in '%s'",adr,s)
if i:=strings.Index(s,adr);i!=-1{
glog.V(debug).Infof("%s has been found in '%s'",adr,s)
f= true
break
}
}
if!f{
continue
}
p:=fmt.Sprintf("%s at 0x%x: ",a,aa)
if n,err:=io.WriteString(os.Stdout,p);err!=nil||n!=len(p){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",p,n,err)
return
}
for _,s:=range o{
glog.V(debug).Info(s)
s= rp.Replace(s)
if n,err:=io.WriteString(os.Stdout,s);err!=nil||n!=len(s){
glog.V(debug).Infof("can't write '%s' to stdout, %d bytes has been written: %s",s,n,err)
return
}
}
}
}
}



/*:17*/


//line findref.w:167

}




/*:15*/


//line findref.w:137


}



/*:12*/


//line findref.w:112

}
}

}



/*:9*/


//line findref.w:40

}



/*:2*/


