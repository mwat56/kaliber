/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"flag"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mwat56/ini"
)

// `parseFlagDebug()` calls `parseFlags()` and returns `AppArgs`.
//
// This function is meant for unit testing only.
func parseFlagDebug() *TAppArgs {
	flag.CommandLine = flag.NewFlagSet(`Kaliber`, flag.ExitOnError)

	// Define some flags used by `testing` to avoid
	// bailing out during the test.
	var coverprofile, run, testlogfile, timeout string
	flag.CommandLine.StringVar(&coverprofile, `test.coverprofile`, coverprofile,
		"coverprofile for tests")
	flag.CommandLine.StringVar(&run, `test.run`, run,
		"run for tests")
	flag.CommandLine.StringVar(&testlogfile, `test.testlogfile`, testlogfile,
		"testlogfile for tests")
	flag.CommandLine.StringVar(&timeout, `test.timeout`, timeout,
		"timeout for tests")

	setFlags()
	parseFlags()

	return &AppArgs
} // parseFlagDebug

// `readFlagsDebug()` calls `readFlags()` and returns `AppArgs`.
//
// This function is meant for unit testing only.
func readFlagsDebug() *TAppArgs {
	flag.CommandLine = flag.NewFlagSet(`Kaliber`, flag.ExitOnError)

	AppArgs = TAppArgs{}
	// Set up some required values:
	AppArgs.DataDir, _ = filepath.Abs(`./`)
	AppArgs.LibName = `testing`
	AppArgs.LibPath = `/var/opt/Calibre`

	readFlags()

	return &AppArgs
} // readFlagsDebug()

// `setFlagsDebug()` calls `setFlags()` and returns `AppArgs`.
//
// This function is meant for unit testing only.
func setFlagsDebug() *TAppArgs {
	flag.CommandLine = flag.NewFlagSet(`Kaliber`, flag.ExitOnError)

	var ini1 ini.TIniList
	// Clear/reset the INI values to simulate missing INI file(s):
	appArguments = tArguments{*ini1.GetSection(``)}

	setFlags()

	return &AppArgs
} // setFlagsDebug()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func Test_parseFlagDebug(t *testing.T) {
	w1 := &TAppArgs{
		AccessLog:     ``,
		AuthAll:       true,
		BooksPerPage:  24,
		CertKey:       ``,
		CertPem:       ``,
		DataDir:       `/home/matthias/devel/Go/src/github.com/mwat56/kaliber`,
		DelWhitespace: true,
		ErrorLog:      ``,
		GZip:          true,
		Lang:          `en`,
		LibName:       ``,
		LibPath:       `/var/opt/Calibre`,
		Listen:        `127.0.0.1`,
		LogStack:      false,
		PassFile:      ``,
		Port:          8383,
		Realm:         `eBooks Host`,
		SessionDir:    ``,
		SessionTTL:    1200,
		SidName:       `sid`,
		SQLTrace:      ``,
		Theme:         `dark`,
		UserAdd:       ``,
		UserCheck:     ``,
		UserDelete:    ``,
		UserList:      false,
		UserUpdate:    ``,
	}
	tests := []struct {
		name string
		want *TAppArgs
	}{
		// TODO: Add test cases.
		{" 1", w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFlagDebug(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFlagDebug() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_parseFlagDebug()

func Test_readFlagsDebug(t *testing.T) {
	w1 := &TAppArgs{
		AccessLog:     ``,
		AuthAll:       false,
		BooksPerPage:  24,
		CertKey:       ``,
		CertPem:       ``,
		DataDir:       `/home/matthias/devel/Go/src/github.com/mwat56/kaliber`,
		DelWhitespace: false,
		ErrorLog:      ``,
		GZip:          false,
		Lang:          `en`,
		LibName:       `testing`,
		LibPath:       `/var/opt/Calibre`,
		Listen:        `127.0.0.1`,
		LogStack:      false,
		PassFile:      ``,
		Port:          8383,
		Realm:         `eBooks Host`,
		SessionDir:    `/home/matthias/devel/Go/src/github.com/mwat56/kaliber/sessions`,
		SessionTTL:    1200,
		SidName:       `sid`,
		SQLTrace:      ``,
		Theme:         `dark`,
		UserAdd:       ``,
		UserCheck:     ``,
		UserDelete:    ``,
		UserList:      false,
		UserUpdate:    ``,
	}
	tests := []struct {
		name string
		want *TAppArgs
	}{
		// TODO: Add test cases.
		{" 1", w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readFlagsDebug(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFlagsDebug() = %v,\nwant %v", got, tt.want)
			}
		})
	}

	AppArgs = TAppArgs{} // clear/reset the structure
} // Test_readFlagsDebug()

func Test_readIniFiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{" 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readIniFiles()
		})
	}
} // Test_readIniFiles()

func Test_setFlagsDebug(t *testing.T) {
	w1 := &TAppArgs{
		AccessLog:     ``,
		AuthAll:       true,
		BooksPerPage:  24,
		CertKey:       ``,
		CertPem:       ``,
		DataDir:       `/home/matthias/devel/Go/src/github.com/mwat56/kaliber`,
		DelWhitespace: true,
		ErrorLog:      ``,
		GZip:          true,
		Lang:          `en`,
		LibName:       ``,
		LibPath:       `/var/opt/Calibre`,
		Listen:        `127.0.0.1`,
		LogStack:      false,
		PassFile:      ``,
		Port:          8383,
		Realm:         `eBooks Host`,
		SessionDir:    ``,
		SessionTTL:    1200,
		SidName:       `sid`,
		SQLTrace:      ``,
		Theme:         `dark`,
		UserAdd:       ``,
		UserCheck:     ``,
		UserDelete:    ``,
		UserList:      false,
		UserUpdate:    ``,
	}
	tests := []struct {
		name string
		want *TAppArgs
	}{
		// TODO: Add test cases.
		{" 1", w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setFlagsDebug(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setFlagsDebug() = %v,\nwant %v", got, tt.want)
			}
		})
	}

	AppArgs = TAppArgs{} // clear/reset the structure
} // Test_setFlagsDebug()

func TestShowHelp(t *testing.T) {
	_ = setFlagsDebug()
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{" 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ShowHelp()
		})
	}
} // TestShowHelp()
