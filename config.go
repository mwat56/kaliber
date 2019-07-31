/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/mwat56/ini"
	"github.com/mwat56/sessions"
)

type (
	// tAguments is the list structure for the cmdline argument values
	// merged with the key-value pairs from the INI file.
	tAguments struct {
		ini.TSection // embedded INI section
	}
)

var (
	// AppArguments is the merged list for the cmdline arguments
	// and INI values for the application.
	AppArguments tAguments
)

// `set()` adds/sets another key-value pair.
//
// If `aValue` is empty gthen `aKey` gets removed.
func (al *tAguments) set(aKey, aValue string) {
	if 0 < len(aValue) {
		al.AddKey(aKey, aValue)
	} else {
		al.RemoveKey(aKey)
	}
} // set()

// Get returns the value associated with `aKey` and `nil` if found,
// or an empty string and an error.
//
//	`aKey` The requested value's key to lookup.
func (al *tAguments) Get(aKey string) (string, error) {
	if result, ok := al.AsString(aKey); ok {
		return result, nil
	}

	//lint:ignore ST1005 – capitalisation wanted
	return "", fmt.Errorf("Missing config value: %s", aKey)
} // Get()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `absolute()` return `aDir` as an absolute path
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

// `iniData()` returns the config values read from INI file(s).
// The steps here are:
// (1) read the global `/etc/kaliber.ini`,
// (2) read the user-local `~/.kaliber.ini`,
// (3) read the `-ini` commandline argument.
func iniData() {
	// (1) ./
	fName, _ := filepath.Abs("./kaliber.ini")
	ini1, err := ini.New(fName)
	if nil == err {
		ini1.AddSectionKey("", "inifile", fName)
	}

	// (2) /etc/
	fName = "/etc/kaliber.ini"
	if ini2, err := ini.New(fName); nil == err {
		ini1.Merge(ini2)
		ini1.AddSectionKey("", "inifile", fName)
	}

	// (3) ~user/
	if usr, err := user.Current(); nil != err {
		fName = os.Getenv("HOME")
	} else {
		fName = usr.HomeDir
	}
	if 0 < len(fName) {
		fName, _ = filepath.Abs(filepath.Join(fName, ".kaliber.ini"))
		if ini2, err := ini.New(fName); nil == err {
			ini1.Merge(ini2)
			ini1.AddSectionKey("", "inifile", fName)
		}
	}

	// (4) cmdline
	aLen := len(os.Args)
	for i := 1; i < aLen; i++ {
		if `-ini` == os.Args[i] {
			i++
			if i < aLen {
				fName, _ = filepath.Abs(os.Args[i])
				if ini2, _ := ini.New(fName); nil == err {
					ini1.Merge(ini2)
					ini1.AddSectionKey("", "inifile", fName)
				}
			}
			break
		}
	}

	AppArguments = tAguments{*ini1.GetSection("")}
} // iniData()

func init() {
	initArguments()
} // init()

