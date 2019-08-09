/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Constants defining the ORDER_BY clause
const (
	qoSortUnsorted = uint8(iota)
	qoSortByAuthor
	qoSortByLanguage
	qoSortByPublisher
	qoSortByRating
	qoSortBySeries
	qoSortBySize
	qoSortByTags
	qoSortByTime
	qoSortByTitle
)

// Definition of the layout type
const (
	qoLayoutList = uint8(0)
	qoLayoutGrid = uint8(1)
)

type (
	// TQueryOptions hold properties configuring a query.
	//
	// This type is used by the HTTP pagehandler when receiving
	// a page's data.
	TQueryOptions struct {
		ID          TID    // an entity ID to lookup
		Descending  bool   // sort direction
		Entity      string // limiting query to a certain entity (author, publisher, series, tag)
		Layout      uint8  // either `qoLayoutList` or `qoLayoutGrid`
		LimitLength uint   // number of documents per page
		LimitStart  uint   // starting number
		Matching    string // text to lookup in all documents
		SortBy      uint8  // display order of documents (`qoSortByXXX`)
		QueryCount  uint   // number of DB records matching the query option
	}
)

// Pattern used by `String()` and `Scan()`:
const (
	qoStringPattern = `|%d|%t|%q|%d|%d|%d|%q|%d|%d|`
	//                   |  |  |  |  |  |  |  |  + SortBy
	//                   |  |  |  |  |  |  |  + QueryCount
	//                   |  |  |  |  |  |  + Matching
	//                   |  |  |  |  |  + LimitStart
	//                   |  |  |  |  + LimitLength
	//                   |  |  |  + Layout
	//                   |  |  + Entity
	//                   |  + Descending
	//                   + ID
)

// CGI returns the object's query escaped string representation
// fit for use as the `qoc` CGI argument.
func (qo *TQueryOptions) CGI() string {
	return `?qoc=` + base64.StdEncoding.EncodeToString([]byte(qo.String()))
} // CGI()

// DecLimit decrements the LIMIT values.
func (qo *TQueryOptions) DecLimit() *TQueryOptions {
	if 0 < qo.LimitStart {
		if qo.LimitStart <= qo.LimitLength {
			qo.LimitStart = 0
		} else {
			qo.LimitStart -= qo.LimitLength
		}
	}

	return qo
} // DecLimit()

// IncLimit increments the LIMIT values.
func (qo *TQueryOptions) IncLimit() *TQueryOptions {
	qo.LimitStart += qo.LimitLength

	return qo
} // IncLimit()

// Scan returns the options read from `aString`.
func (qo *TQueryOptions) Scan(aString string) *TQueryOptions {
	_, _ = fmt.Sscanf(aString, qoStringPattern,
		&qo.ID, &qo.Descending, &qo.Entity, &qo.Layout,
		&qo.LimitLength, &qo.LimitStart, &qo.Matching,
		&qo.QueryCount, &qo.SortBy)

	return qo
} // Scan()

// SelectLayoutOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectLayoutOptions() *TStringMap {
	result := make(TStringMap, 2)
	if qoLayoutList == qo.Layout {
		result["list"] = `<option SELECTED value="list">`
		result["grid"] = `<option value="grid">`
	} else {
		result["list"] = `<option value="list">`
		result["grid"] = `<option SELECTED value="grid">`
	}

	return &result
} // SelectLayoutOptions()

// SelectLimitOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectLimitOptions() *TStringMap {
	result := make(TStringMap, 4)
	qo.selectLimitPrim(&result, 10, "10")
	qo.selectLimitPrim(&result, 25, "25")
	qo.selectLimitPrim(&result, 50, "50")
	qo.selectLimitPrim(&result, 100, "100")

	return &result
} // SelectLimitOptions()

func (qo *TQueryOptions) selectLimitPrim(aMap *TStringMap, aLimit uint, aIndex string) {
	if aLimit == qo.LimitLength {
		(*aMap)[aIndex] = `<option SELECTED value="` + aIndex + `">`
	} else {
		(*aMap)[aIndex] = `<option value="` + aIndex + `">`
	}
} // selectLimitPrim()

// SelectOrderOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectOrderOptions() *TStringMap {
	result := make(TStringMap, 2)
	if qo.Descending {
		result["ascending"] = `<option value="ascending">`
		result["descending"] = `<option SELECTED value="descending">`
	} else {
		result["ascending"] = `<option SELECTED value="ascending">`
		result["descending"] = `<option value="descending">`
	}

	return &result
} // SelectOrderOptions()

// SelectSortByOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectSortByOptions() *TStringMap {
	result := make(TStringMap, 9)
	qo.selectSortByPrim(&result, qoSortByAuthor, "author")
	qo.selectSortByPrim(&result, qoSortByLanguage, "language")
	qo.selectSortByPrim(&result, qoSortByPublisher, "publisher")
	qo.selectSortByPrim(&result, qoSortByRating, "rating")
	qo.selectSortByPrim(&result, qoSortBySeries, "series")
	qo.selectSortByPrim(&result, qoSortBySize, "size")
	qo.selectSortByPrim(&result, qoSortByTags, "tags")
	qo.selectSortByPrim(&result, qoSortByTime, "time")
	qo.selectSortByPrim(&result, qoSortByTitle, "title")

	return &result
} // SelectSortByOptions()

func (qo *TQueryOptions) selectSortByPrim(aMap *TStringMap, aSort uint8, aIndex string) {
	if aSort == qo.SortBy {
		(*aMap)[aIndex] = `<option SELECTED value="` + aIndex + `">`
	} else {
		(*aMap)[aIndex] = `<option value="` + aIndex + `">`
	}
} // sortSelectOptionsPrim()

// String returns the options as a `|` delimited string.
func (qo *TQueryOptions) String() string {
	return fmt.Sprintf(qoStringPattern,
		qo.ID, qo.Descending, qo.Entity, qo.Layout,
		qo.LimitLength, qo.LimitStart, qo.Matching,
		qo.QueryCount, qo.SortBy)
} // String()

// UnCGI unescapes the given `aCGI`.
//
// If there are errors during unescaping the current values remain unchanged.
func (qo *TQueryOptions) UnCGI(aCGI string) *TQueryOptions {
	qoc, err := base64.StdEncoding.DecodeString(aCGI)
	if nil != err {
		//TODO better error handling
		log.Printf("TQueryOptions.UnCGI('%s'): %v", aCGI, err) //FIXME REMOVE
		return qo
	}

	return qo.Scan(string(qoc))
} // UnCGI()

// Update returns a `TQueryOptions` instance with values
// read from the `aRequest` data.
func (qo *TQueryOptions) Update(aRequest *http.Request) *TQueryOptions {
	if fol := aRequest.FormValue("layout"); 0 < len(fol) {
		var l uint8
		switch fol {
		case "grid":
			l = qoLayoutGrid
		default:
			l = qoLayoutList
		}
		if l != qo.Layout {
			qo.Layout = l
			qo.LimitStart = 0
		}
	}
	if fll := aRequest.FormValue("limitlength"); 0 < len(fll) {
		if ll, err := strconv.Atoi(fll); nil == err {
			limlen := uint(ll)
			if limlen != qo.LimitLength {
				qo.DecLimit()
				qo.LimitLength = limlen
			}
		}
	}
	matching := aRequest.FormValue("matching")
	if matching != qo.Matching {
		qo.Matching = matching
		qo.LimitStart = 0
	}
	if fob := aRequest.FormValue("order"); 0 < len(fob) {
		desc := ("descending" == fob)
		if desc != qo.Descending {
			qo.Descending = desc
			qo.LimitStart = 0
		}
	}
	if fsb := aRequest.FormValue("sortby"); 0 < len(fsb) {
		var sb uint8
		switch fsb {
		case "author":
			sb = qoSortByAuthor
		case "language":
			sb = qoSortByLanguage
		case "publisher":
			sb = qoSortByPublisher
		case "rating":
			sb = qoSortByRating
		case "series":
			sb = qoSortBySeries
		case "size":
			sb = qoSortBySize
		case "tags":
			sb = qoSortByTags
		case "time":
			sb = qoSortByTime
		case "title":
			sb = qoSortByTitle
		case "":
			sb = qoSortUnsorted // just to actually use this const
		}
		if sb != qo.SortBy {
			qo.SortBy = sb
			qo.LimitStart = 0
		}
	}

	return qo
} // Update()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// NewQueryOptions returns a new `TQueryOptions` instance.
func NewQueryOptions() *TQueryOptions {
	result := TQueryOptions{
		Descending:  true,
		LimitLength: 25,
		SortBy:      qoSortByTime,
	}
	if s, _ := AppArguments.Get("booksperpage"); 0 < len(s) {
		if _, err := fmt.Sscanf(s, "%d", &result.LimitLength); nil != err {
			result.LimitLength = 25
		}
	}

	return &result
} // NewQueryOptions()

/* _EoF_ */
