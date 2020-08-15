

/*2:*/


//line mfind.w:13

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
//line mfind.w:15

package main

import(


/*3:*/


//line mfind.w:45

"fmt"
"strings"
"flag"



/*:3*/



/*6:*/


//line mfind.w:101

"github.com/santucco/gdbsh/common"



/*:6*/



/*10:*/


//line mfind.w:138

"io"



/*:10*/


//line mfind.w:19

"github.com/golang/glog"
"os"
"bufio"
)

var(


/*4:*/


//line mfind.w:51

options string
values[]string
help bool
size string
num uint



/*:4*/



/*7:*/


//line mfind.w:105

sections[]string



/*:7*/


//line mfind.w:26

debug glog.Level= 1
)

func main(){
defer glog.Flush()
glog.V(debug).Infoln("main")
defer glog.V(debug).Infoln("main is done")


/*5:*/


//line mfind.w:59

{
flag.BoolVar(&help,"help",false,"print the help")
flag.StringVar(&size,"size","","search query size: b (bytes), h (halfwords - two bytes), w (words - four bytes), g (giant words - eight bytes)")
flag.UintVar(&num,"number",0,"maximum number of finds (default all)")
flag.Usage= func(){
fmt.Fprint(os.Stderr,
"mfind 0.31, GDB extention command for using from GDBSh\n",
"Copyright (C) 2015, 2016 Alexander Sychev\n",
"Search memory for the sequences of bytes\n",
"Usage:\n\tmfind [options] <sequence1> [<sequence2>...]\n",
"Options:\n")
flag.PrintDefaults()
}
flag.Parse()
if help{
flag.Usage()
return
}
if len(size)!=0{
if len(size)> 1{
fmt.Fprint(os.Stderr,"wrong search query size: %s",size)
flag.Usage()
return
}
switch size[0]{
case'b','h','w','g':
options+= "/"+size
default:
fmt.Fprint(os.Stderr,"wrong search query size: %s",size)
flag.Usage()
return
}
}
if num!=0{
options+= fmt.Sprintf(" /%d",num)
}
values= flag.Args()

}



/*:5*/


//line mfind.w:34

gdbin:=os.NewFile(uintptr(3),"input")
gdbout:=os.NewFile(uintptr(4),"output")
defer gdbin.Close()
defer gdbout.Close()
defer os.Stdout.Close()


/*8:*/


//line mfind.w:109

{
var err error
sections,err= common.Sections(gdbin,gdbout)
if err!=nil{
fmt.Fprintf(os.Stderr,"can't get sections from GDB: %s\n",err)
return
}
}



/*:8*/


//line mfind.w:40



/*9:*/


//line mfind.w:120

{
glog.V(debug).Infof("%#v",sections)
if len(values)!=0{
for _,val:=range values{


/*11:*/


//line mfind.w:142

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


//line mfind.w:125

}
}else{
stdr:=bufio.NewReader(os.Stdin)
for val,err:=stdr.ReadString('\n');err==nil;val,err= stdr.ReadString('\n'){
val= strings.TrimSpace(val)


/*11:*/


//line mfind.w:142

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


//line mfind.w:131

}
}

}



/*:9*/


//line mfind.w:41

}



/*:2*/


