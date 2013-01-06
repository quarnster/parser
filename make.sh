go run exe/main.go -peg xml.peg -dumptree -testfile "/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/C++/C.tmLanguage" -ignore "Spacing,Comment,EndTag,Tag,XmlFile" -notest
go run exe/main.go -peg json.peg -testfile ../sample.json -ignore "Spacing,Values,Value,QuotedText,KeyValuePairs,JsonFile" -notest
go run exe/main.go -peg plistxml.peg -dumptree -testfile "/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/C++/C.tmLanguage" -ignore "Spacing,KeyValuePair,KeyTag,StringTag,Value,Values,PlistFile,Plist"  -notest
go run exe/main.go -peg ini.peg -testfile /Volumes/BOOTCAMP/Windows/win.ini -ignore "EndOfLine,KeyValuePair,IniFile" -notest
go test parser/json parser/xml parser/peg parser/plistxml parser/ini
