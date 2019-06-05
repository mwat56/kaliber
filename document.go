/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*

This file provides functions and method to handle a single document.

*/

type (
	tCSVstring = string // comma separated list

	// TID is the database index type.
	TID = int

	// TEntity is a basic entity structure.
	TEntity struct {
		ID   TID
		Name string
		URL  string
	}

	// TEntityList is a list of entities
	TEntityList []TEntity

	// a single author
	// tAuthor TEntity
	/*  struct {
		id   TID
		name string
		sort string
		link string
	}
	*/

	// a list of authors
	tAuthorList = TEntityList

	// a single book format
	// tFormat TEntity

	// a list of formats
	tFormatList = TEntityList

	// a single identifier
	// tIdentifier TEntity

	// a list of identifiers
	tIdentifierList = TEntityList

	// a single language code
	tLanguage = TEntity
	/* struct {
		id       TID
		langCode string
	}
	*/

	// a list of languages
	// tLanguageList []tLanguage

	// TPathList is a map of document formats holding the
	// respective library file.
	TPathList map[string]string

	// a single publisher
	tPublisher = TEntity

	// a list of publisher
	// tPublisherList []tPublisher

	// a single series
	tSeries = TEntity

	// a list of series'
	// tSeriesList []tSeries

	// a single tag
	// tTag TEntity

	// a list of tags
	tTagList = TEntityList

	// a single document (e.g. book, magazin etc.)
	tDocument struct {
		ID          TID
		authors     *tAuthorList
		authorSort  string
		comments    string
		flags       int
		formats     *tFormatList
		hasCover    bool
		identifiers *tIdentifierList
		ISBN        string
		language    *tLanguage
		lccn        string
		Pages       int
		path        string
		pubdate     time.Time // SQL: timestamp
		publisher   *tPublisher
		Rating      int
		Size        int64
		series      *tSeries
		seriesindex float32 // SQL: real
		tags        *tTagList
		timestamp   time.Time // SQL: timestamp
		Title       string
		TitleSort   string
		uuid        string
	}

	// TDocument represents a single document (e.g. book)
	TDocument tDocument

	// TDocList is a list of `TDocument` instances.
	TDocList []TDocument
)

// Authors returns a list of ID/Name author fields.
func (doc *TDocument) Authors() *TEntityList {
	if nil == doc.authors {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.authors))
	for _, author := range *doc.authors {
		ent := TEntity{
			ID:   author.ID,
			Name: author.Name,
			URL:  fmt.Sprintf("/author/%d/%s", doc.ID, url.PathEscape(author.Name)),
		}
		result = append(result, ent)
	}

	return &result
} // Authors()

// Comment returns the comments of the document.
func (doc *TDocument) Comment() template.HTML {
	return template.HTML(doc.comments)
} // Comment()

// Cover returns the relative path-filename of the document's cover image.
func (doc *TDocument) Cover() string {
	return fmt.Sprintf("/cover/%d/cover.jpg", doc.ID)
} // Cover()

// `cover()` returns the absolute path-filename of the document's cover image.
func (doc *TDocument) coverAbs(aRelative bool) (string, error) {
	dir := filepath.Join(calibreLibraryPath, doc.path)
	filenames, err := filepath.Glob(dir + "/cover.*")
	if (nil != err) || (1 > len(filenames)) {
		return "", err
	}
	if !aRelative {
		return filenames[0], nil
	}
	dir, err = filepath.Rel(calibreLibraryPath, filenames[0])
	if nil != err {
		return "", err
	}

	return dir, nil
} // coverAbs()

// Filename returns the path-/filename od `aFormat`.
func (doc *TDocument) Filename(aFormat string, aRelative bool) string {
	format := strings.ToUpper(aFormat)
	list := doc.Filenames()
	result, ok := (*list)[format]
	if !ok {
		return ""
	}
	if !aRelative {
		return result
	}
	if dir, err := filepath.Rel(calibreLibraryPath, result); nil == err {
		return dir
	}

	return ""
} // Filename()

// Filenames returns a list of path-/filename for this document
func (doc *TDocument) Filenames() *TPathList {
	result := make(TPathList, len(*doc.formats))
	dir := filepath.Join(calibreLibraryPath, doc.path)
	for _, format := range *doc.formats {
		if "ORIGINAL_EPUB" == format.Name {
			continue // we don't need these documents
		}
		ext := strings.ToLower(format.Name)
		if filenames, err := filepath.Glob(dir + "/*." + ext); nil == err {
			result[format.Name] = filenames[0]
		}
	}

	return &result
} // Filenames()

// Formats returns a list of ID/Name author fields.
func (doc *TDocument) Formats() *TEntityList {
	if nil == doc.formats {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.formats))
	for _, format := range *doc.formats {
		ent := TEntity{
			ID:   format.ID,
			Name: format.Name,
			URL:  fmt.Sprintf("/format/%d/%s", doc.ID, format.Name),
		}
		result = append(result, ent)
	}

	return &result
} // Formats()

