# gdbsh

**gdbsh** is a command shell for **GDB** with possibility to attach outer extentions.
It supports all the commands of gdb with using them in pipes.
A pipe can contain eather **GDB** commands or outer commands.

A special command *args* can use with conjunction of **GDB** command to get input data from pipe.
The **AWK**-like approach is used: input data is separated into fields with space like a field separator.
The fields can be addressed by *$<n>*, where *n* - a number of a field starts with 1.
A special *$0* address can be used for whole input data.
