/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

const (
	// `quCalibreBaseQuery` is the default query to get document data.
	// By appending `WHERE` and `LIMIT` clauses the resultset gets stinted.
	quCalibreBaseQuery = `SELECT b.id,
b.title,
IFNULL((SELECT group_concat(a.name || "|" || a.id, ", ")
	FROM authors a
	JOIN books_authors_link bal ON(bal.author = a.id)
	WHERE (bal.book = b.id)
), "") authors,
IFNULL((SELECT group_concat(p.name || "|" || p.id)
	FROM publishers p
	JOIN books_publishers_link bpl ON(p.id = bpl.publisher)
	WHERE (bpl.book = b.id)
), "") publisher,
IFNULL((SELECT r.rating
	FROM ratings r
	WHERE r.id IN (
		SELECT brl.rating
		from books_ratings_link brl
		WHERE (brl.book = b.id)
	)
), 0) rating,
b.timestamp,
IFNULL((SELECT MAX(data.uncompressed_size)
	FROM data
	WHERE (data.book = b.id)
), 0) size,
IFNULL((SELECT group_concat(t.name || "|" || t.id, ", ")
	FROM tags t
	JOIN books_tags_link btl ON(btl.tag = t.id)
	WHERE (btl.book = b.id)
), "") tags,
IFNULL((SELECT c.text
	FROM comments c
	WHERE (c.book = b.id)
), "") comments,
IFNULL((SELECT group_concat(s.name || "|" || s.id, ", ")
	FROM series s
	JOIN books_series_link bsl ON(bsl.series = s.id)
	WHERE (bsl.book = b.id)
), "") series,
b.series_index,
b.sort AS title_sort,
b.author_sort,
IFNULL((SELECT group_concat(d.format || "|" || d.id, ", ")
	FROM data d
	WHERE (d.book = b.id)
), "") formats,
IFNULL((SELECT group_concat(l.lang_code || "|" || l.id, ", ")
	FROM books_languages_link bll
	JOIN languages l ON(bll.lang_code = l.id)
	WHERE (bll.book = b.id)
), "") languages,
b.isbn,
IFNULL((SELECT group_concat(i.type || "|" || i.id || "|" || i.val, ", ")
	FROM identifiers i
	WHERE (i.book = b.id)
), "") identifiers,
b.path,
b.lccn,
b.pubdate,
b.flags,
b.uuid,
b.has_cover
FROM books b `

	// see `QueryBy()`. `QuerySearch()`
	quCalibreCountQuery = `SELECT COUNT(b.id) FROM books b `

	// see `()`
	quCalibreCustomColumnsQuery = `SELECT id, label, name, datatype FROM custom_columns`

	// see `QueryIDs()`
	quCalibreIDQuery = `SELECT id, path FROM books `

	// see `QueryDoc()`
	quCalibreMiniQuery = `SELECT b.id, IFNULL((SELECT group_concat(d.format, ", ")
FROM data d WHERE d.book = b.id), "") formats,
b.path,
b.title
FROM books b
WHERE b.id = %d`
)

var (
	quHaving = map[string]string{
		"all":       ``,
		"authors":   `JOIN books_authors_link a ON(a.book = b.id) WHERE (a.author = %d) `,
		"format":    `JOIN data d ON(b.id = d.book) JOIN data dd ON (d.format = dd.format) WHERE (dd.id = %d) `,
		"languages": `JOIN books_languages_link l ON(l.book = b.id) WHERE (l.lang_code = %d) `,
		"publisher": `JOIN books_publishers_link p ON(p.book = b.id) WHERE (p.publisher = %d) `,
		"series":    `JOIN books_series_link s ON(s.book = b.id) WHERE (s.series = %d) `,
		"tags":      `JOIN books_tags_link t ON(t.book = b.id) WHERE (t.tag = %d) `,
	}
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

const (
	// Name of the `Calibre` database
	quCalibreDatabaseFilename = "metadata.db"

	// Calibre's metadata/preferences store
	quCalibrePreferencesFile = "metadata_db_prefs_backup.json"
)

type (
	tDataBase struct {
		*sql.DB           // the embedded database connection
		dbFileName string // the SQLite database file
		doCheck    chan struct{}
		wasCopied  chan struct{}
	}

	// A comma separated value string
	tPSVstring = string
)

var (
	// Pathname to the cached `Calibre` database
	quCalibreCachePath = ""

	// Pathname to the original `Calibre` database
	quCalibreLibraryPath = ""

	// The active `tDatabase` instance initialised by `DBopen()`.
	quSqliteDB tDataBase

	// Optional file to log all SQL queries.
	quSQLTraceFile = ""
)

// CalibreCachePath returns the directory of the copied `Calibre` databse.
func CalibreCachePath() string {
	return quCalibreCachePath
} // CalibreCachePath()

// CalibreLibraryPath returns the base directory of the `Calibre` library.
func CalibreLibraryPath() string {
	return quCalibreLibraryPath
} // CalibreLibraryPath()

// CalibrePreferencesFile returns the complete path-/filename of the
// `Calibre` library's preferences file.
func CalibrePreferencesFile() string {
	return filepath.Join(quCalibreLibraryPath, quCalibrePreferencesFile)
} // CalibrePreferencesFile()

// SetCalibreCachePath sets the directory of the `Calibre` database copy.
//
//	`aPath` is the directory path to use for caching the Calibre library.
func SetCalibreCachePath(aPath string) {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		quCalibreCachePath = aPath
	} else if err := os.MkdirAll(aPath, os.ModeDir|0775); nil == err {
		quCalibreCachePath = aPath
	}
} // SetCalibreCachePath()

