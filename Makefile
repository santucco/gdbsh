# This file is part of GDBSh toolset
# Author Alexander Sychev
#
# Copyright © 2015, 2016 Alexander Sychev. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
#
#    * Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
#    * Redistributions in binary form must reproduce the above
# copyright notice, this list of conditions and the following disclaimer
# in the documentation and/or other materials provided with the
# distribution.
#    * The name of author may not be used to endorse or promote products derived from
# this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

IFILES= \
	$(patsubst %.w, %.idx, $(wildcard *.w) $(wildcard mfind/*.w) $(wildcard common/*.w) $(wildcard findref/*.w)) \
	$(patsubst %.w, %.toc, $(wildcard *.w) $(wildcard mfind/*.w) $(wildcard common/*.w) $(wildcard findref/*.w)) \
	$(patsubst %.w, %.scn, $(wildcard *.w) $(wildcard mfind/*.w) $(wildcard common/*.w) $(wildcard findref/*.w)) \
	$(patsubst %.w, %.log, $(wildcard *.w) $(wildcard mfind/*.w) $(wildcard common/*.w) $(wildcard findref/*.w)) \
	$(patsubst %.w, %.tex, $(wildcard *.w) $(wildcard mfind/*.w) $(wildcard common/*.w) $(wildcard findref/*.w))

.INTERMEDIATE: $(IFILES)

TEXP?=xetex
#gcflags=-gcflags '-N -l'

TARGETS=common gdbsh mfind/mfind findref/findref

all: $(TARGETS)

%.go: %.w
	gotangle $< - $@

%.pdf %.idx %.toc %.log: %.tex header.tex
	$(TEXP) $<

%.tex %.scn: %.w
	goweave  $<

gdbsh: gdbsh.go common/common.go
	go build $(gcflags)
	@echo done

mfind/mfind: mfind/mfind.go common/common.go
	(cd mfind; go build  $(gcflags))

mfind/mfind.go: mfind.w
	-mkdir -p mfind
	gotangle $< - $@

findref/findref: findref/findref.go common/common.go
	(cd findref; go build  $(gcflags))

findref/findref.go: findref.w
	-mkdir -p findref
	gotangle $< - $@

common: common/common.go
	(cd common; go build  $(gcflags))

common/common.go: common.w
	-mkdir -p common
	gotangle $< - $@

%.pdf %.idx %.toc %.log: %.tex gowebmac.tex
	$(TEXP) -output-directory $(dir $<) $<

%.tex %.scn: %.w
	goweave $< - $(patsubst %.w, %, $<)

doc: gdbsh.pdf common.pdf mfind.pdf findref.pdf

install: common gdbsh mfind findref
	go install
	(cd mfind; go install)
	(cd findref; go install)

clean:
	go clean
	(cd mfind; go clean)
	(cd findref; go clean)
	rm -f *.pdf $(IFILES)




