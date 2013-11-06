.SUFFIXES: .peg .go

CP = cp
PEGPARSER = $(GOPATH)/bin/pegparser
buildPeg = $(PEGPARSER) "-peg=$(1)" -notest -ignore="$(2)" -testfile="$(3)" -outpath "$(dir $@)"

$(PEGPARSER):
	go install github.com/quarnster/parser/pegparser

%.go: %.peg $(PEGPARSER)
	$(call buildPeg,$<,$(ignore_$(subst .go,,$(notdir $@))),$(testfile_$(subst .go,,$(notdir $@))))
