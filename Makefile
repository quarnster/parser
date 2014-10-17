
PEGS = xml/xml.go json/json.go plistxml/plistxml.go ini/ini.go expression/expression.go

ignore_xml= Spacing,Comment,EndTag,Tag,XmlFile
ignore_json = Spacing,Values,Value,QuotedText,KeyValuePairs,JsonFile
ignore_plistxml = "Spacing,KeyValuePair,KeyTag,StringTag,Value,Values,PlistFile,Plist"
ignore_ini = "EndOfLine,KeyValuePair,IniFile"
ignore_expression = "Spacing,Primary,Op,Expression,Grouping"

testfile_xml= "/Users/quarnster/Library/Application\ Support/Sublime\ Text\ 3/Packages/c.tmbundle/Syntaxes/C.plist"
testfile_json = http://json-test-suite.googlecode.com/files/sample.zip
testfile_plistxml = "/Users/quarnster/Library/Application\ Support/Sublime\ Text\ 3/Packages/c.tmbundle/Syntaxes/C++.plist"
testfile_ini = /Volumes/BOOTCAMP/Windows/win.ini
testfile_expression = "expression.in"

all: $(PEGS) test

clean:
	rm -f $(PEGS)

test:
	go test github.com/quarnster/parser/json github.com/quarnster/parser/xml github.com/quarnster/parser/peg github.com/quarnster/parser/plistxml github.com/quarnster/parser/ini github.com/quarnster/parser/expression

bench: $(PEGS)
	 go test -bench . github.com/quarnster/parser/json github.com/quarnster/parser/xml github.com/quarnster/parser/peg github.com/quarnster/parser/plistxml github.com/quarnster/parser/ini github.com/quarnster/parser/expression


include PegRules.make