// SetCalibreLibraryPath sets the base directory of the `Calibre` library.
//
//	`aPath` is the directory path where the Calibre library resides.
func SetCalibreLibraryPath(aPath string) string {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		quCalibreLibraryPath = aPath
	} else {
		quCalibreLibraryPath = ""
	}

	return quCalibreLibraryPath
} // CalibreLibraryPath()

// CalibreDatabaseFile returns the complete path-/filename of the `Calibre` library.
func CalibreDatabaseFile() string {
	return filepath.Join(quCalibreLibraryPath, quCalibreDatabaseFilename)
} // CalibreDatabaseFile()

// SQLtraceFile returns the optional file used for logging all SQL queries.
func SQLtraceFile() string {
	return quSQLTraceFile
} // SQLtraceFile()

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		var doOnce sync.Once
		doOnce.Do(func() {
			if path, err := filepath.Abs(aFilename); nil == err {
				quSQLTraceFile = path
				// start the background writer:
				go goWriteSQLtrace(quSQLTraceFile, quSQLTraceChannel)
			}
		})

		return
	}

	quSQLTraceFile = ""
} // SetSQLtraceFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `dbReopen()` checks whether the SQLite database file has changed since
// the last access. If so, the current database connection is closed and
// a new one is established.
func (db *tDataBase) dbReopen() error {
	select {
	case _, more := <-db.wasCopied:
		if nil != db.DB {
			_ = db.DB.Close()
		}
		var err error
		if !more {
			return err // channel closed
		}
		// "cache=shared" is essential to avoid running out of file
		// handles since each query seems to hold its own file handle.
		// "mode=ro" is self-explanatory since we don't change the
		// DB in any way.
		dsn := `file:` + db.dbFileName + `?cache=shared&mode=ro&_case_sensitive_like=1&immutable=0&_query_only=1`
		if db.DB, err = sql.Open("sqlite3", dsn); nil != err {
			return err
		}
		go goSQLtrace("-- reOpened "+dsn, time.Now())
		return db.DB.Ping()

	default:
		return nil
	}
} // dbReopen()

// Query executes a query that returns rows, typically a SELECT.
// The `args` are for any placeholder parameters in the query.
func (db *tDataBase) Query(aQuery string, args ...interface{}) (*sql.Rows, error) {
	if err := db.dbReopen(); nil != err {
		return nil, err
	}
	go goSQLtrace(aQuery, time.Now())

	rows, err := db.DB.Query(aQuery, args...)
	db.doCheck <- struct{}{}

	return rows, err
} // Query()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `copyDatabaseFile()` copies Calibre's database file to our cache directory.
func copyDatabaseFile() (bool, error) {
	sName := filepath.Join(quCalibreLibraryPath, quCalibreDatabaseFilename)
	dName := filepath.Join(quCalibreCachePath, quCalibreDatabaseFilename)
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
	if sFI, err = os.Stat(sName); nil != err {
		return false, err
	}
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
	go goSQLtrace("-- copied "+sName+" to "+dName, time.Now())

	return true, os.Rename(tName, dName)
} // copyDatabaseFile()