// Identifiers returns a list of ID/Name/URL identifier fields.
func (doc *TDocument) Identifiers() *TEntityList {
	if nil == doc.identifiers {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.identifiers))
	for _, ident := range *doc.identifiers {
		ent := TEntity{
			ID:   ident.ID,
			Name: ident.Name,
		}
		switch ident.Name {
		case "amazon", "mobi-asin":
			ent.URL = fmt.Sprintf("https://www.amazon.com/dp/%s", ident.URL)
		case "amazon_de":
			ent.URL = fmt.Sprintf("https://www.amazon.de/dp/%s", ident.URL)
		case "barnesnoble":
			ent.URL = fmt.Sprintf("https://www.barnesandnoble.com/%s", ident.URL)
		case "edelweiss":
			ent.URL = fmt.Sprintf("https://www.edelweiss.plus/#sku=%s&page=1", ident.URL)
		case "google":
			ent.URL = fmt.Sprintf("https://books.google.com/books?id=%s", ident.URL)
		case "isbn":
			ent.URL = fmt.Sprintf("https://www.worldcat.org/isbn/%s", ident.URL)
		case "issn":
			ent.URL = fmt.Sprintf("https://www.worldcat.org/issn/%s", ident.URL)
		case "uri", "url":
			ent.URL = strings.ReplaceAll(ident.URL, "|", ":")

		default:
			continue
		}
		result = append(result, ent)
	}

	return &result
} // Identifiers()

// Language returns an ID/Name series struct.
func (doc *TDocument) Language() *TEntity {
	if nil == doc.language {
		return nil
	}
	result := TEntity{
		ID:   doc.language.ID,
		Name: doc.language.Name,
		URL:  fmt.Sprintf("/lang/%d/%s", doc.language.ID, doc.language.Name),
	}

	return &result
} // Language()

// PubDate returns the formatted `pubdate` property.
func (doc *TDocument) PubDate() string {
	return doc.pubdate.Format("2006-01")
} // PubDate()

// Publisher returns an ID/Name publisher struct.
func (doc *TDocument) Publisher() *TEntity {
	if nil == doc.publisher {
		return nil
	}
	result := TEntity{
		ID:   doc.publisher.ID,
		Name: doc.publisher.Name,
		URL:  fmt.Sprintf("/publisher/%d/%s", doc.publisher.ID, doc.publisher.Name),
	}

	return &result
} // Publisher ()

// Series returns an ID/Name series struct.
func (doc *TDocument) Series() *TEntity {
	if nil == doc.series {
		return nil
	}
	result := TEntity{
		ID:   doc.series.ID,
		Name: doc.series.Name,
		URL:  fmt.Sprintf("/series/%d/%s", doc.series.ID, doc.series.Name),
	}

	return &result
} // Series()

// SeriesIndex returns the document's series index as formatted string.
func (doc *TDocument) SeriesIndex() string {
	return fmt.Sprintf("%f", doc.seriesindex)
} // SeriesIndex()

var (
	// RegEx to find a document's number of pages
	pagesRE = regexp.MustCompile(`(?si)<meta name="calibre:user_metadata:#pages" .*?, &quot;#value#&quot;: (\d+),`)
)

func (doc *TDocument) setPages() int {
	fName := filepath.Join(calibreLibraryPath, doc.path, "metadata.opf")
	if fi, err := os.Stat(fName); (nil != err) || (0 >= fi.Size()) {
		return 0
	}
	metadata, err := ioutil.ReadFile(fName)
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
	doc.Pages = num

	return num
} // setPages()

// Tags returns a list of ID/Name tag fields.
func (doc *TDocument) Tags() *TEntityList {
	if nil == doc.tags {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.tags))
	for _, tag := range *doc.tags {
		ent := TEntity{
			ID:   tag.ID,
			Name: tag.Name,
			URL:  fmt.Sprintf("/tag/%d/%s", tag.ID, tag.Name),
		}
		result = append(result, ent)
	}

	return &result
} // Tags()

// Timestamp returns the formatted `timestamp` property.
func (doc *TDocument) Timestamp() string {
	return doc.timestamp.Format("2006-01-02 15:04:05")
} // Timestamp()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// newDocument returns a new `TDocument` instance.
func newDocument() *TDocument {
	result := &TDocument{
		authors:     &tAuthorList{},
		formats:     &tFormatList{},
		identifiers: &tIdentifierList{},
		tags:        &tTagList{},
	}

	return result
} // newDocument()

// newDocList returns a new `TDocList` instance.
func newDocList() *TDocList {
	result := make(TDocList, 0, 32)

	return &result
} // newDocList()

/* _EoF_ */
