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
	SortBySize
	SortBySeries
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
	return "?qo=" + url.QueryEscape(qo.String())
} // CGI()

// IncLimit increments the LIMIT values.
func (qo *TQueryOptions) IncLimit() *TQueryOptions {
	qo.LimitStart = qo.NextStart
	qo.NextStart += qo.LimitLength

	return qo
} // IncLimit()

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
		LimitLength: 25,
		SortBy:      SortByTime,
	}

	return &result
} // NewQueryOptions()

/* _EoF_ */
