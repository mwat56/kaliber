/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

const (
	// Name of the `Calibre` database
	dbCalibreDatabaseFilename = `metadata.db`

	// Calibre's metadata/preferences store
	dbCalibrePreferencesFile = `metadata_db_prefs_backup.json`
)

type (
	// TDataBase An opaque structure providing the properties and
	// methods to access the `Calibre` database.
	TDataBase struct {
		sqlDB *sql.DB // the embedded database connection
	}
)

var (
	// Pathname to the cached `Calibre` database
	dbCalibreCachePath = ``

	// Pathname to the original `Calibre` database
	dbCalibreLibraryPath = ``

	// // The active `tDatabase` instance initialised by `OpenDatabase()`.
	// dbSqliteDB TDataBase
)

// CalibreCachePath returns the directory of the copied `Calibre` database.
func CalibreCachePath() string {
	return dbCalibreCachePath
} // CalibreCachePath()

// CalibreLibraryPath returns the base directory of the `Calibre` library.
func CalibreLibraryPath() string {
	return dbCalibreLibraryPath
} // CalibreLibraryPath()

// CalibrePreferencesFile returns the complete path-/filename of the
// `Calibre` library's preferences file.
func CalibrePreferencesFile() string {
	return filepath.Join(dbCalibreLibraryPath, dbCalibrePreferencesFile)
} // CalibrePreferencesFile()

// OpenDatabase establishes a new database connection.
//
//	`aContext` The current HTTP request's context.
func OpenDatabase(aContext context.Context) (rDB *TDataBase, rErr error) {
	// Prepare the local database copy:
	if _, rErr = copyDatabaseFile(); nil != rErr {
		return
	}
	// Signal for `rDB.reOpen()`:
	syncCopiedChan <- struct{}{}

	// Start monitoring the original database file:
	go goCheckFile(syncCheckChan, syncCopiedChan)

	rDB = &TDataBase{}
	rErr = rDB.reOpen(aContext)

	return
} // OpenDatabase()

// SetCalibreCachePath sets the directory of the `Calibre` database copy.
//
//	`aPath` is the directory path to use for caching the Calibre library.
func SetCalibreCachePath(aPath string) {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		dbCalibreCachePath = aPath
	} else if err := os.MkdirAll(aPath, os.ModeDir|0775); nil == err {
		dbCalibreCachePath = aPath
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
		dbCalibreLibraryPath = aPath
	} else {
		dbCalibreLibraryPath = ``
	}

	return dbCalibreLibraryPath
} // SetCalibreLibraryPath()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// `dbHaving` defines a JOIN operation limiting the result-set
	// to records matching a certain condition.
	dbHaving = map[string]string{
		`all`:       ``,
		`authors`:   `JOIN books_authors_link a ON(a.book = b.id) WHERE (a.author = %d) `,
		`format`:    `JOIN data d ON(b.id = d.book) JOIN data dd ON (d.format = dd.format) WHERE (dd.id = %d) `,
		`languages`: `JOIN books_languages_link l ON(l.book = b.id) WHERE (l.lang_code = %d) `,
		`publisher`: `JOIN books_publishers_link p ON(p.book = b.id) WHERE (p.publisher = %d) `,
		`series`:    `JOIN books_series_link s ON(s.book = b.id) WHERE (s.series = %d) `,
		`tags`:      `JOIN books_tags_link t ON(t.book = b.id) WHERE (t.tag = %d) `,
	}
)

// `having()` returns a string limiting the query to the given `aEntity`
// with `aID`.
func having(aEntity string, aID TID) string {
	if (0 == len(aEntity)) || (`all` == aEntity) || (0 == aID) {
		return ``
	}

	return fmt.Sprintf(dbHaving[aEntity], aID)
} // having()

// `limit()` returns a LIMIT clause defined by `aStart` and `aLength`.
func limit(aStart, aLength uint) string {
	return `LIMIT ` + strconv.FormatInt(int64(aStart), 10) +
		`,` + strconv.FormatInt(int64(aLength), 10)
} // limit()

