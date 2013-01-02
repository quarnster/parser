go run exe/main.go -peg xml.peg -testfile "/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/C++/C.tmLanguage" -ignore "Spacing,Comment,EndTag,Tag" -notest
go run exe/main.go -peg json.peg -testfile sample.json -ignore "Spacing,Values,Value,QuotedText,KeyValuePairs" -notest
go test parser/json parser/xml parser/peg
