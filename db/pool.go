/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * This file implements a simple connection pool to recycle
 * used connections.
 */

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
)

type (
	// A List of database connections.
	tDBlist []*sql.DB

	// TDBpool The list of database connections.
	TDBpool struct {
		pList tDBlist
		pMtx  *sync.Mutex
	}
)

var (
	// Pool The list of database connections.
	//
	// NOTE: This variable as such should be considered R/O.
	// To retrieve or store a certain connection use the `Get()`
	// and `Put()` methods respectively.
	Pool = &TDBpool{
		pList: make(tDBlist, 0, 63),
		pMtx:  new(sync.Mutex),
	}
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Clear empties the list.
//
// All connections are closed.
func (p *TDBpool) Clear() *TDBpool {
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	for idx, conn := range p.pList {
		if nil != conn {
			_ = conn.Close()
		}
		p.pList[idx] = nil
	}
	p.pList = p.pList[:0] // empty the list

	return p
} // Clear()

// Get selects a single database connection from the Pool, removes it
// from the Pool, and returns it to the caller.
//
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
//	`aContext` The current request's context.
func (p *TDBpool) Get(aContext context.Context) (rConn *sql.DB, rErr error) {
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	// There are 3 cases to consider:
	//
	// (1) the list is empty,
	// (2) the list has one entry,
	// (3) the list has more than one entry.
	//
	// We unroll these cases here to handle each case most efficiently.

	sLen := len(p.pList)
	if 0 == sLen { // case (1)
		rConn, rErr = p.open(aContext)
		return
	}

	select {
	case <-aContext.Done():
		rErr = aContext.Err()
		return

	default:
		rConn = p.pList[0]
		p.pList[0] = nil // erase element

		if 1 == sLen { // case (2)
			p.pList = p.pList[:0] // empty the list list
		} else { // case (3)
			p.pList = p.pList[1:] // remove first item from list
		}
	}

	return
} // Get()

// `open()` establishes a new database connection.
//
//	`aContext` The current request's context.
func (p *TDBpool) open(aContext context.Context) (rConn *sql.DB, rErr error) {
	//XXX Are there custom functions to inject?

	// `cache=shared` is essential to avoid running out of file
	// handles since each query seems to hold its own file handle.
	// `loc=auto` gets time.Time with current locale.
	// `mode=ro` is self-explanatory since we don't change the DB
	// in any way.
	dsn := `file:` +
		filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename) +
		`?cache=shared&case_sensitive_like=1&immutable=0&loc=auto&mode=ro&query_only=1`

	select {
	case <-aContext.Done():
		rErr = aContext.Err()
		return

	default:
		if rConn, rErr = sql.Open(`sqlite3`, dsn); nil != rErr {
			return
		}
	}
	// rConn.Exec("PRAGMA xxx=yyy")

	go goSQLtrace(`-- newOpen ` + dsn) //FIXME REMOVE
	rErr = rConn.PingContext(aContext)

	return
} // open()

// Put adds `aConnection` to the pool.
//
//	`aConnection` The database connection to add to the pool.
func (p *TDBpool) Put(aConnection *sql.DB) *TDBpool {
	if nil != aConnection {
		p.pMtx.Lock()
		defer p.pMtx.Unlock()

		p.pList = append(p.pList, aConnection)
	}

	return p
} // Put()

/* _EoF_ */
