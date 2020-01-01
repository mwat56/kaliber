/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	// Filename suffix for trimmes CSS files.
	cssNameSuffix = `.min`
)

type (
	// Simple struct embedding a `http.FileSystem` that
	// serves minified CSS file.
	tCSSFilesFilesystem struct {
		fs http.FileSystem
	}

	// Internal list of regular expressions used by
	// the `removeCSSwhitespace()` function.
	tCSSre struct {
		regEx   *regexp.Regexp
		replace string
	}
)

var (
	// RegExes to find whiteplace in a CSS file.
	cssREs = []tCSSre{
		{regexp.MustCompile(`(?s)\s*/\x2A.*?\x2A/\s*`), ` `}, /* comment */
		{regexp.MustCompile(`\s*([:;\{,+!])\s*`), `$1`},
		{regexp.MustCompile(`\s*\}\s*\}\s*`), `}}`},
		{regexp.MustCompile(`\s*;?\s*\}\s*`), `}`},
		{regexp.MustCompile(`^\s+`), ``},
		{regexp.MustCompile(`\s+$`), ``},
	}
)

// `createMinFile()` generates a minified version of `aCSSName`.
//
//	`aCSSName` The filename of the original CSS file.
//	`aMinName` The filename of the minified CSS file.
func createMinFile(aCSSName, aMinName string) error {
	css, err := ioutil.ReadFile(aCSSName) // #nosec G304
	if err != nil {
		return err
	}
	for _, re := range cssREs {
		css = re.regEx.ReplaceAll(css, []byte(re.replace))
	}

	return ioutil.WriteFile(aMinName, css, 0640)
} // createMinFile()

// `removeCSSwhitespace()` removes unneeded whitespace from `aCSS`.
//
//	`aCSS` The raw CSS data.
func removeCSSwhitespace(aCSS []byte) []byte {
	for _, re := range cssREs {
		aCSS = re.regEx.ReplaceAll(aCSS, []byte(re.replace))
	}

	return aCSS
} // removeCSSwhitespace()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Open is a wrapper around the `Open()` method of the embedded FileSystem
// that returns a `http.File` containing a minified CSS file.
//
//	`aName` is the name of the CSS file to open.
func (cf tCSSFilesFilesystem) Open(aName string) (http.File, error) {
	cName, _ := filepath.Abs(aName)
	mName := cName + cssNameSuffix

	mInfo, err := os.Stat(mName)
	if (nil != err) || (0 == mInfo.Size()) {
		createMinFile(cName, mName)
		return cf.fs.Open(mName)
	}

	cInfo, err := os.Stat(cName)
	if nil != err {
		return nil, err
	}
	if mTime := mInfo.ModTime(); mTime.Before(cInfo.ModTime()) {
		createMinFile(cName, mName)
	}

	return cf.fs.Open(mName)
} // Open()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// FileServer returns a handler that serves HTTP requests
// with the contents of the file system rooted at `aRoot`.
//
// To use the operating system's file system implementation,
// use `http.Dir()`:
//
//	myHandler := http.FileServer(http.Dir("/tmp")))
//
// To use this implementation you'd use:
//
//	myHandler := css.FileServer(http.Dir("/tmp")))
//
//	`aRoot` The root of the filesystem to serve.
func FileServer(aRoot http.FileSystem) http.Handler {
	return http.FileServer(tCSSFilesFilesystem{aRoot})
} // FileServer()

/* _EoF_ */
