/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/mwat56/passlist"
	"github.com/mwat56/sessions"
)

type (
	// TPageHandler provides the handling of HTTP request/response.
	TPageHandler struct {
		addr     string              // listen address ("1.2.3.4:5678")
		dd       string              // datadir: base dir for data
		docFS    http.Handler        // document file server
		lang     string              // default language
		ln       string              // the library's name
		realm    string              // host/domain to secure by BasicAuth
		staticFS http.Handler        // static file server
		theme    string              // `dark` or `light` display theme
		ul       *passlist.TPassList // user/password list
		viewList *TViewList          // list of template/views
	}
)

// NewPageHandler returns a new `TPageHandler` instance.
func NewPageHandler() (*TPageHandler, error) {
	var (
		err error
		s   string
	)
	result := new(TPageHandler)

	if s, err = AppArguments.Get("datadir"); nil != err {
		return nil, err
	}
	result.dd = s

	result.docFS = http.FileServer(http.Dir(calibreLibraryPath))

	if s, err = AppArguments.Get("lang"); nil == err {
		result.lang = s
	}

	if s, err = AppArguments.Get("libraryname"); nil == err {
		result.ln = s
	}

	s, err = AppArguments.Get("listen")
	// an empty value means: listen on all interfaces
	result.addr = s

	if s, err = AppArguments.Get("port"); nil != err {
		return nil, err
	}
	result.addr += ":" + s

	result.staticFS = http.FileServer(http.Dir(result.dd))

	if s, err = AppArguments.Get("uf"); nil != err {
		log.Printf("NewPageHandler(): %v\nAUTHENTICATION DISABLED!", err)
	} else if result.ul, err = passlist.LoadPasswords(s); nil != err {
		log.Printf("NewPageHandler(): %v\nAUTHENTICATION DISABLED!", err)
		result.ul = nil
	}

	if s, err = AppArguments.Get("realm"); nil == err {
		result.realm = s
	}

	if s, err = AppArguments.Get("theme"); nil != err {
		result.theme = "dark"
	} else {
		result.theme = s
	}

	if result.viewList, err = newViewList(filepath.Join(result.dd, "views")); nil != err {
		return nil, err
	}

	// update the thumbnails cache:
	go ThumbnailUpdate()

	return result, nil
} // NewPageHandler()

// `newViewList()` returns a list of views found in `aDirectory`
// and a possible I/O error.
func newViewList(aDirectory string) (*TViewList, error) {
	var v *TView
	result := NewViewList()

	files, err := filepath.Glob(aDirectory + "/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, fName := range files {
		fName := filepath.Base(fName[:len(fName)-7]) // remove extension
		if v, err = NewView(aDirectory, fName); nil != err {
			return nil, err
		}
		result = result.Add(v)
	}

	return result, nil
} // newViewList()

var (
	// RegEx to find path and possible added path components
	urlPartsRE = regexp.MustCompile(`(?i)^/?([\w._-]+)?/?(.*)?`)
)

// URLparts returns two parts: `rDir` holds the base-directory of `aURL`,
// `rPath` holds the remaining part of `aURL`.
//
// Depending on the actual value of `aURL` both return values may be
// empty or both may be filled; none of both will hold a leading slash.
func URLparts(aURL string) (rDir, rPath string) {
	if result, err := url.QueryUnescape(aURL); nil == err {
		aURL = result
	}
	matches := urlPartsRE.FindStringSubmatch(aURL)
	if 2 < len(matches) {
		return matches[1], matches[2]
	}

	return aURL, ""
} // URLparts()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// GetErrorPage returns an error page for `aStatus`,
// implementing the `TErrorPager` interface.
func (ph *TPageHandler) GetErrorPage(aData []byte, aStatus int) []byte {
	var empty []byte
	pageData := ph.basicTemplateData().Set("ShowForm", false)

	switch aStatus {
	case 404:
		if page, err := ph.viewList.RenderedPage("404", pageData); nil == err {
			return page
		}

	default:
		pageData = pageData.Set("Error", template.HTML(aData))
		if page, err := ph.viewList.RenderedPage("error", pageData); nil == err {
			return page
		}
	}

	return empty
} // GetErrorPage()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Address returns the configured `IP:Port` address to use for listening.
func (ph *TPageHandler) Address() string {
	return ph.addr
} // Address()

// `basicTemplateData()` returns a list of common Head entries.
func (ph *TPageHandler) basicTemplateData() *TemplateData {
	y, m, d := time.Now().Date()

	return NewTemplateData().
		Set("CSS", template.HTML(`<link rel="stylesheet" type="text/css" title="mwat's styles" href="/css/stylesheet.css"><link rel="stylesheet" type="text/css" href="/css/`+ph.theme+`.css"><link rel="stylesheet" type="text/css" href="/css/fonts.css">`)).
		Set("HasLast", false).
		Set("HasNext", false).
		Set("HasPrev", false).
		Set("Lang", ph.lang).
		Set("LibraryName", ph.ln).
		Set("Robots", "noindex,nofollow").
		Set("Title", ph.realm+fmt.Sprintf(": %d-%02d-%02d", y, m, d))
} // basicTemplateData()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// extract ID and file format from URL
	fileParseRE = regexp.MustCompile(`^(\d+)/([^/]+?)/(.*)`)
)

