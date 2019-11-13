/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package main

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/mwat56/apachelogger"
	"github.com/mwat56/errorhandler"
	"github.com/mwat56/kaliber"
	"github.com/mwat56/sessions"
)

// `fatal()` Log `aMessage` and terminate the program.
func fatal(aMessage string) {
	apachelogger.Log("Kaliber/main", aMessage)
	runtime.Gosched() // let the logger write
	log.Fatalln(aMessage)
} // fatal()

// `userCmdline()` checks for and executes user/password handling functions.
func userCmdline() {
	var (
		err   error
		fn, s string
	)
	if fn, err = kaliber.AppArguments.Get("uf"); (nil != err) || (0 == len(fn)) {
		return // without user file nothing to do
	}
	// All the following `kaliber.xxxUser()` calls terminate the program
	if s, err = kaliber.AppArguments.Get("ua"); (nil == err) && (0 < len(s)) {
		kaliber.AddUser(s, fn)
	}
	if s, err = kaliber.AppArguments.Get("uc"); (nil == err) && (0 < len(s)) {
		kaliber.CheckUser(s, fn)
	}
	if s, err = kaliber.AppArguments.Get("ud"); (nil == err) && (0 < len(s)) {
		kaliber.DeleteUser(s, fn)
	}
	if s, err = kaliber.AppArguments.Get("ul"); (nil == err) && (0 < len(s)) {
		kaliber.ListUsers(fn)
	}
	if s, err = kaliber.AppArguments.Get("uu"); (nil == err) && (0 < len(s)) {
		kaliber.UpdateUser(s, fn)
	}
} // userCmdline()

// `setupSignals()` configures the capture of the interrupts `SIGINT`,
// `SIGKILL`, and `SIGTERM` to terminate the program gracefully.
//
//	`aServer` The server instance to shutdown if a signal arrives.
func setupSignals(aServer *http.Server) {
	// handle `CTRL-C`, and `kill(15)`.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for signal := range c {
			msg := fmt.Sprintf("%s captured '%v', stopping program and exiting ...", os.Args[0], signal)
			log.Println(msg)
			apachelogger.Log(`Kaliber/catchSignals`, msg)
			runtime.Gosched() // let the logger write
			if err := aServer.Shutdown(context.Background()); nil != err {
				log.Fatalf("%s: %v\n", os.Args[0], err)
			}
		}
	}()
} // setupSignals()

func main() {
	var (
		err error
		ph  *kaliber.TPageHandler
		s   string
	)
	Me, _ := filepath.Abs(os.Args[0])
	kaliber.InitConfig()

	if err = kaliber.OpenDatabase(); nil != err {
		kaliber.ShowHelp()
		fatal(fmt.Sprintf("%s: %v", Me, err))
	}

	// Handle commandline user/password maintenance:
	userCmdline()

	if ph, err = kaliber.NewPageHandler(); nil != err {
		kaliber.ShowHelp()
		fatal(fmt.Sprintf("%s: %v", Me, err))
	}
	// Setup the errorpage handler:
	handler := errorhandler.Wrap(ph, ph)

	// Inspect `sessiondir` commandline argument and setup the session handler
	if s, err = kaliber.AppArguments.Get("sessiondir"); (nil == err) && (0 < len(s)) {
		// we assume, an error means: no automatic session handling
		handler = sessions.Wrap(handler, s)
	}

	// Inspect `gzip` commandline argument and setup the Gzip handler:
	if s, err = kaliber.AppArguments.Get("gzip"); (nil == err) && ("true" == s) {
		// we assume, an error means: no gzip compression
		handler = gziphandler.GzipHandler(handler)
	}

	// Inspect `logfile` commandline argument and setup the `ApacheLogger`
	if s, err = kaliber.AppArguments.Get("logfile"); (nil == err) && (0 < len(s)) {
		// we assume, an error means: no logfile
		handler = apachelogger.Wrap(handler, s)
	}

	// We need a `server` reference to use it in `setupSinals()`
	// and to set some reasonable timeouts:
	server := &http.Server{
		Addr:              ph.Address(),
		Handler:           handler,
		IdleTimeout:       2 * time.Minute,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      20 * time.Minute, // enough time for book download
	}
	if (nil == err) && (0 < len(s)) { // values from "logfile" test
		apachelogger.SetErrLog(server)
	}
	setupSignals(server)

	ck, _ := kaliber.AppArguments.Get("certKey")
	cp, _ := kaliber.AppArguments.Get("certPem")
	if (0 < len(ck)) && (0 < len(cp)) {
		s = fmt.Sprintf("%s listening HTTPS at %s", Me, server.Addr)
		log.Println(s)
		apachelogger.Log("Kaliber/main", s)
		fatal(fmt.Sprintf("%s: %v", Me, server.ListenAndServeTLS(cp, ck)))
		return
	}

	s = fmt.Sprintf("%s listening HTTP at %s", Me, server.Addr)
	log.Println(s)
	apachelogger.Log("Kaliber/main", s)
	fatal(fmt.Sprintf("%s: %v", Me, server.ListenAndServe()))
} // main()

/* _EoF_ */
