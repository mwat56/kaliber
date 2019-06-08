/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Constants defining the ORDER_BY clause
const (
	SortUnsorted = uint8(iota)
	SortByAuthor
	SortByLanguage
	SortByPublisher
	SortByRating
	SortBySeries
	SortBySize
	SortByTags
	SortByTime
	SortByTitle
)

type (
	// TQueryOptions holds properties configuring a query.
	//
	// This type is used by the HTTP pagehandler when receiving
	// a page's data.
	TQueryOptions struct {
		ID          TID    // an entity ID to lookup
		Descending  bool   // sort direction
		Entity      string // limiting query to a certain entity (author,publisher, series, tag)
		LimitLength uint   // number of documents per page
		LimitStart  uint   // starting number
		Matching    string // text to lookup in all documents
		SortBy      uint8  // display order of documents
		QueryCount  uint   // number of DB records matching the query option
	}
)

// CGI returns the object's query escaped string representation
// fit for use as the `qos` CGI argument.
func (qo *TQueryOptions) CGI() string {
	return `?qo="` + base64.StdEncoding.EncodeToString([]byte(qo.String())) + `"`
	// return `?qo="` + url.QueryEscape(qo.String()) + `"`
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
} // decLimit()

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
	qo.selectSortByPrim(&result, SortByAuthor, "author")
	qo.selectSortByPrim(&result, SortByLanguage, "language")
	qo.selectSortByPrim(&result, SortByPublisher, "publisher")
	qo.selectSortByPrim(&result, SortByRating, "rating")
	qo.selectSortByPrim(&result, SortBySeries, "series")
	qo.selectSortByPrim(&result, SortBySize, "size")
	qo.selectSortByPrim(&result, SortByTags, "tags")
	qo.selectSortByPrim(&result, SortByTime, "time")
	qo.selectSortByPrim(&result, SortByTitle, "title")

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
	return fmt.Sprintf(`|%d|%t|%q|%d|%d|%q|%d|`,
		qo.ID, qo.Descending, qo.Entity,
		qo.LimitLength, qo.LimitStart,
		qo.Matching, qo.SortBy)
} // String()

// Scan returns the options read from `aString`.
func (qo *TQueryOptions) Scan(aString string) *TQueryOptions {
	fmt.Sscanf(aString, `|%d|%t|%q|%d|%d|%q|%d|`, &qo.ID, &qo.Descending, &qo.Entity, &qo.LimitLength, &qo.LimitStart, &qo.Matching, &qo.SortBy)

	return qo
} // Scan()

// UnCGI unescapes the given `aCGI`.
//
// If there are errors during unescaping the current values remain unchanged.
func (qo *TQueryOptions) UnCGI(aCGI string) *TQueryOptions {
	if qosu, err := base64.StdEncoding.DecodeString(aCGI); nil == err {
		return qo.Scan(string(qosu))
	}
	if qosu, err := url.QueryUnescape(aCGI); nil == err {
		return qo.Scan(qosu)
	}

	return qo
} // UnCGI()

// Update returns a `TQueryOptions` instance with values
// read from the `aRequest` data.
func (qo *TQueryOptions) Update(aRequest *http.Request) *TQueryOptions {
	if fll := aRequest.FormValue("limitlength"); 0 < len(fll) {
		if ll, err := strconv.Atoi(fll); nil == err {
			limlen := uint(ll)
			if limlen != qo.LimitLength {
				qo.DecLimit()
				qo.LimitLength = limlen
			}
		}
	}
	if fob := aRequest.FormValue("order"); 0 < len(fob) {
		desc := ("descending" == fob)
		if desc != qo.Descending {
			qo.Descending = desc
			qo.LimitStart = 0
		}
	}
	if matching := aRequest.FormValue("matching"); 0 < len(matching) {
		if matching != qo.Matching {
			qo.Matching = matching
			qo.LimitStart = 0
		}
	}
	if fsb := aRequest.FormValue("sortby"); 0 < len(fsb) {
		var sb uint8
		switch fsb {
		case "author":
			sb = SortByAuthor
		case "language":
			sb = SortByLanguage
		case "publisher":
			sb = SortByPublisher
		case "rating":
			sb = SortByRating
		case "series":
			sb = SortBySeries
		case "size":
			sb = SortBySize
		case "tags":
			sb = SortByTags
		case "time":
			sb = SortByTime
		case "title":
			sb = SortByTitle
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
		Entity:      "all",
		LimitLength: 25,
		SortBy:      SortByTime,
	}

	return &result
} // NewQueryOptions()

/* _EoF_ */