// `handleGET()` processes the HTTP GET requests.
func (ph *TPageHandler) handleGET(aWriter http.ResponseWriter, aRequest *http.Request) {
	qo := NewQueryOptions() // in `queryoptions.go`
	so := sessions.GetSession(aRequest)
	if qos, ok := so.GetString("QOS"); ok {
		qo.Scan(qos)
	} else if qoc := aRequest.FormValue("qoc"); 0 < len(qoc) {
		qo.UnCGI(qoc)
	}
	pageData := ph.basicTemplateData().
		Set("SLL", qo.SelectLimitOptions()).
		Set("SSB", qo.SelectSortByOptions()).
		Set("SOO", qo.SelectOrderOptions())
	path, tail := URLparts(aRequest.URL.Path)
	switch path {
	case "all", "author", "format", "lang", "publisher", "series", "tag":
		var (
			id   TID
			term string
		)
		if _, err := fmt.Sscanf(tail, "%d/%s", &id, &term); nil == err {
			qo.ID = id
		}
		qo.Entity = path
		qo.LimitStart = 0 // it's the first page of a new selection
		ph.handleQuery(qo, aWriter, so)

	case "back":
		// log.Printf("handleGET(back): %v", qo) //FIXME REMOVE
		qo.DecLimit()
		qo.Navigation = qoNext
		ph.handleQuery(qo, aWriter, so)

	case "certs": // these files are handled internally
		so.Destroy() // no session data needed
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "cover":
		so.Destroy() // no session data needed
		var (
			id    TID
			dummy string
		)
		fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		doc := QueryDocMini(id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		file, err := doc.coverAbs(true)
		if (nil != err) || (0 >= len(file)) {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.docFS.ServeHTTP(aWriter, aRequest)

	case "css":
		so.Destroy() // no session data needed
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "doc":
		var (
			id    TID
			dummy string
		)
		fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		qo.ID = id
		doc := QueryDocument(id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		pageData.Set("Document", doc)
		so.Set("QOS", qo.String())
		ph.viewList.Render("document", aWriter, pageData)

	case "favicon.ico":
		so.Destroy() // no session data needed
		http.Redirect(aWriter, aRequest, "/img/"+path, http.StatusMovedPermanently)

	case "file":
		so.Destroy() // no session data needed
		matches := fileParseRE.FindStringSubmatch(tail)
		if (nil == matches) || (3 > len(matches)) {
			http.NotFound(aWriter, aRequest)
			return
		}
		qo.ID, _ = strconv.Atoi(matches[1])
		doc := QueryDocMini(qo.ID)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		file := doc.Filename(matches[2])
		if 0 == len(file) {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.docFS.ServeHTTP(aWriter, aRequest)

	case "first":
		qo.Navigation = qoFirst
		ph.handleQuery(qo, aWriter, so)

	case "fonts":
		so.Destroy() // no session data needed
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "img":
		so.Destroy() // no session data needed
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "imprint", "impressum":
		so.Set("QOS", qo.String())
		ph.viewList.Render("imprint", aWriter, pageData)

	case "last":
		qo.Navigation = qoLast
		ph.handleQuery(qo, aWriter, so)

	case "licence", "license", "lizenz":
		so.Set("QOS", qo.String())
		ph.viewList.Render("licence", aWriter, pageData)

	case "next":
		qo.Navigation = qoNext
		ph.handleQuery(qo, aWriter, so)

	case "post":
		ph.handleQuery(qo, aWriter, so)

	case "prev":
		qo.Navigation = qoPrev
		ph.handleQuery(qo, aWriter, so)

	case "privacy", "datenschutz":
		so.Set("QOS", qo.String())
		ph.viewList.Render("privacy", aWriter, pageData)

	case "robots.txt":
		so.Destroy() // no session data needed
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "thumb":
		so.Destroy() // no session data needed
		var (
			id    TID
			dummy string
		)
		fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		doc := QueryDocMini(id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		tName, err := Thumbnail(doc)
		if nil != err {
			http.NotFound(aWriter, aRequest)
			return
		}
		file, err := filepath.Rel(ph.dd, tName)
		if nil != err {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "views": // this files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "":
		ph.handleQuery(qo, aWriter, so)

	default:
		// if nothing matched (above) reply to the request
		// with an HTTP 404 not found error.
		http.NotFound(aWriter, aRequest)

	} // switch
} // handleGET()

// `handlePOST()` process the HTTP POST requests.
func (ph *TPageHandler) handlePOST(aWriter http.ResponseWriter, aRequest *http.Request) {
	path, _ := URLparts(aRequest.URL.Path)
	switch path {
	case "qo": // the only valid POST destination
		qo := NewQueryOptions()
		so := sessions.GetSession(aRequest)
		if qos, ok := so.GetString("QOS"); ok {
			qo.Scan(qos).Update(aRequest)
		} else if qos := aRequest.FormValue("qos"); 0 < len(qos) {
			qo.Scan(qos).Update(aRequest)
		}

		// check which of the four possible SUBMIT buttons was activated
		if search := aRequest.FormValue("search"); 0 < len(search) {
			qo.DecLimit()
		} else if first := aRequest.FormValue("first"); 0 < len(first) {
			qo.Navigation = qoFirst
		} else if last := aRequest.FormValue("last"); 0 < len(last) {
			qo.Navigation = qoLast
		} else if prev := aRequest.FormValue("prev"); 0 < len(prev) {
			qo.Navigation = qoPrev
		} else if next := aRequest.FormValue("next"); 0 < len(next) {
			qo.Navigation = qoNext
		}
		ph.handleQuery(qo, aWriter, so)

	default:
		// if nothing matched (above) reply to the request
		// with an HTTP 404 "not found" error.
		http.NotFound(aWriter, aRequest)
	}
} // handlePOST()

// `handleQuery()` serves the logical web-root directory.
func (ph *TPageHandler) handleQuery(aOption *TQueryOptions, aWriter http.ResponseWriter, aSession *sessions.TSession) {
	var (
		count   int
		doclist *TDocList
		err     error
	)
	if 0 < len(aOption.Matching) {
		count, doclist, err = QuerySearch(aOption)
	} else {
		count, doclist, err = QueryBy(aOption)
	}
	if nil != err {
		//TODO better error handling
		log.Printf("handleQuery() QeueryBy/QuerySearch: %v\n", err)
	}
	if 0 < count {
		aOption.QueryCount = uint(count)
	} else {
		aOption.QueryCount = 0
	}

	switch aOption.Navigation {
	case qoFirst:
		aOption.LimitStart = 0
	case qoPrev:
		aOption.DecLimit().DecLimit()
	case qoLast:
		if aOption.QueryCount <= aOption.LimitLength {
			aOption.LimitStart = 0
		} else {
			aOption.LimitStart = aOption.QueryCount - aOption.LimitLength
		}
	default:
		// qoNext: nothing to do here
	}

	BFirst := aOption.LimitStart + 1 // zero-based to one-based
	BCount := aOption.QueryCount
	BLast := aOption.LimitStart + aOption.LimitLength
	if BLast > aOption.QueryCount {
		BLast = aOption.QueryCount
	}
	hasFirst := 0 < aOption.LimitStart
	hasLast := aOption.QueryCount > (aOption.LimitStart + aOption.LimitLength + 1)
	hasNext := aOption.QueryCount > (aOption.LimitStart + aOption.LimitLength)
	hasPrev := aOption.LimitStart > aOption.LimitLength
	aOption.IncLimit()
	pageData := ph.basicTemplateData().
		Set("BFirst", BFirst).
		Set("BLast", BLast).
		Set("BCount", BCount).
		Set("Documents", doclist).
		Set("HasFirst", hasFirst).
		Set("HasLast", hasLast).
		Set("HasNext", hasNext).
		Set("HasPrev", hasPrev).
		Set("IsGrid", qoLayoutGrid == aOption.Layout).
		Set("Matching", aOption.Matching).
		Set("SIDNAME", sessions.SIDname()).
		Set("SID", aSession.ID()).
		Set("SLO", aOption.SelectLayoutOptions()).
		Set("SLL", aOption.SelectLimitOptions()).
		Set("SOO", aOption.SelectOrderOptions()).
		Set("SSB", aOption.SelectSortByOptions()).
		Set("ShowForm", true)
	aSession.Set("QOS", aOption.String())
	if err = ph.viewList.Render("index", aWriter, pageData); nil != err {
		//TODO better error handling
		log.Printf("handleQuery() Render: %v\n", err)
	}
} // handleQuery()

// NeedAuthentication returns `true` if authentication is needed,
// or `false` otherwise.
//
// `aRequest` is the request to check.
func (ph *TPageHandler) NeedAuthentication(aRequest *http.Request) bool {
	return (nil != ph.ul)
} // NeedAuthentication()

// ServeHTTP handles the incoming HTTP requests.
func (ph TPageHandler) ServeHTTP(aWriter http.ResponseWriter, aRequest *http.Request) {
	if ph.NeedAuthentication(aRequest) {
		if !ph.ul.IsAuthenticated(aRequest) {
			passlist.Deny(ph.realm, aWriter)
			return
		}
	}

	switch aRequest.Method {
	case "GET":
		ph.handleGET(aWriter, aRequest)

	case "POST":
		ph.handlePOST(aWriter, aRequest)

	default:
		http.Error(aWriter, "HTTP Method Not Allowed", http.StatusMethodNotAllowed)
	}
} // ServeHTTP()

/* _EoF_ */
