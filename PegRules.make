.SUFFIXES: .peg .go

CP = cp
PEGPATH = "$(GOPATH)/bin/pegparser"
buildPeg = $(PEGPATH) "-peg=$(1)" -notest -ignore="$(2)" -testfile="$(3)" -outpath "$(dir $@)" -generator="$(4)"

PEGPARSER:
	go install github.com/quarnster/parser/pegparser

%.go: %.peg PEGPARSER
	$(call buildPeg,$<,$(ignore_$(subst .go,,$(notdir $@))),$(testfile_$(subst .go,,$(notdir $@))),go)

%.c: %.peg PEGPARSER
	$(call buildPeg,$<,$(ignore_$(subst .c,,$(notdir $@))),$(testfile_$(subst .c,,$(notdir $@))),c)

%.cpp: %.peg PEGPARSER
	$(call buildPeg,$<,$(ignore_$(subst .cpp,,$(notdir $@))),$(testfile_$(subst .cpp,,$(notdir $@))),cpp)
