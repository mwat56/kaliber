/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"database/sql"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

// Constants defining the ORDER_BY clause
const (
	SortUnsorted = uint8(iota)
	SortByAuthor
	SortByLanguage
	SortByPublisher
	SortByRating
	SortBySize
	SortBySeries
	SortByTags
	SortByTime
	SortByTitle
)

const (
	// `baseQueryString` is the default query to get document data.
	baseQueryString = `SELECT b.id,
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
)

var (
	having = map[string]string{
		"author":    `, books_authors_link a WHERE ((a.book = b.id) AND (a.author = %d)) `,
		"lang":      `, books_languages_link l WHERE ((l.book = b.id) AND (l.lang_code = %d))`,
		"publisher": `, books_publishers_link p WHERE ((p.book = b.id) AND (p.publisher = %d)) `,
		"series":    `, books_series_link s WHERE ((s.book = b.id) AND (s.series = %d)) `,
		"tag":       `, books_tags_link t WHERE ((t.book = b.id) AND (t.tag = %d)) `,
	}
)

type (
	// TQueryOptions holds properties configuring a query.
	//
	// This type is used by the HTTP pagehandler when receiving
	// a FORM's data.
	TQueryOptions struct {
		ID          TID    // an entity ID to lookup
		Descending  bool   // sort direction
		Entity      string // for limiting to a certin entity (author,publisher, series, tag)
		LimitLength uint   // number of documents per page
		LimitStart  uint   // starting number
		Matching    string // text to lookup in all documents
		SortBy      uint8  // display order of documents
	}
)

// NewQueryOptions returns a new `TQueryOptions` instance.
func NewQueryOptions() *TQueryOptions {
	result := TQueryOptions{
		LimitLength: 25,
		SortBy:      SortByTime,
		Descending:  true,
	}

	return &result
} // NewQueryOptions()

// CGI returns the object's query escaped string representation
// fit for use as the `qos` CGI argument.
func (qo *TQueryOptions) CGI() string {
	return "?qos=" + url.QueryEscape(qo.String())
} // CGI()

// String returns the options as a `|` delimited string.
func (qo *TQueryOptions) String() string {
	return fmt.Sprintf("%d|%v|%d|%d|%s|%d|%s", qo.ID, qo.Descending, qo.LimitLength, qo.LimitStart, "", qo.SortBy, qo.Entity)
} // String()

// Scan returns the options read from `aString`.
func (qo *TQueryOptions) Scan(aString string) *TQueryOptions {
	var desc string
	fmt.Sscanf(aString, "%d|%v|%d|%d|%s|%d|%s", &qo.ID, &desc, &qo.LimitLength, &qo.LimitStart, &qo.Matching, &qo.SortBy, qo.Entity)
	if "true" == desc {
		qo.Descending = true
	}

	return qo
} // Scan()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// The active Database instance set by `DBopen()`.
	sqliteDatabase *sql.DB
)

// DBopen returns a new database connection.
//
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the Open
// function should be called just once. It is rarely necessary to
// close a DB.
//
// `aFilename` is the path-/filename of the SQLite database
func DBopen(aFilename string) error {
	var err error

	if sqliteDatabase, err = sql.Open("sqlite3", aFilename); nil != err {
		return err
	}

	return sqliteDatabase.Ping()
} // DBopen()

// `dbDocListQuery()` returns a list of documents.
func dbDocListQuery(aQuery string) (*TDocList, error) {
	rows, err := sqliteDatabase.Query(aQuery)
	if nil != err {
		return nil, err
	}

	result := newDocList()
	for rows.Next() {
		var authors, formats, identifiers, language, publisher, series, tags tCSVstring

		doc := newDocument()
		rows.Scan(&doc.ID, &doc.Title, &authors, &publisher, &doc.Rating, &doc.timestamp, &doc.Size, &tags, &doc.comments, &series, &doc.seriesindex, &doc.TitleSort, &doc.authorSort, &formats, &language, &doc.ISBN, &identifiers, &doc.path, &doc.lccn, &doc.pubdate, &doc.flags, &doc.uuid, &doc.hasCover)
		doc.authors = prepAuthors(authors)
		doc.formats = prepFormats(formats)
		doc.identifiers = prepIdentifiers(identifiers)
		doc.language = prepLanguage(language)
		doc.publisher = prepPublisher(publisher)
		doc.series = prepSeries(series)
		doc.tags = prepTags(tags)
		doc.setPages()

		*result = append(*result, *doc)
	}

	return result, nil
} // dbQuery()

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
	result := "ORDER BY "
	switch aOrder {
	case SortByAuthor:
		result += "b.author_sort, b.pubdate "
	case SortByLanguage:
		result += "language, b.author_sort, b.sort "
	case SortByPublisher:
		result += "publisher, b.author_sort, b.sort "
	case SortByRating:
		result += "rating, b.author_sort, b.sort "
	case SortBySeries:
		result += "series, b.series_index, b.sort "
	case SortBySize:
		result += "size, b.author_sort "
	case SortByTags:
		result += "tags, b.author_sort, b.sort "
	case SortByTime:
		result += "b.pubdate, b.sort "
	case SortByTitle:
		result += "b.sort, b.pubdate "
	default:
		return ""
	}
	if aDescending {
		return result + "DESC "
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

// QeueryBy returns all documents according to `aOption`.
func QeueryBy(aOption *TQueryOptions) (*TDocList, error) {
	query := baseQueryString +
		orderBy(aOption.SortBy, aOption.Descending) +
		limit(aOption.LimitStart, aOption.LimitLength)

	return dbDocListQuery(query)
} // orderBy()

const (
	miniQueryString = `SELECT IFNULL((SELECT group_concat(d.format, ", ")
	FROM data d WHERE d.book = b.id), "") formats,
b.path,
b.author_sort,
b.title
FROM books b
WHERE b.id = %d`
)

// QueryDocMini returns the document identified by `aID`.
//
// This function fills only the document fields `ID`, `formats`,
// `path`, `authorSort`, and `Title`.
func QueryDocMini(aID TID) *TDocument {
	query := fmt.Sprintf(miniQueryString, aID)
	rows, err := sqliteDatabase.Query(query)
	if nil != err {
		return nil
	}
	for rows.Next() {
		var formats tCSVstring
		doc := newDocument()
		doc.ID = aID
		rows.Scan(&formats, &doc.path, &doc.authorSort, &doc.Title)
		doc.formats = prepFormats(formats)

		return doc
	}

	return nil
} // QueryDocMini()

// `queryDocument()` returns the `TDocument` identified by `aID`.
func queryDocument(aID TID) *TDocument {
	list, _ := dbDocListQuery(baseQueryString + fmt.Sprintf("WHERE b.id=%d ", aID))
	if 0 < len(*list) {
		doc := (*list)[0]

		return &doc
	}

	return nil
} // queryDocument()

// `queryEntity()` returns a list of documents as defined by `aOption`.
func queryEntity(aOption *TQueryOptions) (*TDocList, error) {
	return dbDocListQuery(baseQueryString +
		fmt.Sprintf(having[aOption.Entity], aOption.ID) +
		orderBy(aOption.SortBy, aOption.Descending) +
		limit(aOption.LimitStart, aOption.LimitLength))
} // queryEntity()

// QueryLimit returns a list of `TDocument` objects.
func QueryLimit(aStart, aLength uint) (*TDocList, error) {
	query := baseQueryString +
		limit(aStart, aLength)

	return dbDocListQuery(query)
} // QueryLimit()

/* _EoF_ */
