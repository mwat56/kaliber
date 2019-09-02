/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

/*

This file provides functions and methods to handle a single document.

*/

type (
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

	// A list of authors
	tAuthorList = TEntityList

	// A list of formats
	tFormatList = TEntityList

	// A list of identifiers
	tIdentifierList = TEntityList

	// A single language code
	tLanguage = TEntity

	// TStringMap is a map of strings indexed by string.
	TStringMap map[string]string

	// TPathList is a map of document formats holding the
	// respective library file.
	TPathList TStringMap

	// A single publisher
	tPublisher = TEntity

	// A single series
	tSeries = TEntity

	// A list of tags
	tTagList = TEntityList

	// A single document (e.g. book, magazin etc.)
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
			URL:  fmt.Sprintf("/authors/%d/%s", author.ID, url.PathEscape(author.Name)),
		}
		result = append(result, ent)
	}

	return &result
} // Authors()

// Comment returns the comments of the document.
func (doc *TDocument) Comment() template.HTML {
	return template.HTML(doc.comments) // #nosec G203
} // Comment()

// Cover returns the absolute URL path-filename of the document's cover image.
func (doc *TDocument) Cover() string {
	return fmt.Sprintf("/cover/%d/cover.gif", doc.ID)
} // Cover()

// `coverAbs()` returns the path-filename of the document's cover image.
func (doc *TDocument) coverAbs(aRelative bool) (string, error) {
	dir := filepath.Join(CalibreLibraryPath(), doc.path)
	if 0 <= strings.Index(dir, `[`) {
		// make sure to escape the meta-character
		dir = strings.Replace(dir, `[`, `\[`, -1)
	}
	filenames, err := filepath.Glob(dir + "/cover.*")
	if nil != err {
		//TODO better error handling
		return "", err
	}
	if 1 > len(filenames) {
		return "", errors.New(`TDocument.coverAbs(): no matching filenames found`)
	}
	if !aRelative {
		return filenames[0], nil
	}
	dir, err = filepath.Rel(CalibreLibraryPath(), filenames[0])
	if nil != err {
		//TODO better error handling
		return "", err
	}

	return dir, nil
} // coverAbs()

// CoverFile returns the complete path/filename of the document's cover file.
func (doc *TDocument) CoverFile() (string, error) {
	return doc.coverAbs(false)
} // CoverFile()

// DocLink returns a link to this document's page.
func (doc *TDocument) DocLink() string {
	return fmt.Sprintf("/doc/%d/doc.html", doc.ID)
} // DocLink()

// Filename returns the path-/filename of `aFormat`.
func (doc *TDocument) Filename(aFormat string) string {
	list := *doc.Filenames()
	if pName, ok := list[strings.ToUpper(aFormat)]; ok {
		if fName, err := filepath.Rel(CalibreLibraryPath(), pName); nil == err {
			return fName
		}
	}

	return ""
} // Filename()

// Filenames returns a list of path-/filename for this document
func (doc *TDocument) Filenames() *TPathList {
	result := make(TPathList, len(*doc.formats))
	dir := filepath.Join(CalibreLibraryPath(), doc.path)
	for _, format := range *doc.formats {
		if "ORIGINAL_EPUB" == format.Name {
			continue // we ignore these documents
		}
		ext := strings.ToLower(format.Name)
		if filenames, err := filepath.Glob(dir + "/*." + ext); nil == err {
			result[format.Name] = filenames[0]
		}
	}

	return &result
} // Filenames()

// Files returns a list of ID/Name/URL fields for doc format files.
func (doc *TDocument) Files() *TEntityList {
	if nil == doc.formats {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.formats))
	for _, file := range *doc.formats {
		if "ORIGINAL_EPUB" == file.Name {
			continue // we ignore this format
		}
		fName := url.PathEscape(strings.ReplaceAll(doc.Title, ` `, `_`)) + `.` + strings.ToLower(file.Name)
		ent := TEntity{
			ID:   file.ID,
			Name: file.Name,
			URL:  fmt.Sprintf("/file/%d/%s/%s", doc.ID, file.Name, fName),
		}
		result = append(result, ent)
	}
	if 0 < len(result) {
		return &result
	}

	return nil
} // Files()

