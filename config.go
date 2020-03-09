/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"crypto/md5" // #nosec G501
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mwat56/ini"
	"github.com/mwat56/kaliber/db"
	"github.com/mwat56/sessions"
)

type (
	// tArguments is the list structure for the cmdline argument values
	// merged with the key-value pairs from the INI file.
	tArguments struct {
		ini.TSection // embedded INI section
	}
)

// Get returns the value associated with `aKey` and `nil` if found,
// or an empty string and an error.
//
//	`aKey` The requested value's key to lookup.
func (al *tArguments) Get(aKey string) (string, error) {
	if result, ok := al.AsString(aKey); ok {
		return result, nil
	}

	//lint:ignore ST1005 – capitalisation wanted
	return "", fmt.Errorf("Missing config value: %s", aKey)
} // Get()

// `set()` adds/sets another key-value pair.
//
// If `aValue` is empty then `aKey` gets removed.
func (al *tArguments) set(aKey, aValue string) {
	if 0 < len(aValue) {
		al.AddKey(aKey, aValue)
	} else {
		al.RemoveKey(aKey)
	}
} // set()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `absolute()` returns `aDir` as an absolute path.
func absolute(aBaseDir, aDir string) string {
	if 0 == len(aDir) {
		return aDir
	}
	if '/' == aDir[0] {
		s, _ := filepath.Abs(aDir)
		return s
	}

	return filepath.Join(aBaseDir, aDir)
} // absolute()

var (
	// AppArguments is the merged list for the cmdline arguments
	// and INI values for the application.
	AppArguments tArguments
)

// `readIniData()` returns the config values read from INI file(s).
//
// The steps here are:
//	(1) read the local `./.kaliber.ini`,
//	(2) read the global `/etc/kaliber.ini`,
//	(3) read the user-local `~/.kaliber.ini`,
//	(4) read the user-local `~/.config/kaliber.ini`,
//	(5) read the `-ini` commandline argument.
func readIniData() {
	// (1) ./
	fName, _ := filepath.Abs("./kaliber.ini")
	ini1, err := ini.New(fName)
	if nil == err {
		ini1.AddSectionKey("", "iniFile", fName)
	}

	// (2) /etc/
	fName = "/etc/kaliber.ini"
	if ini2, err2 := ini.New(fName); nil == err2 {
		ini1.Merge(ini2)
		ini1.AddSectionKey("", "iniFile", fName)
	}

	// (3) ~user/
	if fName, _ = os.UserHomeDir(); 0 < len(fName) {
		fName, _ = filepath.Abs(filepath.Join(fName, ".kaliber.ini"))
		if ini2, err2 := ini.New(fName); nil == err2 {
			ini1.Merge(ini2)
			ini1.AddSectionKey("", "iniFile", fName)
		}
	}

	// (4) ~/.config/
	if confDir, err3 := os.UserConfigDir(); nil == err3 {
		fName, _ = filepath.Abs(filepath.Join(confDir, "kaliber.ini"))
		if ini2, err2 := ini.New(fName); nil == err2 {
			ini1.Merge(ini2)
			ini1.AddSectionKey("", "iniFile", fName)
		}
	}

	// (5) cmdline
	for aLen, i := len(os.Args), 1; i < aLen; i++ {
		if `-ini` == os.Args[i] {
			//XXX Note that this works only if `-ini` and
			// filename are two separate arguments. It will
			// fail if it's given in the form `-ini=filename`.
			i++
			if i < aLen {
				fName, _ = filepath.Abs(os.Args[i])
				if ini2, _ := ini.New(fName); nil == err {
					ini1.Merge(ini2)
					ini1.AddSectionKey("", "iniFile", fName)
				}
			}
			break
		}
	}

	AppArguments = tArguments{*ini1.GetSection("")}
} // readIniData()

/*
func init() {
	// see: https://github.com/microsoft/vscode-go/issues/2734
	testing.Init() // workaround for Go 1.13
	InitConfig()
} // init()
*/

