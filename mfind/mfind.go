

/*2:*/


//line mfind.w:14

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
//line mfind.w:16

package main

import(


/*3:*/


//line mfind.w:46

"fmt"
"strings"



/*:3*/



/*6:*/


//line mfind.w:86

"bitbucket.org/santucco/gdbsh/common"



/*:6*/



/*10:*/


//line mfind.w:124

"io"



/*:10*/


//line mfind.w:20

"github.com/golang/glog"
"os"
"bufio"
)

var(


/*4:*/


//line mfind.w:51

options string
values string



/*:4*/



/*7:*/


//line mfind.w:90

sections[]string



/*:7*/


//line mfind.w:27

debug glog.Level= 0
)

func main(){
defer glog.Flush()
glog.V(debug).Infoln("main")
defer glog.V(debug).Infoln("main is done")


/*5:*/


//line mfind.w:56

{
if len(os.Args)==2&&strings.TrimSpace(os.Args[1])=="-h"{
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
for i:=1;i<len(os.Args);i++{
if os.Args[i][0]=='/'{
options+= os.Args[i]
}else{
values+= " "+os.Args[i]
}
}
}



/*:5*/


//line mfind.w:35

gdbin:=os.NewFile(uintptr(3),"input")
gdbout:=os.NewFile(uintptr(4),"output")
defer gdbin.Close()
defer gdbout.Close()
defer os.Stdout.Close()


/*8:*/


//line mfind.w:94

{
var err error
sections,err= common.Sections(gdbin,gdbout)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't get sections from GDB: %s\n",err)
return
}
}



/*:8*/


//line mfind.w:41



/*9:*/


//line mfind.w:105

{
glog.V(debug).Infof("%#v",sections)
vl:=strings.Fields(values)
if len(vl)!=0{
for _,val:=range vl{


/*11:*/


//line mfind.w:128

{
v:=fmt.Sprintf("%s:\n",val)
for _,a:=range sections{
al,err:=common.FindAddress(gdbin,gdbout,options,a,val)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't find address %s in section: %s\n",val,a,err)
continue
}
glog.V(debug).Infof("%s found in %#v",val,al)
for _,s:=range al{
v= fmt.Sprintf("%s\t%s\n",v,s)
if n,err:=io.WriteString(os.Stdout,v);err!=nil||n!=len(v){
glog.Warningf("can't write '%s' to stdout, %d bytes has been written: %s",v,n,err)
return
}
v= ""
}
}
}



/*:11*/


//line mfind.w:111

}
}else{
stdr:=bufio.NewReader(os.Stdin)
for val,err:=stdr.ReadString('\n');err==nil;val,err= stdr.ReadString('\n'){
val= strings.TrimSpace(val)


/*11:*/


//line mfind.w:128

{
v:=fmt.Sprintf("%s:\n",val)
for _,a:=range sections{
al,err:=common.FindAddress(gdbin,gdbout,options,a,val)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't find address %s in section: %s\n",val,a,err)
continue
}
glog.V(debug).Infof("%s found in %#v",val,al)
for _,s:=range al{
v= fmt.Sprintf("%s\t%s\n",v,s)
if n,err:=io.WriteString(os.Stdout,v);err!=nil||n!=len(v){
glog.Warningf("can't write '%s' to stdout, %d bytes has been written: %s",v,n,err)
return
}
v= ""
}
}
}



/*:11*/


//line mfind.w:117

}
}

}



/*:9*/


//line mfind.w:42

}



/*:2*/


