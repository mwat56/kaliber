/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
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

// `fatal()` logs `aMessage` and terminates the program.
func fatal(aMessage string) {
	apachelogger.Err("Kaliber/main", aMessage)
	runtime.Gosched() // let the logger write
	apachelogger.Close()
	log.Fatalln(aMessage)
} // fatal()

// `userCmdline()` checks for and executes user/password handling functions.
func userCmdline() {
	if 0 == len(kaliber.AppArgs.PassFile) {
		return // without user file nothing to do
	}

	// All the following `kaliber.xxxUser()` calls terminate the program
	if 0 < len(kaliber.AppArgs.UserAdd) {
		kaliber.AddUser(kaliber.AppArgs.UserAdd, kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserCheck) {
		kaliber.CheckUser(kaliber.AppArgs.UserCheck, kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserDelete) {
		kaliber.DeleteUser(kaliber.AppArgs.UserDelete, kaliber.AppArgs.PassFile)
	}
	if kaliber.AppArgs.UserList {
		kaliber.ListUsers(kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserUpdate) {
		kaliber.UpdateUser(kaliber.AppArgs.UserUpdate, kaliber.AppArgs.PassFile)
	}
} // userCmdline()

// `setupSignals()` configures the capture of the interrupts `SIGINT`
// and `SIGTERM` to terminate the program gracefully.
//
//	`aServer` The server instance to shutdown if a signal arrives.
func setupSignals(aServer *http.Server) {
	// handle `CTRL-C` and `kill(15)`.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for signal := range c {
			msg := fmt.Sprintf("%s captured '%v', stopping program and exiting ...", os.Args[0], signal)
			apachelogger.Err(`Kaliber/catchSignals`, msg)
			log.Println(msg)
			runtime.Gosched() // let the logger write
			if err := aServer.Shutdown(context.Background()); nil != err {
				fatal(fmt.Sprintf("%s: %v", os.Args[0], err))
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

	// Handle commandline user/password maintenance:
	userCmdline()

	if ph, err = kaliber.NewPageHandler(); nil != err {
		kaliber.ShowHelp()
		fatal(fmt.Sprintf("%s: %v", Me, err))
	}
	// Setup the errorpage handler:
	handler := errorhandler.Wrap(ph, ph)

	// Inspect `sessiondir` config option and setup the session handler
	if 0 < len(kaliber.AppArgs.SessionDir) {
		// an empty string means: no automatic session handling
		handler = sessions.Wrap(handler, kaliber.AppArgs.SessionDir)
	}

	// Inspect `gzip` config option and setup the Gzip handler:
	if kaliber.AppArgs.GZip {
		// a FALSE means: no gzip compression
		handler = gziphandler.GzipHandler(handler)
	}

	// Inspect logging config options and setup the `ApacheLogger`:
	if 0 < len(kaliber.AppArgs.AccessLog) {
		// an empty string means: no logfile
		if 0 < len(kaliber.AppArgs.ErrorLog) {
			handler = apachelogger.Wrap(handler, kaliber.AppArgs.AccessLog, kaliber.AppArgs.ErrorLog)
		} else {
			handler = apachelogger.Wrap(handler, kaliber.AppArgs.AccessLog, ``)
		}
		// err = nil // for use by test for `apachelogger.SetErrLog()` (below)
	} else if 0 < len(kaliber.AppArgs.ErrorLog) {
		handler = apachelogger.Wrap(handler, ``, kaliber.AppArgs.ErrorLog)
	} else {
		handler = apachelogger.Wrap(handler, ``, ``)
	}

	// We need a `server` reference to use it in `setupSignals()`
	// and to set some reasonable timeouts:
	server := &http.Server{
		Addr:              ph.Address(),
		Handler:           handler,
		IdleTimeout:       1,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       20 * time.Second,
		// enough time for book download with little bandwidth:
		WriteTimeout: 20 * time.Minute,
	}
	if (nil == err) && (0 < len(s)) { // values from logfile test
		apachelogger.SetErrLog(server)
	}
	setupSignals(server)

	if (0 < len(kaliber.AppArgs.CertKey)) && (0 < len(kaliber.AppArgs.CertPem)) {
		s = fmt.Sprintf("%s listening HTTPS at %s", Me, server.Addr)
		log.Println(s)
		apachelogger.Log("Kaliber/main", s)
		if err = server.ListenAndServeTLS(kaliber.AppArgs.CertPem, kaliber.AppArgs.CertKey); nil != err {
			fatal(fmt.Sprintf("%s: %v", Me, err))
		}
		return
	}

	s = fmt.Sprintf("%s listening HTTP at %s", Me, server.Addr)
	log.Println(s)
	apachelogger.Log("Kaliber/main", s)
	if err = server.ListenAndServe(); nil != err {
		fatal(fmt.Sprintf("%s: %v", Me, err))
	}
} // main()

/* _EoF_ */
