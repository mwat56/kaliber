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
	// sqlite "github.com/mattn/go-sqlite3"
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
			if copied, err := syncDatabaseFile(); copied && (nil == err) {
				aCopied <- struct{}{}
			}
			_ = timer.Reset(time.Minute)
		}
	}
} // goCheckFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// The channel to send SQL to and read trace messages from.
	syncSQLTraceChannel = make(chan string, 127)

	// Optional file to log all SQL queries.
	syncSQLTraceFile = ``
)

// `goSQLtrace()` runs in background to log `aQuery`
// (if a tracefile is set).
//
//	`aQuery` The SQL query to log.
func goSQLtrace(aQuery string) {
	if 0 == len(syncSQLTraceFile) {
		return
	}
	when := time.Now()
	aQuery = strings.Replace(aQuery, "\t", ` `, -1)
	aQuery = strings.Replace(aQuery, "\n", ` `, -1)

	syncSQLTraceChannel <- when.Format(`2006-01-02 15:04:05.000 `) +
		strings.Replace(aQuery, `  `, ` `, -1)
} // goSQLtrace()

const (
	threeSex = 3 * time.Second
)

var (
	// Mode of opening the logfile(s).
	syncOpenFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY | os.O_SYNC
)

// `goWriteSQLtrace()` performs the actual file writes.
//
// This function is called only once, handling all write requests
// while running in background.
//
//	`aSource` R/O channel to read the log messages to write.
func goWriteSQLtrace(aSource <-chan string) {
	var (
		err        error
		file       *os.File
		fileCloser *time.Timer
	)
	defer func() {
		if (nil != file) && (os.Stderr != file) {
			_ = file.Close()
		}
		if nil != fileCloser {
			_ = fileCloser.Stop()
		}
	}()

	// Let the application initialise:
	time.Sleep(threeSex)
	fileCloser = time.NewTimer(threeSex)

	for { // wait for strings to write
		select {
		case txt, more := <-aSource:
			if !more { // channel closed
				log.Println(`syncdb.goWriteSQLtrace(): trace channel closed`)
				return
			}
			if 0 < len(syncSQLTraceFile) {
				if nil == file {
					if file, err = os.OpenFile(syncSQLTraceFile,
						syncOpenFlags, 0640); /* #nosec G302 */ nil != err {
						file = os.Stderr // a last resort
					}
				}
				fmt.Fprintln(file, txt)
				fileCloser.Reset(threeSex)
			}

		case <-fileCloser.C:
			if nil != file {
				if os.Stderr != file {
					_ = file.Close()
				}
				file = nil
			}
			fileCloser.Reset(threeSex)
		}
	}
} // goWriteSQLtrace()

var (
	// Make sure the filename is set only once.
	syncFilenameOnce sync.Once
)

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		if path, err := filepath.Abs(aFilename); nil == err {
			syncSQLTraceFile = path
		}
		syncFilenameOnce.Do(func() {
			// start the background writer:
			go goWriteSQLtrace(syncSQLTraceChannel)
		})

		return
	}

	syncSQLTraceFile = ``
} // SetSQLtraceFile()

// SQLtraceFile returns the file used for the optional logging of SQL queries.
func SQLtraceFile() string {
	return syncSQLTraceFile
} // SQLtraceFile()

// `syncDatabaseFile()` copies Calibre's original database file
// to the configured cache directory.
func syncDatabaseFile() (bool, error) {
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
	go goSQLtrace(`-- copied ` + sName + ` to ` + dName)

	return true, os.Rename(tName, dName)
} // syncDatabaseFile()

/*
func syncBackupDataBase() (bool, error) {
	syncCopyMtx.Lock()
	defer syncCopyMtx.Unlock()

	var (
		backup           sqlite.SQLiteBackup
		dstConn, srcConn *sql.DB
		destConn         *sqlite.SQLiteConn
		done             bool
		err              error
		sFI, dFI         os.FileInfo
	)
	sName := filepath.Join(dbCalibreLibraryPath, dbCalibreDatabaseFilename)
	if sFI, err = os.Stat(sName); nil != err {
		return false, err
	}

	dName := filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
	if dFI, err = os.Stat(dName); nil == err {
		if sFI.ModTime().Before(dFI.ModTime()) {
			return false, nil
		}
	} // ELSE: the dest file doesn't exist yet

	if srcConn, err = sql.Open(`sqlite3`, `file:`+sName+`?cache=shared&mode=ro`); nil != err {
		return false, err
	}
	defer srcConn.Close()

	tName := dName + `~`
	if dstConn, err = sql.Open(`sqlite3`, `file:`+tName); nil != err {
		return false, err
	}
	defer dstConn.Close()

	if backup, err = dstConn.Backup(`main`, srcConn, `main`); nil != err {
		return false, err
	}
	defer backup.Close()

	if done, err = backup.Step(-1); nil != err {
		return false, err
	}

	if !done {
		return false, errors.New(`Problem backing up ` + sName + ` to ` + tName)
	}

	go goSQLtrace(`-- copied ` + sName + ` to ` + dName)

	return true, os.Rename(tName, dName)
} // syncBackupDataBase()
*/

/* _EoF_ */