// InitConfig reads the commandline arguments into a list
// structure merging it with key-value pairs read from INI file(s).
//
// The steps here are:
//	(a) read the INI file(s),
//	(b) merge the commandline arguments with the INI values
// into the global `AppArguments` variable.
func InitConfig() {
	readIniData()

	authBool, _ := AppArguments.AsBool("authAll")
	flag.BoolVar(&authBool, "authAll", authBool,
		"<boolean> whether to require authentication for all pages ")

	bppInt, _ := AppArguments.AsInt("booksPerPage")
	flag.IntVar(&bppInt, "booksPerPage", bppInt,
		"<number> the default number of books shown per page ")

	s, _ := AppArguments.Get("dataDir")
	dataDir, _ := filepath.Abs(s)
	flag.StringVar(&dataDir, "dataDir", dataDir,
		"<dirName> the directory with CSS, FONTS, IMG, SESSIONS, and VIEWS sub-directories\n")

	s, _ = AppArguments.Get("accessLog")
	accessLog := absolute(dataDir, s)
	flag.StringVar(&accessLog, "accessLog", accessLog,
		"<filename> Name of the access logfile to write to\n")

	s, _ = AppArguments.Get("certKey")
	certKey := absolute(dataDir, s)
	flag.StringVar(&certKey, "certKey", certKey,
		"<fileName> the name of the TLS certificate key\n")

	s, _ = AppArguments.Get("certPem")
	certPem := absolute(dataDir, s)
	flag.StringVar(&certPem, "certPem", certPem,
		"<fileName> the name of the TLS certificate PEM\n")

	s, _ = AppArguments.Get("errorLog")
	errorLog := absolute(dataDir, s)
	flag.StringVar(&errorLog, "errorlog", errorLog,
		"<filename> Name of the error logfile to write to\n")

	gzipBool, _ := AppArguments.AsBool("gzip")
	flag.BoolVar(&gzipBool, "gzip", gzipBool,
		"<boolean> use gzip compression for server responses")

	/*
		s, _ = AppArguments.Get("intl")
		intlStr := absolute(dataStr, s)
		flag.StringVar(&intlStr, "intl", intlStr,
			"<fileName> the path/filename of the localisation file\n")
	*/

	iniFile, _ := AppArguments.Get("iniFile")
	flag.StringVar(&iniFile, "ini", iniFile,
		"<fileName> the path/filename of the INI file to use\n")

	langStr, _ := AppArguments.Get("lang")
	flag.StringVar(&langStr, "lang", langStr,
		"the default language to use ")

	libName, _ := AppArguments.Get("libraryName")
	flag.StringVar(&libName, "libraryName", libName,
		"Name of this Library (shown on every page)\n")

	libPath, _ := AppArguments.Get("libraryPath")
	flag.StringVar(&libPath, "libraryPath", libPath,
		"<pathname> Path name of/to the Calibre library\n")

	listenStr, _ := AppArguments.Get("listen")
	flag.StringVar(&listenStr, "listen", listenStr,
		"the host's IP to listen at ")

	logStack, _ := AppArguments.AsBool("logStack")
	flag.BoolVar(&logStack, "logStack", logStack,
		"<boolean> Log a stack trace for recovered runtime errors ")

	portInt, _ := AppArguments.AsInt("port")
	flag.IntVar(&portInt, "port", portInt,
		"<portNumber> The IP port to listen to ")

	realmStr, _ := AppArguments.Get("realm")
	flag.StringVar(&realmStr, "realm", realmStr,
		"<hostName> Name of host/domain to secure by BasicAuth\n")

	sessionTTL, _ := AppArguments.AsInt("sessionTTL")
	flag.IntVar(&sessionTTL, "sessionTTL", sessionTTL,
		"<seconds> Number of seconds an unused session keeps valid ")

	sidName, _ := AppArguments.Get("sidName")
	flag.StringVar(&sidName, "sidName", sidName,
		"<name> The name of the session ID to use\n")

	s, _ = AppArguments.Get("sqlTrace")
	sqlTrace := absolute(dataDir, s)
	flag.StringVar(&sqlTrace, "sqlTrace", sqlTrace,
		"<filename> Name of the SQL logfile to write to\n")

	themeStr, _ := AppArguments.Get("theme")
	flag.StringVar(&themeStr, "theme", themeStr,
		"<name> The display theme to use ('light' or 'dark')\n")

	uaStr := ""
	flag.StringVar(&uaStr, "ua", uaStr,
		"<userName> User add: add a username to the password file")

	ucStr := ""
	flag.StringVar(&ucStr, "uc", ucStr,
		"<userName> User check: check a username in the password file")

	udStr := ""
	flag.StringVar(&udStr, "ud", udStr,
		"<userName> User delete: remove a username from the password file")

	s, _ = AppArguments.Get("passFile")
	ufStr := absolute(dataDir, s)
	flag.StringVar(&ufStr, "uf", ufStr,
		"<fileName> Passwords file storing user/passwords for BasicAuth\n")

	ulBool := false
	flag.BoolVar(&ulBool, "ul", ulBool,
		"<boolean> User list: show all users in the password file")

	uuStr := ""
	flag.StringVar(&uuStr, "uu", uuStr,
		"<userName> User update: update a username in the password file")

	flag.Usage = ShowHelp
	flag.Parse() // // // // // // // // // // // // // // // // // // //

	if authBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("authAll", s)

	AppArguments.set("booksPerPage", fmt.Sprintf("%d", bppInt))

	if 0 < len(dataDir) {
		dataDir, _ = filepath.Abs(dataDir)
	}
	if f, err := os.Stat(dataDir); nil != err {
		log.Fatalf("dataDir == %s` problem: %v", dataDir, err)
	} else if !f.IsDir() {
		log.Fatalf("Error: Not a directory `%s`", dataDir)
	}
	AppArguments.set("dataDir", dataDir)

	if 0 < len(accessLog) {
		accessLog = absolute(dataDir, accessLog)
	}
	AppArguments.set("accessLog", accessLog)

	if 0 < len(certKey) {
		certKey = absolute(dataDir, certKey)
		if fi, err := os.Stat(certKey); (nil != err) || (0 >= fi.Size()) {
			certKey = ""
		}
	}
	AppArguments.set("certKey", certKey)

	if 0 < len(certPem) {
		certPem = absolute(dataDir, certPem)
		if fi, err := os.Stat(certPem); (nil != err) || (0 >= fi.Size()) {
			certPem = ""
		}
	}
	AppArguments.set("certPem", certPem)

	if 0 < len(errorLog) {
		errorLog = absolute(dataDir, errorLog)
	}
	AppArguments.set("errorLog", errorLog)

	if gzipBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("gzip", s)

	if 0 == len(langStr) {
		langStr = "en"
	}
	AppArguments.set("lang", strings.ToLower(langStr))

	if 0 == len(libName) {
		libName = time.Now().Format("2006:01:02:15:04:05")
	}
	AppArguments.set("libraryName", libName)

	// To allow for use of multiple libraries we add the MD5
	// of the libraryPath to our cache path.
	s = fmt.Sprintf("%x", md5.Sum([]byte(libPath))) // #nosec G401
	if ucd, err := os.UserCacheDir(); (nil != err) || (0 == len(ucd)) {
		db.SetCalibreCachePath(filepath.Join(dataDir, "img", s))
	} else {
		db.SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	}
	db.SetCalibreLibraryPath(libPath)

	if "0" == listenStr {
		listenStr = ""
	}
	AppArguments.set("listen", listenStr)

	if logStack {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("logStack", s)

	AppArguments.set("port", fmt.Sprintf("%d", portInt))

	AppArguments.set("realm", realmStr)

	AppArguments.set("sessiondir", absolute(dataDir, "sessions"))

	if 0 == len(sidName) {
		sidName = "SID"
	}
	sessions.SetSIDname(sidName)

	if 0 >= sessionTTL {
		sessionTTL = 1200
	}
	sessions.SetSessionTTL(sessionTTL)

	if 0 < len(sqlTrace) {
		sqlTrace = absolute(dataDir, sqlTrace)
	}
	db.SetSQLtraceFile(sqlTrace)

	AppArguments.set("theme", strings.ToLower(themeStr))
	AppArguments.set("ua", uaStr)
	AppArguments.set("uc", ucStr)
	AppArguments.set("ud", udStr)

	if 0 < len(ufStr) {
		ufStr = absolute(dataDir, ufStr)
	}
	AppArguments.set("uf", ufStr)

	if ulBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("ul", s)

	AppArguments.set("uu", uuStr)
} // InitConfig()

// ShowHelp lists the commandline options to `Stderr`.
func ShowHelp() {
	fmt.Fprintf(os.Stderr, "\n  Usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\n  Most options can be set in an INI file to keep the command-line short ;-)")
} // ShowHelp()

/* _EoF_ */
