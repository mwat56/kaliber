/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"regexp"
	"strings"
)

/*
 * This file provides functions and methods for fulltext search.
 */

type (
	// TSearch provides text search capabilities.
	TSearch struct {
		raw   string // the raw (unprocessed) search expression
		where string // used to build the WHERE clause
		next  string
	}

	tExpression struct {
		entity  string // the DB field to lookup
		matcher string // how to lookup
		not     bool   // flag negating the search result
		op      string // how to concat with the next expression
		term    string // what to lookup
	}
	// tExpressionList []tExpression
)

// `buildSQL()` returns an SQL clause based on `exp` properties
// suitable for the Calibre database.
func (exp *tExpression) buildSQL() (rWhere string) {
	b := 2 // number of brackets to close
	switch exp.entity {
	case "author":
		rWhere = `(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name`

	case "comment":
		rWhere = `(b.id IN (SELECT c.book FROM comments c WHERE (c.text`

	case "format":
		rWhere = `(b.id IN (SELECT d.book FROM data d WHERE (d.format`

	case "language":
		rWhere = `(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code`

	case "publisher":
		rWhere = `(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name`

	case "series":
		rWhere = `(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name`

	case "tag", "tags":
		rWhere = `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name`

	case "title":
		rWhere = `(b.title`
		b = 1

	default: // unknown data field
		return
	}

	if "=" == exp.matcher {
		if exp.not {
			rWhere += ` != "`
		} else {
			rWhere += ` = "`
		}
		rWhere += exp.term + `")`
	} else {
		if exp.not {
			rWhere += ` NOT`
		}
		rWhere += ` LIKE "%` + exp.term + `%")`
	}
	if 1 < b {
		rWhere += `))`
	}

	return
} // buildSQL()

/*
There are several forms to recognise:

"just a search term" -> lookup ALL book entities;
`entity:"=searchterm"` -> lookup exact match of `searchterm` in `entity`;
`entity:"~searchterm"` -> lookup `searchterm` contained in `entity`.

All three expressions can be combined by AND and OR.
All three expressions can be negated by a leading `!`.
*/

// Clause returns the produced FROM/WHERE clause.
func (so *TSearch) Clause() string {
	if 0 < len(so.raw) {
		so.Parse()
	}
	if 0 < len(so.where) {
		return ` WHERE ` + so.where // #nosec G202
	}

	return ""
} // Clause()

var (
	complexExpressionRE = regexp.MustCompile(
		`(?i)^\s*((!?)(\w+):)?"([=~])([^"]*)"(\s+(AND|OR)\s*)?`)
	//           12   3        4     5      6    7

	simpleExpressionRE = regexp.MustCompile(`(?i)^"?([^"]+)"?$`)
	//                                              1------
)

func (so *TSearch) getExpression() *tExpression {
	var exp *tExpression
	matches := complexExpressionRE.FindStringSubmatch(so.raw)
	if (nil == matches) || (0 == len(matches) || (0 == len(matches[0]))) {
		// complex RegEx didn't match
		match2 := simpleExpressionRE.FindStringSubmatch(so.raw)
		if (nil == match2) || (0 == len(match2)) {
			return nil
		}
		exp = &tExpression{
			matcher: `~`,
			term:    match2[1],
		}
		so.raw = so.raw[len(match2[0]):]
	} else {
		exp = &tExpression{
			entity:  strings.ToLower(matches[3]),
			matcher: matches[4],
			term:    matches[5],
			op:      matches[7],
			not:     ("!" == matches[2]),
		}
		so.raw = so.raw[len(matches[0]):]
	}
	if 0 < len(exp.term) {
		return exp
	}

	return nil
} // getExpression()

var (
	//
	searchTermRE = regexp.MustCompile(
		`(?i)((!?)(\w+):)"([=~])([^"]*)"`)
	//       12   3       4     5
	// 	`(?i)((!?)(\w+):)"([=~])([^"]+)"(\s+(AND|OR))?`)
	// //       12   3       4     5       6   7
	// 	`(?i)\s*((!?)(\w+):)"([=~])([^"]+)"(\s+(AND|OR)\s*)?`)
	// //       0  12   3       4     5       6   7
)

func (so *TSearch) p1() *TSearch {
	w := so.raw
	for matches := searchTermRE.FindStringSubmatch(w); 5 < len(matches); matches = searchTermRE.FindStringSubmatch(w) {
		if 0 == len(matches[5]) {
			w = strings.Replace(w, matches[0], "", 1)
			continue
		}
		exp := &tExpression{
			entity:  strings.ToLower(matches[3]),
			matcher: matches[4],
			term:    matches[5],
			not:     ("!" == matches[2]),
		}
		w = strings.Replace(w, matches[0], exp.buildSQL(), 1)

		//FIXME TODO handling (leading|trailing) garbage
	}
	so.next, so.raw, so.where = "", "", w

	return so
} // p1()

// Parse returns the parsed search term(s).
func (so *TSearch) Parse() *TSearch {
	if 0 == len(so.raw) {
		so.next, so.where = "", ""
		return so
	}
	if searchTermRE.MatchString(so.raw) {
		return so.p1()
	}

	exp := &tExpression{
		matcher: "~",
		term:    so.raw,
		op:      "OR",
	}
	exp.entity = "author"
	so.where = exp.buildSQL() + ` OR `
	exp.entity = "comment"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "format"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "language"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "publisher"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "series"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "tags"
	so.where += exp.buildSQL() + ` OR `
	exp.entity = "title"
	so.where += exp.buildSQL()
	so.raw = ""

	return so
} // Parse()

/*
func (so *TSearch) parsePrim(aExpression *tExpression) {
	switch aExpression.entity {
	case "author":
		so.where += so.next + ` b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name`

	case "comment":
		so.where += so.next + ` b.id IN (SELECT c.book FROM comments c WHERE (c.text`

	case "format":
		so.where += so.next + ` b.id IN (SELECT d.book FROM data d WHERE (d.format`

	case "language":
		so.where += so.next + ` b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code`

	case "publisher":
		so.where += so.next + ` b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name`

	case "series":
		so.where += so.next + ` b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name`

	case "tag", "tags":
		so.where += so.next + ` b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name`

	case "title":
		so.where += so.next + ` (b.title`
		if "=" == aExpression.matcher {
			if aExpression.not {
				so.where += ` != `
			} else {
				so.where += ` = `
			}
			so.where += aExpression.term + `") `
		} else {
			if aExpression.not {
				so.where += ` NOT`
			}
			so.where += ` LIKE "%` + aExpression.term + `%") `
		}
		so.next = aExpression.op
		return

	default:
		return
	}
	if "=" == aExpression.matcher {
		so.where += ` = "` + aExpression.term + `")) `
	} else {
		so.where += ` LIKE "%` + aExpression.term + `%")) `
	}
	so.next = aExpression.op
} // parsePrim()
*/

// String returns a stringfied representation.
func (so *TSearch) String() string {
	return `raw: '` + so.raw +
		`' | where: '` + so.where +
		`' | next: '` + so.next + `'`
} // String()

// Where returns the SQL code to use in the WHERE clause.
func (so *TSearch) Where() string {
	return so.where
} // Where()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// NewSearch returns a new `TSearch` instance.
func NewSearch(aSearchTerm string) *TSearch {
	return &TSearch{raw: aSearchTerm}
} // NewSearch()

/* _EoF_ */
