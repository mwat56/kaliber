/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
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
		NextStart   uint   // the next start position for LIMIT
		// `NextStart` is always `LimitStart` + `LimitLength`
		SortBy uint8 // display order of documents
	}
)

// CGI returns the object's query escaped string representation
// fit for use as the `qos` CGI argument.
func (qo *TQueryOptions) CGI() string {
	return `?qo="` + url.QueryEscape(qo.String()) + `"`
} // CGI()

// DescendSelectOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) DescendSelectOptions() *TStringMap {
	result := make(TStringMap, 2)
	if qo.Descending {
		result["ascending"] = `<option value="ascending">`
		result["descending"] = `<option SELECTED value="descending">`
	} else {
		result["ascending"] = `<option SELECTED value="ascending">`
		result["descending"] = `<option alue="descending">`
	}

	return &result
} // ()

// DecLimit decrements the LIMIT values.
func (qo *TQueryOptions) DecLimit() *TQueryOptions {
	if 0 < qo.LimitStart {
		if qo.LimitStart == qo.LimitLength {
			qo.LimitStart = 0
		} else {
			qo.LimitStart -= qo.LimitLength
		}
	}
	qo.NextStart += qo.LimitLength

	return qo
} // DecLimit()

// IncLimit increments the LIMIT values.
func (qo *TQueryOptions) IncLimit() *TQueryOptions {
	qo.LimitStart = qo.NextStart
	qo.NextStart += qo.LimitLength

	return qo
} // decLimit()

func (qo *TQueryOptions) limitSelectOptionsPrim(aMap *TStringMap, aLimit uint, aIndex string) {
	if aLimit == qo.LimitLength {
		(*aMap)[aIndex] = `<option SELECTED value="` + aIndex + `">`
	} else {
		(*aMap)[aIndex] = `<option value="` + aIndex + `">`
	}
} // limitSelectOptionsPrim()

// LimitSelectOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) LimitSelectOptions() *TStringMap {
	result := make(TStringMap, 4)
	qo.limitSelectOptionsPrim(&result, 10, "10")
	qo.limitSelectOptionsPrim(&result, 25, "25")
	qo.limitSelectOptionsPrim(&result, 50, "50")
	qo.limitSelectOptionsPrim(&result, 100, "100")

	return &result
} // LimitSelectOptions()

func (qo *TQueryOptions) sortSelectOptionsPrim(aMap *TStringMap, aSort uint8, aIndex string) {
	if aSort == qo.SortBy {
		(*aMap)[aIndex] = `<option SELECTED value="` + aIndex + `">`
	} else {
		(*aMap)[aIndex] = `<option value="` + aIndex + `">`
	}
} // sortSelectOptionsPrim()

// SortSelectOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SortSelectOptions() *TStringMap {
	result := make(TStringMap, 9)
	qo.sortSelectOptionsPrim(&result, SortByAuthor, "author")
	qo.sortSelectOptionsPrim(&result, SortByLanguage, "language")
	qo.sortSelectOptionsPrim(&result, SortByPublisher, "publisher")
	qo.sortSelectOptionsPrim(&result, SortByRating, "rating")
	qo.sortSelectOptionsPrim(&result, SortBySeries, "series")
	qo.sortSelectOptionsPrim(&result, SortBySize, "size")
	qo.sortSelectOptionsPrim(&result, SortByTags, "tags")
	qo.sortSelectOptionsPrim(&result, SortByTime, "time")
	qo.sortSelectOptionsPrim(&result, SortByTitle, "title")

	return &result
} // SortSelectOptions()

// String returns the options as a `|` delimited string.
func (qo *TQueryOptions) String() string {
	return fmt.Sprintf(`|%d|%t|%q|%d|%d|%q|%d|%d|`,
		qo.ID, qo.Descending, qo.Entity,
		qo.LimitLength, qo.LimitStart,
		qo.Matching, qo.NextStart, qo.SortBy)
} // String()

// Scan returns the options read from `aString`.
func (qo *TQueryOptions) Scan(aString string) *TQueryOptions {
	fmt.Sscanf(aString, `|%d|%t|%q|%d|%d|%q|%d|%d|`, &qo.ID, &qo.Descending, &qo.Entity, &qo.LimitLength, &qo.LimitStart, &qo.Matching, &qo.NextStart, &qo.SortBy)

	return qo
} // Scan()

// UnCGI unescapes the given `aCGI`.
//
// If there are errors during unescaping the current values remain unchanged.
func (qo *TQueryOptions) UnCGI(aCGI string) *TQueryOptions {
	if qosu, err := url.QueryUnescape(aCGI); nil == err {
		return qo.Scan(qosu)
	}

	return qo
} // UnCGI()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `getQueryOptions()` returns a `TQueryOptions` instance with values
// read from the `aRequest` data.
func getQueryOptions(aRequest *http.Request) *TQueryOptions {
	result := NewQueryOptions()
	if qos := aRequest.FormValue("qos"); 0 < len(qos) {
		result.Scan(qos) // form POST
	} else if qoc := aRequest.FormValue("qoc"); 0 < len(qoc) {
		result.UnCGI(qoc) // page GET
	}

	if fll := aRequest.FormValue("limitlength"); 0 < len(fll) {
		if ll, err := strconv.Atoi(fll); nil == err {
			result.LimitLength = uint(ll)
		}
	}
	if fob := aRequest.FormValue("order"); 0 < len(fob) {
		if "descending" == fob {
			result.Descending = true
		}
	}
	if fsb := aRequest.FormValue("sortby"); 0 < len(fsb) {
		switch fsb {
		case "author":
			result.SortBy = SortByAuthor
		case "language":
			result.SortBy = SortByLanguage
		case "rating":
			result.SortBy = SortByRating
		case "series":
			result.SortBy = SortBySeries
		case "size":
			result.SortBy = SortBySize
		case "tags":
			result.SortBy = SortByTags
		case "time":
			result.SortBy = SortByTime
		case "title":
			result.SortBy = SortByTitle
		}
	}

	return result
} // getQueryOptions()

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
