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
 * This file provides helper functions and methods for database searches.
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
)

// `allSQL()` returns an SQL clause to match the current term
// in all suitable tables.
func (exp *tExpression) allSQL() (rWhere string) {
	exp.matcher, exp.op = "~", "OR"

	exp.entity = "author"
	rWhere = exp.buildSQL()
	exp.entity = "comment"
	rWhere += exp.buildSQL()
	exp.entity = "format"
	rWhere += exp.buildSQL()
	exp.entity = "language"
	rWhere += exp.buildSQL()
	exp.entity = "publisher"
	rWhere += exp.buildSQL()
	exp.entity = "series"
	rWhere += exp.buildSQL()
	exp.entity = "tags"
	rWhere += exp.buildSQL()
	exp.entity, exp.op = "title", ""
	rWhere += exp.buildSQL()

	return
} // allSQL()

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
	if 0 < len(exp.op) {
		rWhere += exp.op
	}

	return
} // buildSQL()

// ----------------------------------------------------------------------------

// Clause returns the produced WHERE clause.
func (so *TSearch) Clause() string {
	if 0 < len(so.raw) {
		so.Parse()
	}
	if 0 < len(so.where) {
		return ` WHERE ` + so.where // #nosec G202
	}

	return ""
} // Clause()

/*
There are several forms to recognise:

"just a search term" -> lookup ALL book entities;
`entity:"=searchterm"` -> lookup exact match of `searchterm` in `entity`;
`entity:"~searchterm"` -> lookup `searchterm` contained in `entity`.

All three expressions can be combined by AND and OR.
All three expressions can be negated by a leading `!`.
*/

var (
	// RegEx to find a search expression
	searchExpressionRE = regexp.MustCompile(
		`(?i)((!?)(\w+):)"([=~])([^"]*)"(\s*(AND|OR))?`)
	//       12   3       4     5       6   7

	searchRemainderRE = regexp.MustCompile(`\s*([\w ]+)`)
)

func (so *TSearch) p1() *TSearch {
	op, p, s, w := "", 0, "", so.raw
	for matches := searchExpressionRE.FindStringSubmatch(w); 7 < len(matches); matches = searchExpressionRE.FindStringSubmatch(w) {
		exp := &tExpression{
			entity:  strings.ToLower(matches[3]),
			not:     ("!" == matches[2]),
			matcher: matches[4],
			op:      strings.ToUpper(matches[7]),
			term:    matches[5],
		}
		s = exp.buildSQL()
		w = strings.Replace(w, matches[0], s, 1)
		p = strings.Index(w, s) + len(s)
		op = exp.op // save the latest operant for below
	}
	if p < len(w) { // check whether there's something behind the last expression
		matches := searchRemainderRE.FindStringSubmatch(w[p:])
		if 0 < len(matches) {
			exp := &tExpression{term: matches[1]}
			s = exp.allSQL()
			if 0 == len(op) {
				s = `OR ` + s
			}
			w = strings.Replace(w, matches[0], s, 1)
		}
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
	if searchExpressionRE.MatchString(so.raw) {
		return so.p1()
	}

	exp := &tExpression{term: so.raw}
	so.where, so.raw = exp.allSQL(), ""

	return so
} // Parse()

// String returns a string field representation.
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
