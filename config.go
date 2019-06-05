/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwat56/ini"
)

type (
	// tAguments is the list structure for the cmdline argument values
	// merged with the key-value pairs from the INI file.
	tAguments struct {
		ini.TSection // embedded INI section
	}
)

var (
	// AppArguments is the list for the cmdline arguments and INI values.
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
func (al *tAguments) Get(aKey string) (string, error) {
	if result, ok := al.AsString(aKey); ok {
		return result, nil
	}

	return "", fmt.Errorf("Missing config value: %s", aKey)
} // Get()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

const (
	// Name of the `Calibre` database
	calibreDatabaseName = "metadata.db"
)

var (
	// Pathname to the `Calibre` database
	calibreLibraryPath = ""
)

// CalibreDatabaseName returns the name of the `Calibre` database.
func CalibreDatabaseName() string {
	return calibreDatabaseName
} // CalibreDatabaseName()

// CalibreLibraryPath returns the base directory of the `Calibre` library.
func CalibreLibraryPath() string {
	return calibreLibraryPath
} // CalibreLibraryPath()

// SetCalibreLibraryPath sets the base directory of the `Calibre` library.
func SetCalibreLibraryPath(aPath string) string {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if _, err := os.Stat(aPath); nil == err {
		calibreLibraryPath = aPath
	}

	return calibreLibraryPath
} // CalibreLibraryPath()

// CalibreDatabasePath returns rhe complete path-/filename of the `Calibre` library.
func CalibreDatabasePath() string {
	return filepath.Join(calibreLibraryPath, calibreDatabaseName)
} // CalibreDatabasePath()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func initConfig() {

	/*
		We need:
		- calibreLibraryPath
		- calibreDatabaseName
		- (SQL) LIMIT (rows per page)
		- View: List _or_ Grid

	*/

} // initConfig()

// ShowHelp lists the commandline options to `Stderr`.
func ShowHelp() {
	fmt.Fprintf(os.Stderr, "\n  Usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\n  Most options can be set in an INI file to keep the command-line short ;-)")
} // ShowHelp()

/* _EoF_ */
