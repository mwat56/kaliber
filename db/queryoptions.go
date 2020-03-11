/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type (
	// TSortType is used for the sorting options.
	TSortType uint8
)

// Constants defining the ORDER_BY clause
const (
	qoSortUnsorted      = TSortType(iota)
	qoSortByAcquisition // acquisition date
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

// Definition of the GUI language to use
const (
	QoLangGerman  = uint8(0)
	QoLangEnglish = uint8(1)
)

// Definition of the layout type
const (
	QoLayoutList = uint8(0)
	QoLayoutGrid = uint8(1)
)

// Definition of the CSS theme to use
const (
	QoThemeLight = uint8(0)
	QoThemeDark  = uint8(1)
)

type (
	// TQueryOptions hold properties configuring a query.
	//
	// This type is used by the HTTP pagehandler when receiving
	// a page's data.
	TQueryOptions struct {
		ID          TID       // an entity ID to lookup
		Descending  bool      // sort direction
		Entity      string    // query for a certain entity (authors, publisher, series, tags)
		GuiLang     uint8     // GUI language
		Layout      uint8     // either `qoLayoutList` or `qoLayoutGrid`
		LimitLength uint      // number of documents per page
		LimitStart  uint      // starting number
		Matching    string    // text to lookup in all documents
		QueryCount  uint      // number of DB records matching the query options
		SortBy      TSortType // display order of documents (`qoSortByXXX`)
		Theme       uint8     // CSS presentation theme
		VirtLib     string    // virtual libraries
	}
)

// Pattern used by `String()` and `Scan()`:
const (
	qoStringPattern = `|%d|%t|%q|%d|%d|%d|%d|%q|%d|%d|%d|%q|`
	//                   |  |  |  |  |  |  |  |  |  |  |  + Theme
	//                   |  |  |  |  |  |  |  |  |  |  + Theme
	//                   |  |  |  |  |  |  |  |  |  + SortBy
	//                   |  |  |  |  |  |  |  |  + QueryCount
	//                   |  |  |  |  |  |  |  + Matching
	//                   |  |  |  |  |  |  + LimitStart
	//                   |  |  |  |  |  + LimitLength
	//                   |  |  |  |  + Layout
	//                   |  |  |  + GUI lang
	//                   |  |  + Entity
	//                   |  + Descending
	//                   + ID
)

// DecLimit decrements the LIMIT-start value.
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
//
//	`aString` The value string to scan.
func (qo *TQueryOptions) Scan(aString string) *TQueryOptions {
	var m, v string
	_, _ = fmt.Sscanf(aString, qoStringPattern,
		&qo.ID, &qo.Descending, &qo.Entity, &qo.GuiLang, &qo.Layout,
		&qo.LimitLength, &qo.LimitStart, &m, &qo.QueryCount,
		&qo.SortBy, &qo.Theme, &v)
	qo.Matching = strings.TrimSpace(m)
	if "-" == v {
		qo.VirtLib = ""
	} else {
		qo.VirtLib = strings.TrimSpace(v)
	}

	return qo
} // Scan()

// SelectLanguageOptions returns a list of two SELECT/OPTIONs.
func (qo *TQueryOptions) SelectLanguageOptions() *TStringMap {
	result := make(TStringMap, 2)
	switch qo.GuiLang {
	case QoLangEnglish:
		result["de"] = `<option value="de">`
		result["en"] = `<option SELECTED value="en">`
	case QoLangGerman:
		fallthrough
	default:
		result["de"] = `<option SELECTED value="de">`
		result["en"] = `<option value="en">`
	}

	return &result
} // SelectLanguageOptions()

// SelectLayoutOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectLayoutOptions() *TStringMap {
	result := make(TStringMap, 2)
	if QoLayoutList == qo.Layout {
		result["list"] = `<option SELECTED value="list">`
		result["grid"] = `<option value="grid">`
	} else {
		result["list"] = `<option value="list">`
		result["grid"] = `<option SELECTED value="grid">`
	}

	return &result
} // SelectLayoutOptions()

var (
	optionSelectedLookup = map[bool]string{
		true:  ` SELECTED`,
		false: ``,
	}
)

// SelectLimitOptions returns a list of SELECT/OPTIONs.
func (qo *TQueryOptions) SelectLimitOptions() string {
	iList := [6]uint{9, 24, 48, 99, 249, 498}
	sList := make([]string, len(iList))
	for idx, limit := range iList {
		sList[idx] = fmt.Sprintf(`<option%s value="%d">%d</option>`, optionSelectedLookup[limit == qo.LimitLength], limit, limit)
	}

	return strings.Join(sList, `\n`)
} // SelectLimitOptions()

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
	result := make(TStringMap, 10)
	qo.selectSortByPrim(&result, qoSortByAcquisition, "acquisition")
	qo.selectSortByPrim(&result, qoSortByAuthor, "authors")
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

