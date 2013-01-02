go run exe/main.go -peg xml.peg -testfile "/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/C++/C.tmLanguage" -ignore "Spacing,Comment,EndTag,Tag" -notest
go run exe/main.go -peg json.peg -testfile sample.json -ignore "Spacing,Values,Value,QuotedText,KeyValuePairs" -notest
go run exe/main.go -peg plistxml.peg -testfile "/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/C++/C.tmLanguage" -notest -ignore "Spacing,KeyValuePair,KeyTag,StringTag,Value,Values"
go run exe/main.go -peg ini.peg -testfile /Volumes/BOOTCAMP/Windows/win.ini -dumptree -ignore "EndOfLine,KeyValuePair" -notest
go test parser/json parser/xml parser/peg parser/plistxml parser/ini