// `doQueryAll()` returns a list of documents with all available fields.
func doQueryAll(aQuery string) (*TDocList, error) {
	rows, err := quSqliteDB.Query(aQuery)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	result := NewDocList()
	for rows.Next() {
		var (
			authors, formats, identifiers, languages,
			publisher, series, tags tPSVstring
			notime  time.Time
			visible bool
		)
		doc := NewDocument()
		_ = rows.Scan(&doc.ID, &doc.Title, &authors, &publisher,
			&doc.Rating, &doc.timestamp, &doc.Size, &tags,
			&doc.comments, &series, &doc.seriesindex,
			&doc.titleSort, &doc.authorSort, &formats, &languages,
			&doc.ISBN, &identifiers, &doc.path, &doc.lccn,
			&doc.pubdate, &doc.flags, &doc.uuid, &doc.hasCover)

		// check for (un)visible fields:
		if visible, _ = BookFieldVisible(`authors`); !visible {
			visible, _ = BookFieldVisible(`author_sort`)
		}
		if visible {
			doc.authors = prepAuthors(authors)
		}
		if visible, _ = BookFieldVisible(`comments`); !visible {
			doc.comments = ""
		}
		if visible, _ = BookFieldVisible(`formats`); visible {
			doc.formats = prepFormats(formats)
		}
		if visible, _ = BookFieldVisible(`identifiers`); visible {
			doc.identifiers = prepIdentifiers(identifiers)
		}
		if visible, _ = BookFieldVisible(`languages`); visible {
			doc.languages = prepLanguages(languages)
		}
		if visible, _ = BookFieldVisible(`#pages`); visible {
			doc.Pages = prepPages(doc.path)
		}
		if visible, _ = BookFieldVisible(`path`); !visible {
			doc.path = ""
		}
		if visible, _ = BookFieldVisible(`pubdate`); !visible {
			doc.pubdate = notime
		}
		if visible, _ = BookFieldVisible(`publisher`); visible {
			doc.publisher = prepPublisher(publisher)
		}
		if visible, _ = BookFieldVisible(`rating`); !visible {
			doc.Rating = 0
		}
		if visible, _ = BookFieldVisible(`series`); visible {
			doc.series = prepSeries(series)
		}
		if visible, _ = BookFieldVisible(`tags`); visible {
			doc.tags = prepTags(tags)
		}
		if visible, _ = BookFieldVisible(`timestamp`); !visible {
			doc.timestamp = notime
		}
		if visible, _ = BookFieldVisible(`title`); !visible {
			visible, _ = BookFieldVisible(`sort`)
		}
		if !visible {
			doc.Title = ""
		}
		if visible, _ = BookFieldVisible(`size`); !visible {
			doc.Size = 0
		}
		if visible, _ = BookFieldVisible(`uuid`); !visible {
			doc.uuid = ""
		}

		result.Add(doc)
	}

	return result, nil
} // doQueryAll()

// `escapeQuery()` returns a string with some characters escaped.
//
// see: https://github.com/golang/go/issues/18478#issuecomment-357285669
func escapeQuery(aSource string) string {
	sLen := len(aSource)
	if 0 == sLen {
		return ""
	}
	var (
		character byte
		escape    bool
		j         int
	)
	result := make([]byte, sLen<<1)
	for i := 0; i < sLen; i++ {
		character = aSource[i]
		switch character {
		case '\n', '\r', '\\', '"':
			// We do not escape the apostrophe since it can be
			// a legitimate part of the search term and we use
			// double quotes to enclose those terms.
			escape = true
		case '\032': // Ctrl-Z
			escape = true
			character = 'Z'
		default:
		}

		if escape {
			result[j] = '\\'
			result[j+1] = character
			escape = false
			j += 2
		} else {
			result[j] = character
			j++
		}
	}

	return string(result[0:j])
} // escapeQuery()

// `goCheckFile()` checks in background whether the original database
// file has changed. If so, that file is copied to the cache directory
// from where it is read and used by the `quSQLiteDB` instance.
func goCheckFile(aCheck <-chan struct{}, aCopied chan<- struct{}) {
	//lint:ignore S1000 – we only need the separate `more` field
	for {
		select {
		case _, more := <-aCheck:
			if !more {
				return // channel closed
			}
			if copied, err := copyDatabaseFile(); copied && (nil == err) {
				aCopied <- struct{}{}
			}
		}
	}
} // goCheckFile()

