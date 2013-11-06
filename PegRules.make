.SUFFIXES: .peg .go

CP = cp
PEGPARSER = $(firstword $(subst :, ,$(GOPATH)))/bin/pegparser
buildPeg = $(PEGPARSER) "-peg=$(1)" -notest -ignore="$(2)" -testfile="$(3)" -outpath "$(dir $@)"

$(PEGPARSER):
	go build -o $(PEGPARSER) github.com/quarnster/parser/pegparser

%.go: %.peg $(PEGPARSER)
	$(call buildPeg,$<,$(ignore_$(subst .go,,$(notdir $@))),$(testfile_$(subst .go,,$(notdir $@))))