// `initArguments()` reads the commandline arguments into a list
// structure merging it with key-value pairs read from an INI file.
//
// The steps here are:
// (1) read the INI file(s),
// (2) merge the commandline arguments the INI values
// into the global `AppArguments` variable.
func initArguments() {
	iniData()

	bppInt, _ := AppArguments.AsInt("booksperpage")
	flag.IntVar(&bppInt, "booksperpage", bppInt,
		"<number> the default number of books shown per page ")

	s, _ := AppArguments.Get("datadir")
	dataStr, _ := filepath.Abs(s)
	flag.StringVar(&dataStr, "datadir", dataStr,
		"<dirName> the directory with CACHE, CSS, IMG, and VIEWS sub-directories\n")

	s, _ = AppArguments.Get("certKey")
	ckStr := absolute(dataStr, s)
	flag.StringVar(&ckStr, "certKey", ckStr,
		"<fileName> the name of the TLS certificate key\n")

	s, _ = AppArguments.Get("certPem")
	cpStr := absolute(dataStr, s)
	flag.StringVar(&cpStr, "certPem", cpStr,
		"<fileName> the name of the TLS certificate PEM\n")

	/*
		s, _ = AppArguments.Get("intl")
		intlStr := absolute(dataStr, s)
		flag.StringVar(&intlStr, "intl", intlStr,
			"<fileName> the path/filename of the localisation file\n")
	*/

	iniStr, _ := AppArguments.Get("inifile")
	flag.StringVar(&iniStr, "ini", iniStr,
		"<fileName> the path/filename of the INI file to use\n")

	langStr, _ := AppArguments.Get("lang")
	flag.StringVar(&langStr, "lang", langStr,
		"(optional) the default language to use ")

	lnStr, _ := AppArguments.Get("libraryname")
	flag.StringVar(&lnStr, "libraryname", lnStr,
		"Name of this Library (shown on every page)\n")

	lpStr, _ := AppArguments.Get("librarypath")
	flag.StringVar(&lpStr, "librarypath", lpStr,
		"Path name of/to the Calibre library\n")

	listenStr, _ := AppArguments.Get("listen")
	flag.StringVar(&listenStr, "listen", listenStr,
		"the host's IP to listen at ")

	s, _ = AppArguments.Get("logfile")
	logStr := absolute(dataStr, s)
	flag.StringVar(&logStr, "log", logStr,
		"(optional) name of the logfile to write to\n")

	/*
		ndBool := false
		flag.BoolVar(&ndBool, "nd", ndBool,
			"(optional) no daemon: whether to not daemonise the program")
	*/

	portInt, _ := AppArguments.AsInt("port")
	flag.IntVar(&portInt, "port", portInt,
		"<portNumber> the IP port to listen to ")

	realStr, _ := AppArguments.Get("realm")
	flag.StringVar(&realStr, "realm", realStr,
		"(optional) <hostName> name of host/domain to secure by BasicAuth\n")

	s, _ = AppArguments.Get("sessiondir")
	sessStr := absolute(dataStr, s)
	flag.StringVar(&sessStr, "sessiondir", sessStr,
		"<directory> (optional) the directory to store session files\n")

	sidStr, _ := AppArguments.Get("sidname")
	flag.StringVar(&sidStr, "sidname", sidStr,
		"(optional) <name> the name of the session ID to use\n")

	themStr, _ := AppArguments.Get("theme")
	flag.StringVar(&themStr, "theme", themStr,
		"<name> the display theme to use ('light' or 'dark')\n")

	uaStr := ""
	flag.StringVar(&uaStr, "ua", uaStr,
		"<userName> (optional) user add: add a username to the password file")

	ucStr := ""
	flag.StringVar(&ucStr, "uc", ucStr,
		"<userName> (optional) user check: check a username in the password file")

	udStr := ""
	flag.StringVar(&udStr, "ud", udStr,
		"<userName> (optional) user delete: remove a username from the password file")

	s, _ = AppArguments.Get("passfile")
	ufStr := absolute(dataStr, s)
	flag.StringVar(&ufStr, "uf", ufStr,
		"<fileName> (optional) user passwords file storing user/passwords for BasicAuth\n")

	ulBool := false
	flag.BoolVar(&ulBool, "ul", ulBool,
		"(optional) user list: show all users in the password file")

	uuStr := ""
	flag.StringVar(&uuStr, "uu", uuStr,
		"<userName> (optional) user update: update a username in the password file")

	flag.Usage = ShowHelp
	flag.Parse() // // // // // // // // // // // // // // // // // // //

	AppArguments.set("booksperpage", fmt.Sprintf("%d", bppInt))

	if 0 < len(dataStr) {
		dataStr, _ = filepath.Abs(dataStr)
	}
	if f, err := os.Stat(dataStr); nil != err {
		log.Fatalf("datadir == %s` problem: %v", dataStr, err)
	} else if !f.IsDir() {
		log.Fatalf("Error: Not a directory `%s`", dataStr)
	}
	AppArguments.set("datadir", dataStr)

	if 0 < len(ckStr) {
		ckStr = absolute(dataStr, ckStr)
		if fi, err := os.Stat(ckStr); (nil != err) || (0 >= fi.Size()) {
			ckStr = ""
		}
	}
	AppArguments.set("certKey", ckStr)

	if 0 < len(cpStr) {
		cpStr = absolute(dataStr, cpStr)
		if fi, err := os.Stat(cpStr); (nil != err) || (0 >= fi.Size()) {
			cpStr = ""
		}
	}
	AppArguments.set("certPem", cpStr)

	/*
		if 0 <len(intlStr) {
			intlStr = absolute(dataStr, intlStr)
		}
		AppArguments.set("intl", intlStr)
	*/

	if 0 == len(langStr) {
		langStr = "en"
	}
	AppArguments.set("lang", langStr)

	if 0 == len(lnStr) {
		lnStr = time.Now().Format("2006:01:02:15:04:05")
	}
	AppArguments.set("libraryname", lnStr)

	SetCalibreCachePath(filepath.Join(dataStr, "img"))
	SetCalibreLibraryPath(lpStr)

	if "0" == listenStr {
		listenStr = ""
	}
	AppArguments.set("listen", listenStr)

	if 0 < len(logStr) {
		logStr = absolute(dataStr, logStr)
	}
	AppArguments.set("logfile", logStr)

	/*
		if ndBool {
			s = "true"
		} else {
			s = ""
		}
		AppArguments.set("nd", s)
	*/

	AppArguments.set("port", fmt.Sprintf("%d", portInt))

	AppArguments.set("realm", realStr)

	if 0 < len(sessStr) {
		logStr = absolute(dataStr, sessStr)
	}
	AppArguments.set("sessiondir", sessStr)
	sessions.SetSIDname(sidStr)

	AppArguments.set("theme", themStr)
	AppArguments.set("ua", uaStr)
	AppArguments.set("uc", ucStr)
	AppArguments.set("ud", udStr)

	if 0 < len(ufStr) {
		ufStr = absolute(dataStr, ufStr)
	}
	AppArguments.set("uf", ufStr)

	if ulBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("ul", s)

	AppArguments.set("uu", uuStr)
} // initArguments()

// ShowHelp lists the commandline options to `Stderr`.
func ShowHelp() {
	fmt.Fprintf(os.Stderr, "\n  Usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\n  Most options can be set in an INI file to keep the command-line short ;-)")
} // ShowHelp()

/* _EoF_ */
