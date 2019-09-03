module github.com/mwat56/kaliber

go 1.12

require (
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/mwat56/apachelogger v1.2.5
	github.com/mwat56/errorhandler v1.1.0
	github.com/mwat56/ini v1.3.4
	github.com/mwat56/passlist v1.1.2
	github.com/mwat56/sessions v0.3.2
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190902133755-9109b7679e13 // indirect
	golang.org/x/tools v0.0.0-20190903025054-afe7f8212f0d // indirect
)

replace (
	github.com/mwat56/apachelogger => ../apachelogger
	github.com/mwat56/errorhandler => ../errorhandler
	github.com/mwat56/ini => ../ini
	github.com/mwat56/passlist => ../passlist
	github.com/mwat56/sessions => ../sessions
)
