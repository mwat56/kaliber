module github.com/mwat56/kaliber

go 1.14

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/mwat56/apachelogger v1.4.5
	github.com/mwat56/cssfs v0.2.1
	github.com/mwat56/errorhandler v1.1.5
	github.com/mwat56/ini v1.3.8
	github.com/mwat56/jffs v0.1.0
	github.com/mwat56/kaliber/db v0.0.0-20200328184510-3ae73aad6235
	github.com/mwat56/passlist v1.3.1
	github.com/mwat56/sessions v0.3.8
	github.com/mwat56/whitespace v0.2.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59 // indirect
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775 // indirect
)

replace github.com/mwat56/kaliber/db => ./db
