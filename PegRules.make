.SUFFIXES: .peg .go

CP = cp

buildPeg = pegparser "-peg=$(1)" -notest -ignore="$(2)" -testfile="$(3)" -outpath "$(dir $@)"

ppegparser:
	go install github.com/quarnster/parser/pegparser

.peg.go: ppegparser
	$(call buildPeg,$<,$(ignore_$(subst .go,,$(notdir $@))),$(testfile_$(subst .go,,$(notdir $@))))
