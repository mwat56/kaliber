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

	"github.com/mwat56/apachelogger"
	"github.com/mwat56/errorhandler"
	"github.com/mwat56/kaliber"
	"github.com/mwat56/sessions"
)

// `userCmdline()` checks for and executes user handling functions.
func userCmdline() {
	var (
		err   error
		fn, s string
	)
	if fn, err = kaliber.AppArguments.Get("uf"); (nil != err) || (0 == len(fn)) {
		return // without user file nothing to do
	}
	// All the following `xxxUser()` calls terminate the program
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
		kaliber.ListUser(fn)
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
			apachelogger.Log(`kaliber/catchSignals`, msg)
			runtime.Gosched() // let the logger write
			if err := aServer.Shutdown(context.Background()); nil != err {
				log.Fatalf("%s: %v\n", os.Args[0], err)
			}
		}
	}()
} // setupSignals()

func main() {
	kaliber.InitConfig()
	var (
		err       error
		handler   http.Handler
		ph        *kaliber.TPageHandler
		ck, cp, s string
	)
	Me, _ := filepath.Abs(os.Args[0])
	if err = kaliber.OpenDatabase(); nil != err {
		kaliber.ShowHelp()
		s = fmt.Sprintf("%s: %v", Me, err)
		apachelogger.Log("Kaliber/main", s)
		runtime.Gosched() // let the logger write
		log.Fatalln(s)
	}

	// handle commandline user maintenance:
	userCmdline()

	if ph, err = kaliber.NewPageHandler(); nil != err {
		kaliber.ShowHelp()
		s = fmt.Sprintf("%s: %v", Me, err)
		apachelogger.Log("Kaliber/main", s)
		runtime.Gosched() // let the logger write
		log.Fatalln(s)
	}
	handler = errorhandler.Wrap(ph, ph)

	// inspect `sessiondir` commandline argument and setup the session handler
	if s, err = kaliber.AppArguments.Get("sessiondir"); (nil == err) && (0 < len(s)) {
		// we assume, an error means: no automatic session handling
		handler = sessions.Wrap(handler, s)
	}

	// inspect `logfile` commandline argument and setup the `ApacheLogger`
	if s, err = kaliber.AppArguments.Get("logfile"); (nil == err) && (0 < len(s)) {
		// we assume, an error means: no logfile
		handler = apachelogger.Wrap(handler, s)
	}

	// We need a `server` reference to use it in `setupSinals()` below
	// and to set some reasonable timeouts:
	server := &http.Server{
		Addr:              ph.Address(),
		Handler:           handler,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	setupSignals(server)

	ck, _ = kaliber.AppArguments.Get("certKey")
	cp, _ = kaliber.AppArguments.Get("certPem")
	if (0 < len(ck)) && (0 < len(cp)) {
		s = fmt.Sprintf("%s listening HTTPS at %s", Me, ph.Address())
		log.Println(s)
		apachelogger.Log("Kaliber/main", s)
		if err = server.ListenAndServeTLS(cp, ck); nil != err {
			s = fmt.Sprintf("%s %v", Me, err)
			apachelogger.Log("Kaliber/main", s)
			runtime.Gosched() // let the logger write
			log.Fatalln(s)
		}
		return
	}

	s = fmt.Sprintf("%s listening HTTP at %s", Me, ph.Address())
	log.Println(s)
	apachelogger.Log("Kaliber/main", s)
	if err = server.ListenAndServe(); nil != err {
		s = fmt.Sprintf("%s %v", Me, err)
		apachelogger.Log("Kaliber/main", s)
		runtime.Gosched() // let the logger write
		log.Fatalln(s)
	}
} // main()

/* _EoF_ */