// `orderBy()` returns a ORDER_BY clause defined by `aOrder` and `aDesc`.
//
// The `aOrder` argument can be one of the following constants:
//
//	qoSortByAcquisition = TSortType(iota)
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
//	`aDescending` If `true` the query result is sorted in DESCending order.
func orderBy(aOrder TSortType, aDescending bool) string {
	desc := `` // ` ASC ` is default
	if aDescending {
		desc = ` DESC`
	}
	var result string
	switch aOrder { // constants defined in `queryoptions.go`
	case qoSortByAcquisition:
		result = `b.timestamp` + desc + `, b.pubdate` + desc + `, b.author_sort`
	case qoSortByAuthor:
		result = `b.author_sort` + desc + `, b.pubdate`
	case qoSortByLanguage:
		result = `languages` + desc + `, b.author_sort` + desc + `, b.sort`
	case qoSortByPublisher:
		result = `publisher` + desc + `, b.author_sort` + desc + `, b.sort`
	case qoSortByRating:
		result = `rating` + desc + `, b.author_sort` + desc + `, b.sort`
	case qoSortBySeries:
		result = `series` + desc + `, b.series_index` + desc + `, b.sort`
	case qoSortBySize:
		result = `size` + desc + `, b.author_sort`
	case qoSortByTags:
		result = `tags` + desc + `, b.author_sort`
	case qoSortByTime:
		result = `b.pubdate` + desc + `, b.timestamp` + desc + `, b.author_sort`
	case qoSortByTitle:
		result = `b.sort` + desc + `, b.author_sort`
	default:
		return ``
	}

	return ` ORDER BY ` + result + desc + ` `
} // orderBy()

