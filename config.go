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
	"github.com/mwat56/whitespace"
)

type (
	// TAppArgs Commandline arguments and INI values.
	TAppArgs struct {
		AccessLog     string
		AuthAll       bool
		BooksPerPage  int
		CertKey       string
		CertPem       string
		DataDir       string
		DelWhitespace bool
		ErrorLog      string
		GZip          bool
		// Intl          string
		Lang       string
		LibName    string
		LibPath    string
		Listen     string
		LogStack   bool
		PassFile   string
		Port       int
		Realm      string
		SessionDir string
		SessionTTL int
		SidName    string
		SQLTrace   string
		Theme      string
		UserAdd    string
		UserCheck  string
		UserDelete string
		UserList   bool
		UserUpdate string
	}

	// List structure for the INI values.
	tArguments struct {
		ini.TSection // embedded INI section
	}
)

var (
	// AppArgs Commandline arguments and INI values.
	//
	// This structure should be considered R/O after it was
	// set up by a call to `InitConfig()`.
	AppArgs TAppArgs

	// `appArguments` is the merged list for the INI values.
	appArguments tArguments
)

// `absolute()` returns `aDir` as an absolute path.
//
// If `aDir` starts with a slash (`/`) it's returned after cleaning.
// Otherwise `aBaseDir` gets prepended to `aDir` and returned after cleaning.
//
//	`aBaseDir` The base directory to prepend to `aDir`.
//	`aDir` The directory to make absolute.
func absolute(aBaseDir, aDir string) string {
	if 0 == len(aDir) {
		return aDir
	}
	if '/' == aDir[0] {
		return filepath.Clean(aDir)
	}

	return filepath.Join(aBaseDir, aDir)
} // absolute()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

/*
func init() {
	// see: https://github.com/microsoft/vscode-go/issues/2734
	testing.Init() // workaround for Go 1.13
	InitConfig()
} // init()
*/

// InitConfig sets up and reads all configuration data from INI files
// and commandline arguments.
func InitConfig() {
	flag.CommandLine = flag.NewFlagSet(`Kaliber`, flag.ExitOnError)
	readIniFiles()
	setFlags()
	parseFlags()
	readFlags()
	flag.CommandLine = nil // free unneeded memory
} // InitConfig()

// `parseFlags()` parses the commandline arguments.
func parseFlags() {
	flag.Usage = ShowHelp
	_ = flag.CommandLine.Parse(os.Args[1:])
} // parseFlags()