const (
	// Half a second to sleep in `goWrite()`.
	quHalfSecond = 500 * time.Millisecond
)

var (
	// The channel to send SQL to and read messages from
	quSQLTraceChannel = make(chan string, 64)
)

// `goSQLtrace()` runs in background to log `aQuery` (if a tracefile is set).
func goSQLtrace(aQuery string, aTime time.Time) {
	if 0 == len(quSQLTraceFile) {
		return
	}
	aQuery = strings.ReplaceAll(aQuery, "\t", " ")
	aQuery = strings.ReplaceAll(aQuery, "\n", " ")

	quSQLTraceChannel <- aTime.Format("2006-01-02 15:04:05 ") +
		strings.ReplaceAll(aQuery, "  ", " ")
} // goSQLtrace()

// `goWriteSQLtrace()` performs the actual file write.
//
// This function is run only once, handling all write requests
// in background.
//
//	`aLogfile` The name of the logfile to write to.
//	`aSource` The source of the log messages to write.
func goWriteSQLtrace(aLogfile string, aSource <-chan string) {
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

	// let the application initialise:
	time.Sleep(quHalfSecond)

	for { // wait for strings to write
		select {
		case txt, more = <-aSource:
			if !more { // channel closed
				log.Println("queries.goWrite(): message channel closed")
				return
			}
			if nil == file {
				if file, err = os.OpenFile(aLogfile,
					os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640); /* #nosec G302 */ nil != err {
					file = os.Stderr // a last resort
				}
			}
			fmt.Fprintln(file, txt)

			// let's handle waiting messages
			cCap := cap(aSource)
			for txt = range aSource {
				fmt.Fprintln(file, txt)
				cCap--
				if 0 == cCap {
					break // give a chance to close the file
				}
			}

		default:
			if nil == file {
				time.Sleep(quHalfSecond)
			} else {
				if os.Stderr != file {
					_ = file.Close()
				}
				file = nil
			}
		}
	}
} // goWriteSQLtrace()

// `having()` returns a string limiting the query to the given `aID`.
func having(aEntity string, aID TID) string {
	if (0 == len(aEntity)) || ("all" == aEntity) || (0 == aID) {
		return ""
	}

	return fmt.Sprintf(quHaving[aEntity], aID)
} // having()

// `limit()` returns a LIMIT clause defined by `aStart` and `aLength`.
func limit(aStart, aLength uint) string {
	return fmt.Sprintf("LIMIT %d, %d ", aStart, aLength)
} // limit()

// OpenDatabase establishes a new database connection.
func OpenDatabase() error {
	quSqliteDB.dbFileName = filepath.Join(quCalibreCachePath, quCalibreDatabaseFilename)
	quSqliteDB.doCheck = make(chan struct{}, 64)
	quSqliteDB.wasCopied = make(chan struct{}, 1)

	// prepare the local database copy:
	if _, err := copyDatabaseFile(); nil != err {
		return err
	}
	// signal for `dbReopen()`:
	quSqliteDB.wasCopied <- struct{}{}

	// start monitoring the original database file:
	go goCheckFile(quSqliteDB.doCheck, quSqliteDB.wasCopied)

	return quSqliteDB.dbReopen()
} // OpenDatabase()

// `orderBy()` returns a ORDER_BY clause defined by `aOrder` and `aDesc`.
//
// The `aOrder` argument can be one of the following constants:
//
//	qoSortUnsorted      = TSortType(iota)
//	qoSortByAuthor
//	qoSortByLanguage
//	qoSortByPublisher
//	qoSortByRating
//	qoSortBySeries
//	qoSortBySize
//	qoSortByTags
//	qoSortByTime
//	qoSortByTitle
//
//	`aDescending` if `true` the query result is sorted in DESCending order.
func orderBy(aOrder TSortType, aDescending bool) string {
	desc := "" // " ASC " is default
	if aDescending {
		desc = " DESC"
	}
	result := ""
	switch aOrder { // constants defined in `queryoptions.go`
	case qoSortByAcquisition:
		result = "b.timestamp" + desc + ", b.pubdate" + desc + ", b.author_sort"
	case qoSortByAuthor:
		result = "b.author_sort" + desc + ", b.pubdate"
	case qoSortByLanguage:
		result = "languages" + desc + ", b.author_sort" + desc + ", b.sort"
	case qoSortByPublisher:
		result = "publisher" + desc + ", b.author_sort" + desc + ", b.sort"
	case qoSortByRating:
		result = "rating" + desc + ", b.author_sort" + desc + ", b.sort"
	case qoSortBySeries:
		result = "series" + desc + ", b.series_index" + desc + ", b.sort"
	case qoSortBySize:
		result = "size" + desc + ", b.author_sort"
	case qoSortByTags:
		result = "tags" + desc + ", b.author_sort"
	case qoSortByTime:
		result = "b.pubdate" + desc + ", b.timestamp" + desc + ", b.author_sort"
	case qoSortByTitle:
		result = "b.sort" + desc + ", b.author_sort"
	default:
		return ""
	}

	return " ORDER BY " + result + desc + " "
} // orderBy()