// `prepAuthors()` returns a sorted list of document authors.
//
//	`aAuthor` The document's author(s).
func prepAuthors(aAuthor tPSVstring) *tAuthorList {
	list := strings.Split(aAuthor, `, `)
	result := make(tAuthorList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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
//	`aFormat` The document format(s) available.
func prepFormats(aFormat tPSVstring) *tFormatList {
	list := strings.Split(aFormat, `, `)
	result := make(tFormatList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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
//	`aIdentifier` The document's identifier(s).
func prepIdentifiers(aIdentifier tPSVstring) *tIdentifierList {
	list := strings.Split(aIdentifier, `, `)
	result := make(tIdentifierList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least three indices:
		a := append(strings.SplitN(val, `|`, 3), ``)
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
//	`aLanguage` The document's language(s).
func prepLanguages(aLanguage tPSVstring) *tLanguageList {
	list := strings.Split(aLanguage, `, `)
	result := make(tLanguageList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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
	// RegEx to find a document's number of pages.
	dbPagesRE = regexp.MustCompile(`(?si)<meta name="calibre:user_metadata:#pages" .*?, &quot;#value#&quot;: (\d+),`)
)

// `prepPages()` prepares the document's `Pages` property.
//
// This functions only returns a value `>0` if the respective `pages`
// plugin is installed with `Calibre` _and_ it uses internally a data
// field called `#pages` stored in the document's metadata file.
//
//	`aPath` The relative directory/path of the document's data.
func prepPages(aPath string) int {
	fName := filepath.Join(dbCalibreLibraryPath, aPath, `metadata.opf`)
	if fi, err := os.Stat(fName); (nil != err) || (0 >= fi.Size()) {
		return 0
	}
	metadata, err := ioutil.ReadFile(fName) // #nosec G304
	if nil != err {
		return 0
	}
	match := dbPagesRE.FindSubmatch(metadata)
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
//	`aPublisher` The document's publisher(s).
func prepPublisher(aPublisher tPSVstring) *tPublisher {
	list := strings.Split(aPublisher, `, `)
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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
//	`aSeries` The document's series.
func prepSeries(aSeries tPSVstring) *tSeries {
	list := strings.Split(aSeries, `, `)
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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
//	`aTag` The document's tag(s).
func prepTags(aTag tPSVstring) *tTagList {
	list := strings.Split(aTag, `, `)
	result := make(tTagList, 0, len(list))
	for _, val := range list {
		if 0 == len(val) {
			continue
		}
		// make sure we have at least two indices:
		a := append(strings.SplitN(val, `|`, 2), ``)
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

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type (
	// A pipe separated value string.
	tPSVstring = string
)

// Close terminates the database connection.
func (db *TDataBase) Close() (rErr error) {
	if nil != db.sqlDB {
		rErr = db.sqlDB.Close()
		db.sqlDB = nil
		go goSQLtrace(`-- closed DB connection`, time.Now()) //FIXME REMOVE
	}

	return
} // Close()

// `doQueryAll()` returns a list of documents with all available fields
// and an `error` in case of problems.
//
//	`aContext` The current request's context.
//	`aQuery` The SQL query to run.
func (db *TDataBase) doQueryAll(aContext context.Context, aQuery string) (rList *TDocList, rErr error) {
	var rows *sql.Rows
	if rows, rErr = db.query(aContext, aQuery); nil != rErr {
		return
	}
	defer rows.Close()

	rList = NewDocList()
	for rows.Next() {
		var (
			authors, formats, identifiers, languages,
			publisher, series, tags tPSVstring
			noTime  time.Time
			visible bool
		)
		doc := NewDocument()
		_ = rows.Scan(&doc.ID, &doc.Title, &authors, &publisher,
			&doc.Rating, &doc.timestamp, &doc.Size, &tags,
			&doc.comments, &series, &doc.seriesindex,
			&doc.titleSort, &doc.authorSort, &formats, &languages,
			&doc.ISBN, &identifiers, &doc.path, &doc.lccn,
			&doc.pubdate, &doc.flags, &doc.uuid, &doc.hasCover)

		// check for (in)visible fields:
		if visible, _ = BookFieldVisible(`authors`); !visible {
			visible, _ = BookFieldVisible(`author_sort`)
		}
		if visible {
			doc.authors = prepAuthors(authors)
		}
		if visible, _ = BookFieldVisible(`comments`); !visible {
			doc.comments = ``
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
			doc.path = ``
		}
		if visible, _ = BookFieldVisible(`pubdate`); !visible {
			doc.pubdate = noTime
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
			doc.timestamp = noTime
		}
		if visible, _ = BookFieldVisible(`title`); !visible {
			visible, _ = BookFieldVisible(`sort`)
		}
		if !visible {
			doc.Title = ``
		}
		if visible, _ = BookFieldVisible(`size`); !visible {
			doc.Size = 0
		}
		if visible, _ = BookFieldVisible(`uuid`); !visible {
			doc.uuid = ``
		}

		select {
		case <-aContext.Done():
			rErr = aContext.Err()
			return
		default:
			rList.Add(doc)
		}
	}

	return
} // doQueryAll()

const (
	// `dbBaseQuery` is the default query to get all available document
	// data out of the `Calibre` library.
	// By appending `WHERE` and `LIMIT` clauses the result-set can be
	// restricted to a sub-set.
	//
	// see `QueryBy()`, `QueryDocument()`, `QuerySearch()`
	dbBaseQuery = `SELECT b.id,
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

	// see `QueryBy()`, `QuerySearch()`
	dbCountQuery = `SELECT COUNT(b.id) FROM books b `

	// `dbGridQuery` selects only those document fields which are used
	// by the `grid` layout.
	// By appending `WHERE` and `LIMIT` clauses the result-set can be
	// limited to a sub-set.
	//
	// see `QueryBy()`, `QuerySearch()`
	dbGridQuery = `SELECT b.id,
b.title,
IFNULL((SELECT group_concat(a.name || "|" || a.id, ", ")
	FROM authors a
	JOIN books_authors_link bal ON(bal.author = a.id)
	WHERE (bal.book = b.id)
), "") authors,
IFNULL((SELECT group_concat(l.lang_code || "|" || l.id, ", ")
	FROM books_languages_link bll
	JOIN languages l ON(bll.lang_code = l.id)
	WHERE (bll.book = b.id)
), "") languages,
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
IFNULL((SELECT group_concat(s.name || "|" || s.id, ", ")
	FROM series s
	JOIN books_series_link bsl ON(bsl.series = s.id)
	WHERE (bsl.book = b.id)
), "") series,
IFNULL((SELECT MAX(data.uncompressed_size)
	FROM data
	WHERE (data.book = b.id)
), 0) size,
IFNULL((SELECT group_concat(t.name || "|" || t.id, ", ")
	FROM tags t
	JOIN books_tags_link btl ON(btl.tag = t.id)
	WHERE (btl.book = b.id)
), "") tags,
b.pubdate,
b.sort AS title_sort
FROM books b `
)

// `doQueryGrid()` selects the data for a `grid` layout.
//
//	`aContext` The current web request's context.
//	`aQuery` The SQL query to run.
func (db *TDataBase) doQueryGrid(aContext context.Context, aQuery string) (rList *TDocList, rErr error) {
	var rows *sql.Rows
	if rows, rErr = db.query(aContext, aQuery); nil != rErr {
		return
	}
	defer rows.Close()

	rList = NewDocList()
	for rows.Next() {
		var (
			authors, languages, publisher, series, tags, titleSort tPSVstring
			rating, size                                           int
			pubdate                                                time.Time
			visible                                                bool
		)
		doc := NewDocument()

		_ = rows.Scan(&doc.ID, &doc.Title, &authors, &languages, &publisher, &rating, &series, &size, &tags, &pubdate, &titleSort)

		if visible, _ = BookFieldVisible(`authors`); !visible {
			_, _ = BookFieldVisible(`author_sort`)
		}
		doc.authors = prepAuthors(authors)

		select {
		case <-aContext.Done():
			rErr = aContext.Err()
			return
		default:
			rList.Add(doc)
		}
	}

	return
} // doQueryGrid()

// `query()` executes a query that returns rows, typically a SELECT.
// The `args` are for any placeholder parameters in the query.
//
//	`aContext` The current request's context.
//	`aQuery` The SQL query to run.
func (db *TDataBase) query(aContext context.Context, aQuery string) (rRows *sql.Rows, rErr error) {
	select {
	case <-aContext.Done():
		rErr = aContext.Err()
		return

	default:
		if rErr = db.reOpen(aContext); nil != rErr {
			return
		}
	}
	go goSQLtrace(aQuery, time.Now())

	rRows, rErr = db.sqlDB.QueryContext(aContext, aQuery)

	return
} // query()

// QueryBy returns all documents according to `aOptions`.
//
// The method returns in `rCount` the number of documents found,
// in `rList` either `nil` or a list list of documents,
// in `rErr` either `nil` or the error occurred during the search.
//
//	`aContext` The current web request's context.
//	`aOptions` The options to configure the query.
func (db *TDataBase) QueryBy(aContext context.Context, aOptions *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	var rows *sql.Rows
	rows, rErr = db.query(aContext,
		dbCountQuery+having(aOptions.Entity, aOptions.ID))
	if nil != rErr {
		return
	}
	defer rows.Close()

	if rows.Next() {
		_ = rows.Scan(&rCount)
	}

	select {
	case <-aContext.Done():
		rErr = aContext.Err()
	default:
		if 0 < rCount {
			if QoLayoutList == aOptions.Layout {
				rList, rErr = db.doQueryAll(aContext,
					dbBaseQuery+
						having(aOptions.Entity, aOptions.ID)+
						orderBy(aOptions.SortBy, aOptions.Descending)+
						limit(aOptions.LimitStart, aOptions.LimitLength))
			} else {
				rList, rErr = db.doQueryGrid(aContext,
					dbGridQuery+
						having(aOptions.Entity, aOptions.ID)+
						orderBy(aOptions.SortBy, aOptions.Descending)+
						limit(aOptions.LimitStart, aOptions.LimitLength))
			}
		}
	}

	return
} // QueryBy()

const (
	// see `QueryCustomColumns()`
	dbCustomColumnsQuery = `SELECT id, label, name, datatype FROM custom_columns `
)

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
//
//	`aContext` The current web request's context.
func (db *TDataBase) QueryCustomColumns(aContext context.Context) (*TCustomColumnList, error) {
	rows, err := db.query(aContext, dbCustomColumnsQuery)
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

		select {
		case <-aContext.Done():
			return nil, aContext.Err()
		default:
		}
	}

	return &result, nil
} // QueryCustomColumns()

const (
	// see `QueryDocMini()`
	dbDocMiniQuery = `SELECT b.id,
IFNULL((SELECT group_concat(d.format, ", ")
	FROM data d WHERE d.book = b.id), "") formats,
b.path,
b.title
FROM books b
WHERE b.id = `
)

// QueryDocMini returns the document identified by `aID`.
//
// This function fills only the document properties `ID`, `formats`,
// `path`, and `Title`.
// If a matching document could not be found the function returns `nil`.
//
//	`aContext` The current web request's context.
//	`aID` The document ID to lookup.
func (db *TDataBase) QueryDocMini(aContext context.Context, aID TID) (rDoc *TDocument) {
	rows, err := db.query(aContext,
		dbDocMiniQuery+strconv.FormatInt(int64(aID), 10))
	if nil != err {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var formats tPSVstring
		rDoc = NewDocument()
		rDoc.ID = aID
		_ = rows.Scan(&rDoc.ID, &formats, &rDoc.path, &rDoc.Title)
		rDoc.formats = prepFormats(formats)
	}

	return
} // QueryDocMini()

// QueryDocument returns the `TDocument` identified by `aID`.
//
// In case the document with `aID` can not be found the function
// returns `nil`.
//
//	`aContext` The current web request's context.
//	`aID` The document ID to lookup.
func (db *TDataBase) QueryDocument(aContext context.Context, aID TID) *TDocument {
	list, _ := db.doQueryAll(aContext, dbBaseQuery+
		`WHERE b.id=`+
		strconv.FormatInt(int64(aID), 10)+
		` LIMIT 1`)
	if 0 < len(*list) {
		doc := (*list)[0]

		return &doc
	}

	return nil
} // QueryDocument()

const (
	// see `QueryIDs()`
	dbIDQuery = `SELECT id, path FROM books `
)

// QueryIDs returns a list of documents with only the `ID` and
// `path` fields set.
//
// This method is used by `thumbnails`.
//
//	`aContext` The current web request's context.
func (db *TDataBase) QueryIDs(aContext context.Context) (rList *TDocList, rErr error) {
	var rows *sql.Rows
	rows, rErr = db.query(aContext, dbIDQuery)
	if nil != rErr {
		return
	}
	defer rows.Close()

	rList = NewDocList()
	for rows.Next() {
		doc := NewDocument()
		_ = rows.Scan(&doc.ID, &doc.path)

		select {
		case <-aContext.Done():
			rErr = aContext.Err()
			return
		default:
			rList.Add(doc)
		}
	}

	return
} // QueryIDs()

// QuerySearch returns a list of documents matching the criteria
// in `aOptions`.
//
// The function returns in `rCount` the number of documents found,
// in `rList` either `nil` or a list list of documents,
// in `rErr` either `nil` or an error occurred during the search.
//
//	`aContext` The current request's context.
//	`aOptions` The options to configure the query.
func (db *TDataBase) QuerySearch(aContext context.Context, aOptions *TQueryOptions) (rCount int, rList *TDocList, rErr error) {
	var rows *sql.Rows
	where := NewSearch(aOptions.Matching)
	if rows, rErr = db.query(aContext, dbCountQuery+where.Clause()); nil != rErr {
		return
	}
	defer rows.Close()

	if rows.Next() {
		_ = rows.Scan(&rCount)
	}

	select {
	case <-aContext.Done():
		rErr = aContext.Err()

	default:
		if 0 < rCount {
			if QoLayoutList == aOptions.Layout {
				rList, rErr = db.doQueryAll(aContext,
					dbBaseQuery+
						where.Clause()+
						orderBy(aOptions.SortBy, aOptions.Descending)+
						limit(aOptions.LimitStart, aOptions.LimitLength))
			} else {
				rList, rErr = db.doQueryGrid(aContext,
					dbGridQuery+
						where.Clause()+
						orderBy(aOptions.SortBy, aOptions.Descending)+
						limit(aOptions.LimitStart, aOptions.LimitLength))
			}
		}
	}

	return
} // QuerySearch()

// `reOpen()` checks whether the SQLite database file has changed
// since the last access.
// If so, the current database connection is closed and a new one
// is established.
//
//	`aContext` The current request's context.
func (db *TDataBase) reOpen(aContext context.Context) error {
	select {
	case _, more := <-syncCopiedChan:
		_ = db.Close()
		if !more {
			return nil // channel closed
		}
		var err error

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
			return aContext.Err()

		default:
			if db.sqlDB, err = sql.Open(`sqlite3`, dsn); nil != err {
				return err
			}
		}
		// db.sqlDB.Exec("PRAGMA xxx=yyy")

		go goSQLtrace(`-- reOpened `+dsn, time.Now()) //FIXME REMOVE
		return db.sqlDB.PingContext(aContext)

	default:
		return nil
	}
} // reOpen()

/* _EoF_ */
