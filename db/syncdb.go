/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
 * This file provides functions to sync the used database copy with the
 * original Calibre library.
 *
 * To avoid any LOCKing problems when reading the Calibre database
 * (which happened quite frequently) while it is edited by the original
 * Calibre installation here we simply copy Calibre's database file into
 * the user's cache directory.
 * This way we can use R/O access without the fear that the database might
 * be changed under our feet by other processes.
 *
 * Additionally there are functions to handle an external text file
 * for tracing all used SQL queries.
 */

var (
	// `syncCopiedChan` Signal channel for a new database copy.
	syncCopiedChan = make(chan struct{}, 1)

	// Guard against parallel database copies.
	syncCopyMtx = new(sync.Mutex)
)

// `copyDatabaseFile()` copies Calibre's original database file
// to our cache directory.
func copyDatabaseFile() (bool, error) {
	syncCopyMtx.Lock()
	defer syncCopyMtx.Unlock()

	var (
		err          error
		sFile, tFile *os.File
		sFI, dFI     os.FileInfo
	)
	defer func() {
		if nil != sFile {
			_ = sFile.Close()
		}
		if nil != tFile {
			_ = tFile.Close()
		}
	}()

	sName := filepath.Join(dbCalibreLibraryPath, dbCalibreDatabaseFilename)
	if sFI, err = os.Stat(sName); nil != err {
		return false, err
	}

	dName := filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
	if dFI, err = os.Stat(dName); nil == err {
		if sFI.ModTime().Before(dFI.ModTime()) {
			return false, nil
		}
	}

	if sFile, err = os.Open(sName); /* #nosec G304 */ err != nil {
		return false, err
	}

	tName := dName + `~`
	if tFile, err = os.OpenFile(tName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return false, err
	}

	if _, err = io.Copy(tFile, sFile); nil != err {
		return false, err
	}
	go goSQLtrace(`-- copied `+sName+` to `+dName, time.Now())

	return true, os.Rename(tName, dName)
} // copyDatabaseFile()

// `goCheckFile()` checks in background once a minute whether the
// original database file has changed.
// If so, that file is copied to the cache directory from where it is
// read and used by the `db.TDatabase` instance.
//
//	`aCopied` W/O channel to signal a new database copy.
func goCheckFile(aCopied chan<- struct{}) {
	timer := time.NewTimer(time.Minute)
	defer func() {
		_ = timer.Stop()
	}()

	//lint:ignore S1000 - We can't use `range` here
	for {
		select {
		case <-timer.C:
			if copied, err := copyDatabaseFile(); copied && (nil == err) {
				aCopied <- struct{}{}
			}
			_ = timer.Reset(time.Minute)
		}
	}
} // goCheckFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// The channel to send SQL to and read trace messages from.
	syncSQLTraceChannel = make(chan string, 64)

	// Optional file to log all SQL queries.
	syncSQLTraceFile = ``
)

// `goSQLtrace()` runs in background to log `aQuery`
// (if a tracefile is set).
//
//	`aQuery` The SQL query to log.
//	`aTime` The time at which the query was run.
func goSQLtrace(aQuery string, aTime time.Time) {
	if 0 == len(syncSQLTraceFile) {
		return
	}
	aQuery = strings.Replace(aQuery, "\t", ` `, -1)
	aQuery = strings.Replace(aQuery, "\n", ` `, -1)

	syncSQLTraceChannel <- aTime.Format(`2006-01-02 15:04:05 `) +
		strings.Replace(aQuery, `  `, ` `, -1)
} // goSQLtrace()

// `goWriteSQLtrace()` performs the actual file writes.
//
// This function is called only once, handling all write requests
// while running in background.
//
//	`aTraceLog` The name of the logfile to write to.
//	`aSource` R/O channel to read the log messages to write.
func goWriteSQLtrace(aTraceLog string, aSource <-chan string) {
	var (
		err  error
		file *os.File
		more bool
		txt  string
	)
	defer func() {
		if (nil != file) && (os.Stderr != file) {
			_ = file.Close()
		}
	}()

	// Interval to look for new messages to arrive.
	sleepDuration := time.Second * 3

	// Let the application initialise:
	time.Sleep(sleepDuration)

	for { // wait for strings to write
		select {
		case txt, more = <-aSource:
			if !more { // channel closed
				log.Println(`queries.goWriteSQLtrace(): message channel closed`)
				return
			}
			if nil == file {
				if file, err = os.OpenFile(aTraceLog,
					os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640); /* #nosec G302 */ nil != err {
					file = os.Stderr // a last resort
				}
			}
			fmt.Fprintln(file, txt)

		default:
			if nil == file {
				time.Sleep(sleepDuration)
			} else {
				if os.Stderr != file {
					_ = file.Close()
				}
				file = nil
			}
		}
	}
} // goWriteSQLtrace()

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		var once sync.Once
		once.Do(func() {
			if path, err := filepath.Abs(aFilename); nil == err {
				syncSQLTraceFile = path
				// start the background writer:
				go goWriteSQLtrace(syncSQLTraceFile, syncSQLTraceChannel)
			}
		})

		return
	}

	syncSQLTraceFile = ``
} // SetSQLtraceFile()

// SQLtraceFile returns the file used for the optional logging of SQL queries.
func SQLtraceFile() string {
	return syncSQLTraceFile
} // SQLtraceFile()

/* _EoF_ */