func (qo *TQueryOptions) selectSortByPrim(aMap *TStringMap, aSort TSortType, aIndex string) {
	if aSort == qo.SortBy {
		(*aMap)[aIndex] = `<option SELECTED value="` + aIndex + `">`
	} else {
		(*aMap)[aIndex] = `<option value="` + aIndex + `">`
	}
} // sortSelectOptionsPrim()

// String returns the options as a `|` delimited string.
func (qo *TQueryOptions) String() string {
	return fmt.Sprintf(qoStringPattern,
		qo.ID, qo.Descending, qo.Entity, qo.GuiLang, qo.Layout,
		qo.LimitLength, qo.LimitStart, qo.Matching,
		qo.QueryCount, qo.SortBy, qo.Theme, qo.VirtLib)
} // String()

// SelectThemeOptions returns a list of two SELECT/OPTIONs.
func (qo *TQueryOptions) SelectThemeOptions() *TStringMap {
	result := make(TStringMap, 2)
	switch qo.Theme {
	case QoThemeLight:
		result["light"] = `<option SELECTED value="light">`
		result["dark"] = `<option value="dark">`
	case QoThemeDark:
		result["light"] = `<option value="light">`
		result["dark"] = `<option SELECTED value="dark">`
	}

	return &result
} // SelectThemeOptions()

// SelectVirtLibOptions returns the SELECT/OPTIONs of virtual libraries.
func (qo *TQueryOptions) SelectVirtLibOptions() string {
	return VirtLibOptions(qo.VirtLib)
} // SelectVirtLibOptions()

// Update returns a `TQueryOptions` instance with updated values
// read from the `aRequest` data.
func (qo *TQueryOptions) Update(aRequest *http.Request) *TQueryOptions {
	// The form fields are defined/used in `02header.gohtml`
	if lang := aRequest.FormValue("guilang"); 0 < len(lang) {
		var l uint8 // defaults to `0` == `qoLangGerman`
		if "en" == lang {
			l = QoLangEnglish
		}
		qo.GuiLang = l
	} else {
		qo.GuiLang = QoLangGerman
	}

	if lt := aRequest.FormValue("layout"); 0 < len(lt) {
		var l uint8 // default to `0` == `qoLayoutList`
		if "grid" == lt {
			l = QoLayoutGrid
		}
		qo.Layout = l
	} else {
		qo.Layout = QoLayoutList
	}

	if fll := aRequest.FormValue("limitlength"); 0 < len(fll) {
		if ll, err := strconv.Atoi(fll); nil == err {
			limLen := uint(ll)
			if limLen != qo.LimitLength {
				qo.DecLimit()
				qo.LimitLength = limLen
			}
		}
	}

	if matching := aRequest.FormValue("matching"); 0 < len(matching) {
		if matching != qo.Matching {
			qo.ID, qo.Matching, qo.LimitStart, qo.VirtLib = 0, matching, 0, ""
		}
	} else {
		qo.Entity, qo.ID, qo.Matching = "", 0, ""
	}

	if fob := aRequest.FormValue("order"); 0 < len(fob) {
		desc := ("descending" == fob)
		if desc != qo.Descending {
			qo.Descending, qo.LimitStart = desc, 0
		}
	} else {
		qo.Descending = false
	}

	if fsb := aRequest.FormValue("sortby"); 0 < len(fsb) {
		var sb TSortType
		switch fsb {
		case "acquisition":
			sb = qoSortByAcquisition
		case "authors":
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
			qo.LimitStart, qo.SortBy = 0, sb
		}
	} else {
		qo.SortBy = qoSortByAcquisition
	}

	if theme := aRequest.FormValue("theme"); 0 < len(theme) {
		var t uint8 // defaults to `0` == `qoThemeLight`
		if "dark" == theme {
			t = QoThemeDark
		}
		qo.Theme = t
	} else {
		qo.Theme = QoThemeLight
	}

	if vl := aRequest.FormValue("virtlib"); 0 < len(vl) {
		if vl != qo.VirtLib {
			if "-" == vl {
				qo.VirtLib = ""
			} else {
				qo.VirtLib = vl
			}
			if "" != vl {
				if vlList, err := VirtualLibraryList(); nil == err {
					if vld, ok := vlList[vl]; ok {
						qo.Matching = vld
					}
				}
			}
			qo.Entity, qo.ID, qo.LimitStart = "", 0, 0
		}
	} else {
		qo.VirtLib = ""
	}

	return qo
} // Update()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// NewQueryOptions returns a new `TQueryOptions` instance.
func NewQueryOptions(aDocsPerPage int) *TQueryOptions {
	result := TQueryOptions{
		Descending:  true,
		LimitLength: 24,
		SortBy:      qoSortByAcquisition,
	}
	if 0 < aDocsPerPage {
		result.LimitLength = uint(aDocsPerPage)
	}

	return &result
} // NewQueryOptions()

/* _EoF_ */