// Formats returns a list of ID/Name/URL fields for doc formats.
func (doc *TDocument) Formats() *TEntityList {
	if nil == doc.formats {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.formats))
	for _, format := range *doc.formats {
		if "ORIGINAL_EPUB" == format.Name {
			continue // we ignore this format
		}
		ent := TEntity{
			ID:   format.ID,
			Name: format.Name,
			URL:  fmt.Sprintf("/format/%d/%s", format.ID, format.Name),
		}
		result = append(result, ent)
	}
	if 0 < len(result) {
		return &result
	}

	return nil
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
	if 0 < len(result) {
		return &result
	}

	return nil
} // Identifiers()

// Language returns an ID/Name/URL language struct.
func (doc *TDocument) Language() *TEntity {
	if nil == doc.language {
		return nil
	}
	result := TEntity{
		ID:   doc.language.ID,
		Name: doc.language.Name,
		URL:  fmt.Sprintf("/language/%d/%s", doc.language.ID, doc.language.Name),
	}

	return &result
} // Language()

// PubDate returns the formatted `pubdate` property.
func (doc *TDocument) PubDate() string {
	y, m, _ := doc.pubdate.Date()
	if 101 == y {
		return ""
	}

	return fmt.Sprintf("%d-%02d", y, m)
} // PubDate()

// Publisher returns an ID/Name/URL publisher struct.
func (doc *TDocument) Publisher() *TEntity {
	if nil == doc.publisher {
		return nil
	}
	result := TEntity{
		ID:   doc.publisher.ID,
		Name: doc.publisher.Name,
		URL:  fmt.Sprintf("/publisher/%d/%s", doc.publisher.ID, url.PathEscape(doc.publisher.Name)),
	}

	return &result
} // Publisher ()

// Series returns an ID/Name/URL series struct.
func (doc *TDocument) Series() *TEntity {
	if nil == doc.series {
		return nil
	}
	result := TEntity{
		ID:   doc.series.ID,
		Name: doc.series.Name,
		URL:  fmt.Sprintf("/series/%d/%s", doc.series.ID, url.PathEscape(doc.series.Name)),
	}

	return &result
} // Series()

// SeriesIndex returns the document's series index as formatted string.
func (doc *TDocument) SeriesIndex() string {
	result := fmt.Sprintf("%.2f", doc.seriesindex)
	parts := strings.Split(result, `.`)
	if "00" == parts[1] {
		return parts[0]
	}

	return result
} // SeriesIndex()

// Tags returns a list of ID/Name/URL tag fields.
func (doc *TDocument) Tags() *TEntityList {
	if nil == doc.tags {
		return nil
	}
	result := make(TEntityList, 0, len(*doc.tags))
	for _, tag := range *doc.tags {
		ent := TEntity{
			ID:   tag.ID,
			Name: tag.Name,
			URL:  fmt.Sprintf("/tags/%d/%s", tag.ID, url.PathEscape(tag.Name)),
		}
		result = append(result, ent)
	}
	if 0 < len(result) {
		return &result
	}

	return nil
} // Tags()

// Thumb returns the path-filename of the document's thumbnail image.
func (doc *TDocument) Thumb() string {
	return fmt.Sprintf("/thumb/%d/cover.jpg", doc.ID)
} // Thumb()

// Timestamp returns the formatted `timestamp` property.
func (doc *TDocument) Timestamp() string {
	return doc.timestamp.Format("2006-01-02 15:04:05")
} // Timestamp()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Add appends `aDoc` to the list of documents.
func (dl *TDocList) Add(aDoc *TDocument) *TDocList {
	*dl = append(*dl, *aDoc)

	return dl
} // Add()

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