// `readFlags()` checks all available configurations flags.
func readFlags() {
	if 0 == AppArgs.BooksPerPage {
		AppArgs.BooksPerPage = 24
	}

	if 0 == len(AppArgs.DataDir) {
		log.Fatalln("Error: Missing `dataDir` value")
	}
	AppArgs.DataDir, _ = filepath.Abs(AppArgs.DataDir)
	if f, err := os.Stat(AppArgs.DataDir); nil != err {
		log.Fatalf("`dataDir` == `%s` problem: %v", AppArgs.DataDir, err)
	} else if !f.IsDir() {
		log.Fatalf("Error: `dataDir` not a directory `%s`", AppArgs.DataDir)
	}

	if 0 < len(AppArgs.AccessLog) {
		AppArgs.AccessLog = absolute(AppArgs.DataDir, AppArgs.AccessLog)
	}

	if 0 < len(AppArgs.CertKey) {
		AppArgs.CertKey = absolute(AppArgs.DataDir, AppArgs.CertKey)
		if fi, err := os.Stat(AppArgs.CertKey); (nil != err) || (0 >= fi.Size()) {
			AppArgs.CertKey = ``
		}
	}

	if 0 < len(AppArgs.CertPem) {
		AppArgs.CertPem = absolute(AppArgs.DataDir, AppArgs.CertPem)
		if fi, err := os.Stat(AppArgs.CertPem); (nil != err) || (0 >= fi.Size()) {
			AppArgs.CertPem = ``
		}
	}

	whitespace.UseRemoveWhitespace = AppArgs.DelWhitespace

	if 0 < len(AppArgs.ErrorLog) {
		AppArgs.ErrorLog = absolute(AppArgs.DataDir, AppArgs.ErrorLog)
	}

	if 0 < len(AppArgs.Lang) {
		AppArgs.Lang = strings.ToLower(AppArgs.Lang)
	}
	switch AppArgs.Lang {
	case `de`, `en`:
	default:
		AppArgs.Lang = "en"
	}

	if 0 == len(AppArgs.LibName) {
		AppArgs.LibName = time.Now().Format("2006:01:02:15:04:05")
	}

	if 0 == len(AppArgs.LibPath) {
		log.Fatalln("Error: Missing `libPath` value")
	}
	AppArgs.LibPath, _ = filepath.Abs(AppArgs.LibPath)
	if f, err := os.Stat(AppArgs.LibPath); nil != err {
		log.Fatalf("`libPath` == `%s` problem: %v", AppArgs.LibPath, err)
	} else if !f.IsDir() {
		log.Fatalf("Error: `libPath` not a directory `%s`", AppArgs.LibPath)
	}

	// To allow for use of multiple libraries we add the MD5
	// of the libraryPath to our cache path.
	s := fmt.Sprintf("%x", md5.Sum([]byte(AppArgs.LibPath))) // #nosec G401
	if ucd, err := os.UserCacheDir(); (nil != err) || (0 == len(ucd)) {
		db.SetCalibreCachePath(filepath.Join(AppArgs.DataDir, "img", s))
	} else {
		db.SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	}
	db.SetCalibreLibraryPath(AppArgs.LibPath)

	if `0` == AppArgs.Listen {
		AppArgs.Listen = ``
	}

	if 0 >= AppArgs.Port {
		AppArgs.Port = 8383
	}

	if 0 == len(AppArgs.Realm) {
		AppArgs.Realm = `eBooks Host`
	}

	AppArgs.SessionDir = absolute(AppArgs.DataDir, `sessions`)

	if 0 == len(AppArgs.SidName) {
		AppArgs.SidName = "sid"
	}
	sessions.SetSIDname(AppArgs.SidName)

	if 0 >= AppArgs.SessionTTL {
		AppArgs.SessionTTL = 1200
	}
	sessions.SetSessionTTL(AppArgs.SessionTTL)

	if 0 < len(AppArgs.SQLTrace) {
		AppArgs.SQLTrace = absolute(AppArgs.DataDir, AppArgs.SQLTrace)
	}
	db.SetSQLtraceFile(AppArgs.SQLTrace)

	if 0 < len(AppArgs.Theme) {
		AppArgs.Theme = strings.ToLower(AppArgs.Theme)
	}
	switch AppArgs.Theme {
	case `dark`, `light`:
		// accepted values
	default:
		AppArgs.Theme = `dark`
	}

	if 0 < len(AppArgs.PassFile) {
		AppArgs.PassFile = absolute(AppArgs.DataDir, AppArgs.PassFile)
	}
} // readFlags()

// `readIniFiles()` processes the config values read from INI file(s).
//
// The steps here are:
//	(1) read the local `./.kaliber.ini`,
//	(2) read the global `/etc/kaliber.ini`,
//	(3) read the user-local `~/.kaliber.ini`,
//	(4) read the user-local `~/.config/kaliber.ini`,
//	(5) read the `-ini` commandline argument.
func readIniFiles() {
	// (1) ./
	fName, _ := filepath.Abs(`./kaliber.ini`)
	ini1, err := ini.New(fName)
	if nil == err {
		ini1.AddSectionKey(``, `iniFile`, fName)
	}

	// (2) /etc/
	fName = `/etc/kaliber.ini`
	if ini2, err2 := ini.New(fName); nil == err2 {
		ini1.Merge(ini2)
		ini1.AddSectionKey(``, `iniFile`, fName)
	}

	// (3) ~user/
	if fName, _ = os.UserHomeDir(); 0 < len(fName) {
		fName, _ = filepath.Abs(filepath.Join(fName, `.kaliber.ini`))
		if ini2, err2 := ini.New(fName); nil == err2 {
			ini1.Merge(ini2)
			ini2.Clear()
			ini1.AddSectionKey(``, `iniFile`, fName)
		}
	}

	// (4) ~/.config/
	if confDir, err3 := os.UserConfigDir(); nil == err3 {
		fName, _ = filepath.Abs(filepath.Join(confDir, `kaliber.ini`))
		if ini2, err2 := ini.New(fName); nil == err2 {
			ini1.Merge(ini2)
			ini2.Clear()
			ini1.AddSectionKey(``, `iniFile`, fName)
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
					ini2.Clear()
					ini1.AddSectionKey(``, `iniFile`, fName)
				}
			}
			break
		}
	}

	appArguments = tArguments{*ini1.GetSection(``)}
} // readIniFiles()

