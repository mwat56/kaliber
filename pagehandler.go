/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/mwat56/apachelogger"
	"github.com/mwat56/passlist"
	"github.com/mwat56/sessions"
)

type (
	// TPageHandler provides the handling of HTTP request/response.
	TPageHandler struct {
		addr     string              // listen address ("1.2.3.4:5678")
		cacheFS  http.Handler        // cache file server (i.e. thumbnails)
		dd       string              // datadir: base dir for data
		docFS    http.Handler        // document file server
		lang     string              // default GUI language
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

	result.cacheFS = http.FileServer(http.Dir(CalibreCachePath()))

	if s, err = AppArguments.Get("datadir"); nil != err {
		return nil, err
	}
	result.dd = s

	result.docFS = http.FileServer(http.Dir(CalibreLibraryPath()))

	if s, err = AppArguments.Get("lang"); nil == err {
		result.lang = s
	}

	if s, err = AppArguments.Get("libraryname"); nil == err {
		result.ln = s
	}

	result.addr, _ = AppArguments.Get("listen")
	// an empty value means: listen on all interfaces

	if s, err = AppArguments.Get("port"); nil != err {
		return nil, err
	}
	result.addr += ":" + s

	result.staticFS = http.FileServer(http.Dir(result.dd))

	if s, err = AppArguments.Get("uf"); nil != err {
		s = fmt.Sprintf("%v\nAUTHENTICATION DISABLED!", err)
		apachelogger.Log("NewPageHandler()", s)
	} else if result.ul, err = passlist.LoadPasswords(s); nil != err {
		s = fmt.Sprintf("%v\nAUTHENTICATION DISABLED!", err)
		apachelogger.Log("NewPageHandler()", s)
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

	// avoid sessions fpr certain requests:
	sessions.ExcludePaths("/certs", "/css/", "/favicon", "/file/", "/fonts", "/img/", "/robots")

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
	phSplitTermRE = regexp.MustCompile(`(\d+)/(.+)$`)
)

// `splitIDterm()` splits `aTail` into an ID and a string term.
// This function is a helper of `TPageHandler.handleGET()`.
func splitIDterm(aTail string) (rID TID, rTerm string) {
	matches := phSplitTermRE.FindStringSubmatch(aTail)
	if (nil != matches) && (1 < len(matches)) {
		rID, _ = strconv.Atoi(matches[1])
		rTerm = matches[2]
	}

	return
} // splitIDterm()

var (
	// RegEx to find path and possible added path components
	phURLpartsRE = regexp.MustCompile(`(?i)^/?([\w._-]+)?/?(.*)?`)
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
	matches := phURLpartsRE.FindStringSubmatch(aURL)
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
	qo := NewQueryOptions()
	pageData := ph.basicTemplateData(qo).
		Set("ShowForm", false)

	switch aStatus {
	case 404:
		if page, err := ph.viewList.RenderedPage("404", pageData); nil == err {
			return page
		}

	default:
		pageData.Set("Error", template.HTML(aData)) // #nosec G203
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

// `basicTemplateData()` returns a list of common template values.
func (ph *TPageHandler) basicTemplateData(aOptions *TQueryOptions) *TemplateData {
	y, m, d := time.Now().Date()

	lang := ph.lang
	switch aOptions.GuiLang {
	case qoLangEnglish:
		lang = "en"
	case qoLangGerman:
		lang = "de"
	}

	theme := ph.theme
	switch aOptions.Theme {
	case qoThemeDark:
		theme = "dark"
	case qoThemeLight:
		theme = "light"
	}
	return NewTemplateData().
		Set("CSS", template.HTML(`<link rel="stylesheet" type="text/css" title="mwat's styles" href="/css/stylesheet.css"><link rel="stylesheet" type="text/css" href="/css/`+theme+`.css"><link rel="stylesheet" type="text/css" href="/css/fonts.css">`)).
		Set("GUILANG", aOptions.SelectLanguageOptions()).
		Set("HasLast", false).
		Set("HasNext", false).
		Set("HasPrev", false).
		Set("IsGrid", qoLayoutGrid == aOptions.Layout).
		Set("Lang", lang).
		Set("LibraryName", ph.ln).
		Set("Robots", "noindex,nofollow").
		Set("SLO", aOptions.SelectLayoutOptions()).
		Set("SLL", aOptions.SelectLimitOptions()).
		Set("SOO", aOptions.SelectOrderOptions()).
		Set("SSB", aOptions.SelectSortByOptions()).
		Set("THEME", aOptions.SelectThemeOptions()).
		Set("Title", ph.realm+fmt.Sprintf(": %d-%02d-%02d", y, m, d)).
		Set("VirtLib", aOptions.SelectVirtLibOptions()) // #nosec G203
} // basicTemplateData()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// extract ID and file format from URL
	phFileParseRE = regexp.MustCompile(`^(\d+)/([^/]+?)/(.*)`)
)

// `handleGET()` processes the HTTP GET requests.
func (ph *TPageHandler) handleGET(aWriter http.ResponseWriter, aRequest *http.Request) {
	qo := NewQueryOptions() // in `queryoptions.go`
	so := sessions.GetSession(aRequest)
	if qos, ok := so.GetString("QOS"); ok {
		qo.Scan(qos)
	}
	path, tail := URLparts(aRequest.URL.Path)
	switch path {
	case "all", "authors", "format", "languages", "publisher", "series", "tags":
		id, term := splitIDterm(tail)
		qo.Entity = path
		qo.ID = id
		qo.LimitStart = 0 // it's the first page of a new selection
		if 0 < id {
			qo.Matching = path + `:"=` + term + `"`
		}
		ph.handleQuery(qo, aWriter, so)

	case "back":
		qo.DecLimit()
		ph.handleQuery(qo, aWriter, so)

	case "certs": // these files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "cover":
		var (
			id    TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
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
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "doc":
		var (
			id    TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		qo.ID = id
		doc := QueryDocument(id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		pageData := ph.basicTemplateData(qo).
			Set("Document", doc)
		ph.handleReply("document", aWriter, so, qo, pageData)

	case "faq":
		ph.handleReply("faq", aWriter, so, qo, ph.basicTemplateData(qo))

	case "favicon.ico":
		http.Redirect(aWriter, aRequest, "/img/"+path, http.StatusMovedPermanently)

	case "file":
		matches := phFileParseRE.FindStringSubmatch(tail)
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
		qo.LimitStart = 0
		ph.handleQuery(qo, aWriter, so)

	case "fonts":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "help", "hilfe":
		ph.handleReply("help", aWriter, so, qo, ph.basicTemplateData(qo))

	case "img":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "imprint", "impressum":
		ph.handleReply("imprint", aWriter, so, qo, ph.basicTemplateData(qo))

	case "last":
		if qo.QueryCount <= qo.LimitLength {
			qo.LimitStart = 0
		} else {
			qo.LimitStart = qo.QueryCount - qo.LimitLength
		}
		ph.handleQuery(qo, aWriter, so)

	case "licence", "license", "lizenz":
		ph.handleReply("licence", aWriter, so, qo, ph.basicTemplateData(qo))

	case "next":
		ph.handleQuery(qo, aWriter, so)

	case "post":
		ph.handleQuery(qo, aWriter, so)

	case "prev":
		// Since the current LimitStart points to the _next_ query
		// start we have to decrement the value twice to go _before_.
		qo.DecLimit().DecLimit()
		ph.handleQuery(qo, aWriter, so)

	case "privacy", "datenschutz":
		ph.handleReply("privacy", aWriter, so, qo, ph.basicTemplateData(qo))

	case "robots.txt":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "sessions": // files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "thumb":
		var (
			id    TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
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
		file, err := filepath.Rel(CalibreCachePath(), tName)
		if nil != err {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.cacheFS.ServeHTTP(aWriter, aRequest)

	case "views": // files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "":
		qo.ID = 0         // reset fields
		qo.Entity = ""    // dito
		qo.LimitStart = 0 //
		qo.Matching = ""  //
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
			qo.Scan(qos)
		}
		qo.Update(aRequest)
		// Since the query options hold the LimitStart of the _next_
		// query we have to go back here one page:
		qo.DecLimit()
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
		msg := fmt.Sprintf("QeueryBy/QuerySearch: %v", err)
		apachelogger.Log("TPageHandler.handleQuery()", msg)
	}
	if 0 < count {
		aOption.QueryCount = uint(count)
	} else {
		aOption.QueryCount = 0
	}
	BFirst := aOption.LimitStart + 1 // zero-based to one-based
	BCount := aOption.QueryCount
	BLast := aOption.LimitStart + aOption.LimitLength
	if BLast > aOption.QueryCount {
		BLast = aOption.QueryCount
	}
	hasFirst := 0 < aOption.LimitStart
	hasLast := aOption.QueryCount > (aOption.LimitStart + aOption.LimitLength + 1)
	hasNext := aOption.QueryCount >= (aOption.LimitStart + aOption.LimitLength)
	hasPrev := aOption.LimitStart >= aOption.LimitLength
	aOption.IncLimit()
	pageData := ph.basicTemplateData(aOption).
		Set("BFirst", BFirst).
		Set("BLast", BLast).
		Set("BCount", BCount).
		Set("Documents", doclist).
		Set("HasFirst", hasFirst).
		Set("HasLast", hasLast).
		Set("HasNext", hasNext).
		Set("HasPrev", hasPrev).
		Set("Matching", aOption.Matching).
		Set("SID", aSession.ID()).
		Set("SIDNAME", sessions.SIDname()).
		Set("ShowForm", true)
	ph.handleReply("index", aWriter, aSession, aOption, pageData)
} // handleQuery()

// `handleReply()` sends the resulting page back to the remote user.
func (ph *TPageHandler) handleReply(aPage string, aWriter http.ResponseWriter, aSession *sessions.TSession, aOption *TQueryOptions, pageData *TemplateData) {
	aSession.Set("QOS", aOption.String())
	if err := ph.viewList.Render(aPage, aWriter, pageData); nil != err {
		msg := fmt.Sprintf("viewList.Render(%s): %v", aPage, err)
		apachelogger.Log("TPageHandler.handleReply()", msg)
	}
} // handleReply

// NeedAuthentication returns `true` if authentication is needed,
// or `false` otherwise.
//
//	`aRequest` is the request to check.
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
		msg := fmt.Sprintf("unsupported request method: %v", aRequest.Method)
		apachelogger.Log("TPageHandler.ServeHTTP()", msg)

		http.Error(aWriter, msg, http.StatusMethodNotAllowed)
	}
} // ServeHTTP()

/* _EoF_ */
