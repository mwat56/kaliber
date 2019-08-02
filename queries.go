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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

const (
	// `calibreBaseQuery` is the default query to get document data.
	calibreBaseQuery = `SELECT b.id,
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
), "") language,
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
	calibreCountQuery = `SELECT COUNT(b.id) FROM books b `

	// see `QueryIDs()`
	calibreIDQuery = `SELECT id, path FROM books `

	// see `QueryDoc()`
	calibreMiniQuery = `SELECT b.id, IFNULL((SELECT group_concat(d.format, ", ")
FROM data d WHERE d.book = b.id), "") formats,
b.path,
b.title
FROM books b`
)

var (
	having = map[string]string{
		"all":       ``,
		"author":    `JOIN books_authors_link a ON(a.book = b.id) WHERE (a.author = %d) `,
		"format":    `JOIN data d ON(b.id = d.book) JOIN data dd ON (d.format = dd.format) WHERE (dd.id = %d) `,
		"lang":      `JOIN books_languages_link l ON(l.book = b.id) WHERE (l.lang_code = %d) `,
		"publisher": `JOIN books_publishers_link p ON(p.book = b.id) WHERE (p.publisher = %d) `,
		"series":    `JOIN books_series_link s ON(s.book = b.id) WHERE (s.series = %d) `,
		"tag":       `JOIN books_tags_link t ON(t.book = b.id) WHERE (t.tag = %d) `,
	}
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

const (
	// Name of the `Calibre` database
	calibreDatabaseName = "metadata.db"
)

type (
	tDataBase struct {
		*sql.DB           // the embedded database connection
		dbFileName string // the SQLite database file
		doCheck    chan bool
		isDone     chan bool
	}
)

var (
	// Pathname to the cached `Calibre` database
	calibreCachePath = ""

	// Pathname to the original `Calibre` database
	calibreLibraryPath = ""

	// The active `tDatabase` instance initialised by `DBopen()`.
	// sqliteDatabase *sql.DB
	sqliteDatabase tDataBase

	// Optional file to log all SQL queries.
	sqlTraceFile = ""
)

// CalibreCachePath returns the directory of the copied `Calibre` databse.
func CalibreCachePath() string {
	return calibreCachePath
} // CalibreCachePath()

// CalibreDatabaseName returns the name of the `Calibre` database.
func CalibreDatabaseName() string {
	return calibreDatabaseName
} // CalibreDatabaseName()

// CalibreLibraryPath returns the base directory of the `Calibre` library.
func CalibreLibraryPath() string {
	return calibreLibraryPath
} // CalibreLibraryPath()

// SetCalibreCachePath sets the directory of the `Calibre` database copy.
func SetCalibreCachePath(aPath string) string {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		calibreCachePath = aPath
	} else if err := os.MkdirAll(aPath, os.ModeDir|0775); nil == err {
		calibreCachePath = aPath
	}

	return calibreCachePath
} // SetCalibreCachePath()

// SetCalibreLibraryPath sets the base directory of the `Calibre` library.
func SetCalibreLibraryPath(aPath string) string {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		calibreLibraryPath = aPath
	} else {
		calibreLibraryPath = ""
	}

	return calibreLibraryPath
} // CalibreLibraryPath()

// CalibreDatabasePath returns rhe complete path-/filename of the `Calibre` library.
func CalibreDatabasePath() string {
	return filepath.Join(calibreLibraryPath, calibreDatabaseName)
} // CalibreDatabasePath()

// SQLtraceFile returns the optional file used for logging all SQL queries.
func SQLtraceFile() string {
	return sqlTraceFile
} // SQLtraceFile()

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		if path, err := filepath.Abs(aFilename); nil == err {
			sqlTraceFile = path
			return
		}
	}
	sqlTraceFile = ""
} //SetSQLtraceFile ()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `dbReopen()` checks whether the SQLite database file has changed
// since the last access. If so, the current database connection is
// closed and a new one is established.
//
func (db *tDataBase) dbReopen() error {
	// check whether the DB file changed:
	select {
	case <-db.isDone:
		if nil != db.DB {
			_ = db.DB.Close()
		}
		var err error
		// "cache=shared" is essential to avoid running out of
		// file handles since each query holds its own file handle.
		// "mode=ro" is self-explanatory since we don't change the
		// DB in any way.
		dsn := `file:` + db.dbFileName + `?cache=shared&mode=ro`
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
	db.doCheck <- true

	return rows, err
} // Query()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func copyDatabaseFile(aSrc, aDst string) error {
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
	if sFI, err = os.Stat(aSrc); nil != err {
		return err
	}
	if dFI, err = os.Stat(aDst); nil == err {
		if sFI.ModTime().Before(dFI.ModTime()) {
			return nil
		}
	}

	if sFile, err = os.Open(aSrc); /* #nosec G304 */ err != nil {
		return err
	}

	tName := aDst + `~`
	if tFile, err = os.OpenFile(tName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640); /* #nosec G302 */ err != nil {
		return err
	}

	if _, err = io.Copy(tFile, sFile); nil != err {
		return err
	}
	go goSQLtrace("-- copied "+aSrc+" to "+aDst, time.Now())

	return os.Rename(tName, aDst)
} // copyDatabaseFile()

// DBopen establishes a new database connection.
//
// `aFilename` is the path-/filename of the SQLite database to use.
func DBopen(aFilename string) error {
	sName := filepath.Join(calibreLibraryPath, calibreDatabaseName)
	dName := filepath.Join(calibreCachePath, calibreDatabaseName)
	sqliteDatabase.dbFileName = dName
	sqliteDatabase.doCheck = make(chan bool, 32)
	sqliteDatabase.isDone = make(chan bool, 32)

	// prepare the local database copy:
	if err := copyDatabaseFile(sName, dName); nil != err {
		return err
	}
	// signal for `dbReopen()`:
	sqliteDatabase.isDone <- true

	// start monitoring the original database file:
	go goCheckFile(sqliteDatabase.doCheck, sqliteDatabase.isDone)

	return sqliteDatabase.dbReopen()
} // DBopen()

// `doQueryAll()` returns a list of documents with all available fields.
func doQueryAll(aQuery string) (*TDocList, error) {
	rows, err := sqliteDatabase.Query(aQuery)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	result := newDocList()
	for rows.Next() {
		var authors, formats, identifiers, language,
			publisher, series, tags tCSVstring

		doc := newDocument()
		_ = rows.Scan(&doc.ID, &doc.Title, &authors, &publisher, &doc.Rating, &doc.timestamp, &doc.Size, &tags, &doc.comments, &series, &doc.seriesindex, &doc.TitleSort, &doc.authorSort, &formats, &language, &doc.ISBN, &identifiers, &doc.path, &doc.lccn, &doc.pubdate, &doc.flags, &doc.uuid, &doc.hasCover)
		doc.authors = prepAuthors(authors)
		doc.formats = prepFormats(formats)
		doc.identifiers = prepIdentifiers(identifiers)
		doc.language = prepLanguage(language)
		doc.Pages = prepPages(doc.path)
		doc.publisher = prepPublisher(publisher)
		doc.series = prepSeries(series)
		doc.tags = prepTags(tags)
		result.Add(doc)
	}

	return result, nil
} // doQueryAll()

// `goCheckFile()` checks in the background whether the original database
// file has changed. If so, that file is copied into the cache directory
// from where it is read and used by the `sqliteDatabase` instance.
func goCheckFile(aCheck <-chan bool, aDone chan<- bool) {
	sName := filepath.Join(calibreLibraryPath, calibreDatabaseName)
	dName := filepath.Join(calibreCachePath, calibreDatabaseName)

	for { // wait for a signal to arrive
		select {
		case <-aCheck:
			if err := copyDatabaseFile(sName, dName); nil == err {
				aDone <- true
			}
		default:
			time.Sleep(time.Second)
		}
	}
} // goCheckFile()

// `goSQLtrace()` runs in background to log `aQuery` (if a tracefile is set).
func goSQLtrace(aQuery string, aTime time.Time) {
	if 0 == len(sqlTraceFile) {
		return
	}
	aQuery = strings.ReplaceAll(aQuery, "\t", " ")
	aQuery = strings.ReplaceAll(aQuery, "\n", " ")
	file, err := os.OpenFile(sqlTraceFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640) // #nosec G302
	if nil != err {
		return
	}
	defer file.Close()

	fmt.Fprintln(file, aTime.Format("2006-01-02 15:04:05 ")+strings.ReplaceAll(aQuery, "  ", " "))
} // goSQLtrace()

// `havIng()` returns a string limiting the query to the given `aID`.
func havIng(aEntity string, aID TID) string {
	if (0 == len(aEntity)) || ("all" == aEntity) || (0 == aID) {
		return ""
	}

	return fmt.Sprintf(having[aEntity], aID)
} // havIng()

// `limit()` returns a LIMIT clause defined by `aStart` and `aLength`.
func limit(aStart, aLength uint) string {
	return fmt.Sprintf("LIMIT %d, %d ", aStart, aLength)
} // limit()

// `orderBy()` returns a ORDER_BY clause defined by `aOrder` and `aDesc`.
//
// The `aOrder` argument can be one of the following constants:
//
//	SortUnsorted = uint8(iota)
//	SortByAuthor
//	SortByLanguage
//	SortByPublisher
//	SortByRating
//	SortBySize
//	SortBySeries
//	SortByTags
//	SortByTime
//	SortByTitle
//
// `aDescending` if `true` the query result is sorted in DESCending order.
func orderBy(aOrder uint8, aDescending bool) string {
	desc := ""
	if aDescending {
		desc = " DESC"
	}
	result := "ORDER BY "
	switch aOrder { // constants defined in `queryoptions.go`
	case qoSortByAuthor:
		result += "b.author_sort" + desc + ", b.pubdate" + desc + " "
	case qoSortByLanguage:
		result += "language" + desc + ", b.author_sort" + desc + ", b.sort" + desc + " "
	case qoSortByPublisher:
		result += "publisher" + desc + ", b.author_sort" + desc + ", b.sort" + desc + " "
	case qoSortByRating:
		result += "rating" + desc + ", b.author_sort" + desc + ", b.sort" + desc + " "
	case qoSortBySeries:
		result += "series" + desc + ", b.series_index" + desc + ", b.sort" + desc + " "
	case qoSortBySize:
		result += "size" + desc + ", b.author_sort" + desc + " "
	case qoSortByTags:
		result += "tags" + desc + ", b.author_sort" + desc + " "
	case qoSortByTime:
		result += "b.timestamp" + desc + ", b.author_sort" + desc + " "
	case qoSortByTitle:
		result += "b.sort" + desc + ", b.author_sort" + desc + " "
	default:
		return ""
	}

	return result // " ASC " is default
} // orderBy()

func prepAuthors(aAuthor tCSVstring) *tAuthorList {
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
	if 0 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepAuthors()

func prepFormats(aFormat tCSVstring) *tFormatList {
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
	if 0 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepFormats()

func prepIdentifiers(aIdentifier tCSVstring) *tIdentifierList {
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
	if 0 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepIdentifiers

func prepLanguage(aLanguage tCSVstring) *tLanguage {
	list := strings.Split(aLanguage, ", ")
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, "|", 2), "")
		num, _ := strconv.Atoi(a[1])
		return &tLanguage{
			ID:   num,
			Name: a[0],
		}
	}

	return nil
} // prepLanguage()

func prepPages(aPath string) int {
	fName := filepath.Join(calibreLibraryPath, aPath, "metadata.opf")
	if fi, err := os.Stat(fName); (nil != err) || (0 >= fi.Size()) {
		return 0
	}
	metadata, err := ioutil.ReadFile(fName) // #nosec G304
	if nil != err {
		return 0
	}
	match := pagesRE.FindSubmatch(metadata)
	if (nil == match) || (1 > len(match)) {
		return 0
	}
	num, err := strconv.Atoi(string(match[1]))
	if nil != err {
		return 0
	}

	return num
} // prepPages()

func prepPublisher(aPublisher tCSVstring) *tPublisher {
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

func prepSeries(aSeries tCSVstring) *tSeries {
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

func prepTags(aTag tCSVstring) *tTagList {
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
	if 0 < len(result) {
		sort.Slice(result, func(i, j int) bool {
			return (result[i].Name < result[j].Name) // descending
		})
	}

	return &result
} // prepTags()

// QueryBy returns all documents according to `aOption`.
func QueryBy(aOption *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	if rows, err := sqliteDatabase.Query(calibreCountQuery +
		havIng(aOption.Entity, aOption.ID)); nil == err {
		if rows.Next() {
			_ = rows.Scan(&rCount)
		}
		_ = rows.Close()
	}
	if 0 < rCount {
		rList, rErr = doQueryAll(calibreBaseQuery +
			havIng(aOption.Entity, aOption.ID) +
			orderBy(aOption.SortBy, aOption.Descending) +
			limit(aOption.LimitStart, aOption.LimitLength))
	}

	return
} // QueryBy()

// QueryDocMini returns the document identified by `aID`.
//
// This function fills only the document fields `ID`, `formats`,
// `path`, and `Title`.
func QueryDocMini(aID TID) *TDocument {
	query := calibreMiniQuery + fmt.Sprintf(` WHERE b.id = %d`, aID)
	rows, err := sqliteDatabase.Query(query)
	if nil != err {
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		var formats tCSVstring
		doc := newDocument()
		doc.ID = aID
		_ = rows.Scan(&doc.ID, &formats, &doc.path, &doc.Title)
		doc.formats = prepFormats(formats)

		return doc
	}

	return nil
} // QueryDocMini()

// QueryDocument returns the `TDocument` identified by `aID`.
func QueryDocument(aID TID) *TDocument {
	list, _ := doQueryAll(calibreBaseQuery + fmt.Sprintf("WHERE b.id=%d ", aID))
	if 0 < len(*list) {
		doc := (*list)[0]

		return &doc
	}

	return nil
} // QueryDocument()

// QueryIDs returns a list of documents with only the `ID` and
// `path` fields set.
func QueryIDs() (*TDocList, error) {
	rows, err := sqliteDatabase.Query(calibreIDQuery)
	if nil != err {
		return nil, err
	}

	result := newDocList()
	for rows.Next() {
		doc := newDocument()
		_ = rows.Scan(&doc.ID, &doc.path)
		result.Add(doc)
	}

	return result, nil
} // QueryIDs()

// QueryLimit returns a list of `TDocument` objects.
func QueryLimit(aStart, aLength uint) (*TDocList, error) {
	return doQueryAll(calibreBaseQuery + limit(aStart, aLength))
} // QueryLimit()

// QuerySearch returns a list of documents
func QuerySearch(aOption *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	where := NewSearch(aOption.Matching)
	if rows, err := sqliteDatabase.Query(calibreCountQuery + where.Clause()); nil == err {
		if rows.Next() {
			_ = rows.Scan(&rCount)
		}
		_ = rows.Close()
	}
	if 0 < rCount {
		rList, rErr = doQueryAll(calibreBaseQuery +
			where.Clause() +
			orderBy(aOption.SortBy, aOption.Descending) +
			limit(aOption.LimitStart, aOption.LimitLength))
	}

	return
} // QuerySearch()

/* _EoF_ */