// `prepAuthors()` returns a sorted list of document authors.
//
//	`aAuthor`
func prepAuthors(aAuthor tPSVstring) *tAuthorList {
	alist := strings.Split(aAuthor, ", ")
	result := make(tAuthorList, 0, len(alist))
	for _, val := range alist {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		result = append(result, TEntity{
			ID:   num,
			Name: a[0],
		})
	}
	if 1 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepAuthors()

// `prepFormats()` returns a sorted list of document formats.
//
//	`aFormat`
func prepFormats(aFormat tPSVstring) *tFormatList {
	list := strings.Split(aFormat, ", ")
	result := make(tFormatList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		result = append(result, TEntity{
			ID:   num,
			Name: a[0],
		})
	}
	if 1 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepFormats()

// `prepIdentifiers()` returns a sorted list of document identifiers.
//
//	`aIdentifier`
func prepIdentifiers(aIdentifier tPSVstring) *tIdentifierList {
	list := strings.Split(aIdentifier, ", ")
	result := make(tIdentifierList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least three indices:
		a := append(strings.SplitN(val, "|", 3), "")
		num, _ := strconv.Atoi(a[1])
		result = append(result, TEntity{
			ID:   num,
			Name: a[0],
			URL:  a[2],
		})
	}
	if 1 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepIdentifiers

// `prepLanguages()` returns a document's languages.
//
//	`aLanguage`
func prepLanguages(aLanguage tPSVstring) *tLanguageList {
	llist := strings.Split(aLanguage, ", ")
	result := make(tLanguageList, 0, len(llist))
	for _, val := range llist {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		result = append(result, TEntity{
			ID:   num,
			Name: a[0],
		})
	}
	if 1 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepLanguages()

var (
	// RegEx to find a document's number of pages
	quPagesRE = regexp.MustCompile(`(?si)<meta name="calibre:user_metadata:#pages" .*?, &quot;#value#&quot;: (\d+),`)
)

// `prepPages()` prepares the document's `Pages` property.
//
// This functions only returns a value `>0` if the respective `pages`
// plugin is installed with `Calibre` and it uses internally a data
// field called `#pages` stored in the documents's metadata file.
//
//	`aPath` is the directory/path of the document's data.
func prepPages(aPath string) int {
	fName := filepath.Join(quCalibreLibraryPath, aPath, "metadata.opf")
	if fi, err := os.Stat(fName); (nil != err) || (0 >= fi.Size()) {
		return 0
	}
	metadata, err := ioutil.ReadFile(fName) // #nosec G304
	if nil != err {
		return 0
	}
	match := quPagesRE.FindSubmatch(metadata)
	if (nil == match) || (1 > len(match)) {
		return 0
	}
	if num, _ := strconv.Atoi(string(match[1])); 0 < num {
		// Since the RegEx returns digits only we don't need
		// to check for errors here.
		return num
	}

	return 0
} // prepPages()

// `prepPublisher()` returns a document's publisher.
//
//	`aPublisher`
func prepPublisher(aPublisher tPSVstring) *tPublisher {
	list := strings.Split(aPublisher, ", ")
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		return &tPublisher{
			ID:   num,
			Name: a[0],
		}
	}

	return nil
} // prepPublisher()

// `prepSeries()` returns a document's series.
//
//	`aSeries`
func prepSeries(aSeries tPSVstring) *tSeries {
	list := strings.Split(aSeries, ", ")
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		return &tSeries{
			ID:   num,
			Name: a[0],
		}
	}

	return nil
} // prepSeries()

// `prepTags()` returns a sorted list of document tags.
//
//	`aTag`
func prepTags(aTag tPSVstring) *tTagList {
	list := strings.Split(aTag, ", ")
	result := make(tTagList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		result = append(result, TEntity{
			ID:   num,
			Name: a[0],
		})
	}
	if 1 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepTags()

// QueryBy returns all documents according to `aOptions`.
//
// The function returns in `rCount` the number of documents found,
// in `rList` either `nil` or a list list of documents,
// in `rErr` either `nil` or an error occurred during the search.
//
//	`aOptions` The options to configure the query.
func QueryBy(aOptions *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	if rows, err := quSqliteDB.Query(quCalibreCountQuery +
		having(aOptions.Entity, aOptions.ID)); nil == err {
		if rows.Next() {
			_ = rows.Scan(&rCount)
		}
		_ = rows.Close()
	}
	if 0 < rCount {
		rList, rErr = doQueryAll(quCalibreBaseQuery +
			having(aOptions.Entity, aOptions.ID) +
			orderBy(aOptions.SortBy, aOptions.Descending) +
			limit(aOptions.LimitStart, aOptions.LimitLength))
	}

	return
} // QueryBy()

type (
	// TCustomColumn contains info about a user-defined data field.
	TCustomColumn struct {
		ID                    int
		Label, Name, Datatype string
	}

	// TCustomColumnList is a list/slice of `TCustomColumnRec`.
	TCustomColumnList []TCustomColumn
)

// QueryCustomColumns returns data about user-defined columns in `Calibre`.
func QueryCustomColumns() (*TCustomColumnList, error) {
	rows, err := quSqliteDB.Query(quCalibreCustomColumnsQuery)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	result := make(TCustomColumnList, 0, 8)
	for rows.Next() {
		var cc TCustomColumn
		if err = rows.Scan(&cc.ID, &cc.Label, &cc.Name, &cc.Datatype); nil == err {
			result = append(result, cc)
		}
	}

	return &result, nil
} // QueryCustomColumns()

// QueryDocMini returns the document identified by `aID`.
//
// This function fills only the document fields `ID`, `formats`,
// `path`, and `Title`.
//
//	`aID` The document ID to lookup.
func QueryDocMini(aID TID) *TDocument {
	rows, err := quSqliteDB.Query(fmt.Sprintf(quCalibreMiniQuery, aID))
	if nil != err {
		return nil
	}
	defer rows.Close()

	if rows.Next() {
		var formats tPSVstring
		doc := NewDocument()
		doc.ID = aID
		_ = rows.Scan(&doc.ID, &formats, &doc.path, &doc.Title)
		doc.formats = prepFormats(formats)

		return doc
	}

	return nil
} // QueryDocMini()

// QueryDocument returns the `TDocument` identified by `aID`.
//
//	`aID` The document ID to lookup.
func QueryDocument(aID TID) *TDocument {
	list, _ := doQueryAll(quCalibreBaseQuery + fmt.Sprintf("WHERE b.id=%d ", aID))
	if 0 < len(*list) {
		doc := (*list)[0]

		return &doc
	}

	return nil
} // QueryDocument()

// QueryIDs returns a list of documents with only the `ID` and
// `path` fields set.
//
// This function is used by `thumbnails`.
func QueryIDs() (*TDocList, error) {
	rows, err := quSqliteDB.Query(quCalibreIDQuery)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	result := NewDocList()
	for rows.Next() {
		doc := NewDocument()
		_ = rows.Scan(&doc.ID, &doc.path)
		result.Add(doc)
	}

	return result, nil
} // QueryIDs()

// QuerySearch returns a list of documents.
//
// The function returns in `rCount` the number of documents found,
// in `rList` either `nil` or a list list of documents,
// in `rErr` either `nil` or an error occurred during the search.
//
//	`aOptions` The options to configure the query.
func QuerySearch(aOptions *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	where := NewSearch(aOptions.Matching)
	if rows, err := quSqliteDB.Query(quCalibreCountQuery + where.Clause()); nil == err {
		if rows.Next() {
			_ = rows.Scan(&rCount)
		}
		_ = rows.Close()
	}
	if 0 < rCount {
		rList, rErr = doQueryAll(quCalibreBaseQuery +
			where.Clause() +
			orderBy(aOptions.SortBy, aOptions.Descending) +
			limit(aOptions.LimitStart, aOptions.LimitLength))
	}

	return
} // QuerySearch()

/* _EoF_ */