// `setFlags()` sets up the flags for the commandline arguments.
func setFlags() {
	var (
		ok bool
		s  string // temp. value
	)
	if AppArgs.AuthAll, ok = appArguments.AsBool(`authAll`); !ok {
		AppArgs.AuthAll = true
	}
	flag.CommandLine.BoolVar(&AppArgs.AuthAll, `authAll`, AppArgs.AuthAll,
		"<boolean> whether to require authentication for all pages ")

	if AppArgs.BooksPerPage, ok = appArguments.AsInt(`booksPerPage`); (!ok) || (0 >= AppArgs.BooksPerPage) {
		AppArgs.BooksPerPage = 24
	}
	flag.CommandLine.IntVar(&AppArgs.BooksPerPage, `booksPerPage`, AppArgs.BooksPerPage,
		"<number> the default number of books shown per page ")

	if s, ok = appArguments.AsString("dataDir"); (ok) && (0 < len(s)) {
		AppArgs.DataDir, _ = filepath.Abs(s)
	} else {
		AppArgs.DataDir, _ = filepath.Abs(`./`)
	}
	flag.CommandLine.StringVar(&AppArgs.DataDir, "dataDir", AppArgs.DataDir,
		"<dirName> the directory with CSS, FONTS, IMG, SESSIONS, and VIEWS sub-directories\n")

	if s, ok = appArguments.AsString("accessLog"); (ok) && (0 < len(s)) {
		AppArgs.AccessLog = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.AccessLog, "accessLog", AppArgs.AccessLog,
		"<filename> Name of the access logfile to write to\n")

	if s, ok = appArguments.AsString("certKey"); (ok) && (0 < len(s)) {
		AppArgs.CertKey = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.CertKey, "certKey", AppArgs.CertKey,
		"<fileName> the name of the TLS certificate key\n")

	if s, ok = appArguments.AsString("certPem"); (ok) && (0 < len(s)) {
		AppArgs.CertPem = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.CertPem, "certPem", AppArgs.CertPem,
		"<fileName> the name of the TLS certificate PEM\n")

	if AppArgs.DelWhitespace, ok = appArguments.AsBool("delWhitespace"); !ok {
		AppArgs.DelWhitespace = true
	}
	flag.CommandLine.BoolVar(&AppArgs.DelWhitespace, "delWhitespace", AppArgs.DelWhitespace,
		"(optional) Delete superfluous whitespace in generated pages")

	if s, ok = appArguments.AsString("errorLog"); (ok) && (0 < len(s)) {
		AppArgs.ErrorLog = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.ErrorLog, "errorlog", AppArgs.ErrorLog,
		"<filename> Name of the error logfile to write to\n")

	if AppArgs.GZip, ok = appArguments.AsBool("gzip"); !ok {
		AppArgs.GZip = true
	}
	flag.CommandLine.BoolVar(&AppArgs.GZip, "gzip", AppArgs.GZip,
		"<boolean> use gzip compression for server responses")

	/* * /
	if s, ok = appArguments.AsString("intl"); (ok) && (0 < len(s)) {
		AppArgs.Intl = absolute(AppArgs.DataDir, s)
	}
	appFlags.StringVar(&AppArgs.Intl, "intl", AppArgs.Intl,
		"<fileName> the path/filename of the localisation file\n")
	/* */

	iniFile, _ := appArguments.AsString("iniFile")
	flag.CommandLine.StringVar(&iniFile, "ini", iniFile,
		"<fileName> the path/filename of the INI file to use\n")

	if AppArgs.Lang, ok = appArguments.AsString("lang"); (!ok) || (0 == len(AppArgs.Lang)) {
		AppArgs.Lang = `en`
	}
	flag.CommandLine.StringVar(&AppArgs.Lang, "lang", AppArgs.Lang,
		"the default language to use ")

	AppArgs.LibName, _ = appArguments.AsString("libraryName")
	flag.CommandLine.StringVar(&AppArgs.LibName, "libraryName", AppArgs.LibName,
		"Name of this Library (shown on every page)\n")

	if s, ok = appArguments.AsString("libraryPath"); ok && (0 < len(s)) {
		AppArgs.LibPath, _ = filepath.Abs(s)
	} else {
		AppArgs.LibPath = `/var/opt/Calibre`
	}
	flag.CommandLine.StringVar(&AppArgs.LibPath, "libraryPath", AppArgs.LibPath,
		"<pathname> Path name of/to the Calibre library\n")

	if AppArgs.Listen, ok = appArguments.AsString("listen"); (!ok) || (0 == len(AppArgs.Listen)) {
		AppArgs.Listen = `127.0.0.1`
	}
	flag.CommandLine.StringVar(&AppArgs.Listen, "listen", AppArgs.Listen,
		"the host's IP to listen at ")

	AppArgs.LogStack, _ = appArguments.AsBool("logStack")
	flag.CommandLine.BoolVar(&AppArgs.LogStack, "logStack", AppArgs.LogStack,
		"<boolean> Log a stack trace for recovered runtime errors ")

	if AppArgs.Port, ok = appArguments.AsInt("port"); (!ok) || (0 == AppArgs.Port) {
		AppArgs.Port = 8383
	}
	flag.CommandLine.IntVar(&AppArgs.Port, "port", AppArgs.Port,
		"<portNumber> The IP port to listen to ")

	if AppArgs.Realm, ok = appArguments.AsString("realm"); (!ok) || (0 == len(AppArgs.Realm)) {
		AppArgs.Realm = `eBooks Host`
	}
	flag.CommandLine.StringVar(&AppArgs.Realm, "realm", AppArgs.Realm,
		"<hostName> Name of host/domain to secure by BasicAuth\n")

	if AppArgs.SessionTTL, ok = appArguments.AsInt("sessionTTL"); (!ok) || (0 == AppArgs.SessionTTL) {
		AppArgs.SessionTTL = 1200
	}
	flag.CommandLine.IntVar(&AppArgs.SessionTTL, "sessionTTL", AppArgs.SessionTTL,
		"<seconds> Number of seconds an unused session keeps valid ")

	if AppArgs.SidName, ok = appArguments.AsString("sidName"); (!ok) || (0 == len(AppArgs.SidName)) {
		AppArgs.SidName = `sid`
	}
	flag.CommandLine.StringVar(&AppArgs.SidName, "sidName", AppArgs.SidName,
		"<name> The name of the session ID to use\n")

	if s, ok = appArguments.AsString("sqlTrace"); ok && (0 < len(AppArgs.SQLTrace)) {
		AppArgs.SQLTrace = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.SQLTrace, "sqlTrace", AppArgs.SQLTrace,
		"<filename> Name of the SQL logfile to write to\n")

	if AppArgs.Theme, _ = appArguments.AsString("theme"); 0 < len(AppArgs.Theme) {
		AppArgs.Theme = strings.ToLower(AppArgs.Theme)
	}
	switch AppArgs.Theme {
	case `dark`, `light`:
	default:
		AppArgs.Theme = `dark`
	}
	flag.CommandLine.StringVar(&AppArgs.Theme, "theme", AppArgs.Theme,
		"<name> The display theme to use ('light' or 'dark')\n")

	flag.CommandLine.StringVar(&AppArgs.UserAdd, "ua", AppArgs.UserAdd,
		"<userName> User add: add a username to the password file")

	flag.CommandLine.StringVar(&AppArgs.UserCheck, "uc", AppArgs.UserCheck,
		"<userName> User check: check a username in the password file")

	flag.CommandLine.StringVar(&AppArgs.UserDelete, "ud", AppArgs.UserDelete,
		"<userName> User delete: remove a username from the password file")

	if s, ok = appArguments.AsString("passFile"); ok && (0 < len(s)) {
		AppArgs.PassFile = absolute(AppArgs.DataDir, s)
	}
	flag.CommandLine.StringVar(&AppArgs.PassFile, "uf", AppArgs.PassFile,
		"<fileName> Passwords file storing user/passwords for BasicAuth\n")

	flag.CommandLine.BoolVar(&AppArgs.UserList, "ul", AppArgs.UserList,
		"<boolean> User list: show all users in the password file")

	flag.CommandLine.StringVar(&AppArgs.UserUpdate, "uu", AppArgs.UserUpdate,
		"<userName> User update: update a username in the password file")

	appArguments.Clear() // release unneeded memory
} // setFlags()

// ShowHelp lists the commandline options to `Stderr`.
func ShowHelp() {
	fmt.Fprintf(os.Stderr, "\n  Usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.CommandLine.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\n  Most options can be set in an INI file to keep the command-line short ;-)")
} // ShowHelp()

/* _EoF_ */
