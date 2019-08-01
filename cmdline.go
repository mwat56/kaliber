/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
               All rights reserved
           EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * This file provides functions to maintain a user/password list.
 */

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/mwat56/passlist"
	"golang.org/x/crypto/ssh/terminal"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// AddUser reads a password for `aUser` from the commandline
// and adds it to `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aUser` the username to add to the password file.
//
// `aFilename` name of the password file to use.
func AddUser(aUser, aFilename string) {
	ul := passlist.NewList(aFilename)
	if nil == ul {
		fmt.Fprintf(os.Stderr, "can't open/create password list '%s'\n", aFilename)
		os.Exit(1)
	}
	_ = ul.Load() // ignore error since the file might not exist yet
	if ok := ul.Exists(aUser); ok {
		fmt.Fprintf(os.Stderr, "\n\t'%s' already exists in list\n", aUser)
		os.Exit(1)
	}
	pw := readPassword(true)
	if err := ul.Add(aUser, pw); nil != err {
		fmt.Fprintf(os.Stderr, "\n\tcan't add '%s' to list: %v\n", aUser, err)
		os.Exit(1)
	}
	if _, err := ul.Store(); nil != err {
		fmt.Fprintf(os.Stderr, "\n\tcan't store modified list: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\tadded '%s' to list\n\n", aUser)
	os.Exit(0)
} // AddUser()

// CheckUser reads a password for `aUser` from the commandline
// and compares it with the one stored in `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aUser` the username to check in the password file.
//
// `aFilename` name of the password file to use.
func CheckUser(aUser, aFilename string) {
	ul := userlist(aFilename)
	pw := readPassword(false)
	if ok := ul.Matches(aUser, pw); ok {
		pw = "successful"
	} else {
		pw = "failed"
	}
	fmt.Printf("\n\t'%s' password check %s\n\n", aUser, pw)
	os.Exit(0)
} // CheckUser()

// DeleteUser removes the entry for `aUser` from the password
// list `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aUser` the username to remove from the password file.
//
// `aFilename` name of the password file to use.
func DeleteUser(aUser, aFilename string) {
	ul := userlist(aFilename)
	if ok := ul.Exists(aUser); !ok {
		fmt.Fprintf(os.Stderr, "\n\tcan't find '%s' in list\n", aUser)
		os.Exit(1)
	}
	if _, err := ul.Remove(aUser).Store(); nil != err {
		fmt.Fprintf(os.Stderr, "\n\tcan't store modified list: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n\tremoved '%s' from list\n\n", aUser)
	os.Exit(0)
} // DeleteUser()

// userlist returns a new `TPassList` instance.
func userlist(aFilename string) (rList *passlist.TPassList) {
	rList, err := passlist.LoadPasswords(aFilename)
	if nil != err {
		fmt.Fprintf(os.Stderr, "can't open/create password list '%s'\n", aFilename)
		os.Exit(1)
	}

	return
} // userlist()

// ListUser reads `aFilename` and lists all users stored in there.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aFilename` name of the password file to use.
func ListUser(aFilename string) {
	ul := userlist(aFilename)
	list := ul.List()
	if 0 == len(list) {
		fmt.Fprintf(os.Stderr, "no users found in password list '%s'\n", aFilename)
		os.Exit(1)
	}
	s := strings.Join(list, "\n") + "\n"
	fmt.Println(s)
	os.Exit(0)
} // ListUser()

// `readPassword()` asks the user to input a password on the commandline.
func readPassword(aRepeat bool) (rPass string) {
	var (
		pw1, pw2 string
	)
	for {
		fmt.Print("\n password: ")
		if bPW, err := terminal.ReadPassword(syscall.Stdin); err == nil {
			if 0 < len(bPW) {
				pw1 = string(bPW)
			} else {
				fmt.Println("\n\tempty password not accepted")
				continue
			}
		}
		if aRepeat {
			fmt.Print("\nrepeat pw: ")
			if bPW, err := terminal.ReadPassword(syscall.Stdin); err == nil {
				if 0 < len(bPW) {
					pw2 = string(bPW)
				} else {
					fmt.Println("\n\tempty password not accepted")
					continue
				}
			}
		} else {
			break
		}
		if pw1 == pw2 {
			break
		}
		fmt.Fprintln(os.Stderr, "\n\tthe two passwords don't match")
	}
	fmt.Print("\n")

	return pw1
} // readPassword()

// UpdateUser reads a password for `aUser` from the commandline
// and updates the entry in the password list `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aUser` the username to remove from the password file.
//
// `aFilename` name of the password file to use.
func UpdateUser(aUser, aFilename string) {
	ul := userlist(aFilename)
	if ok := ul.Exists(aUser); !ok {
		fmt.Fprintf(os.Stderr, "\n\tcan't find '%s' in list\n", aUser)
		os.Exit(1)
	}
	pw := readPassword(true)
	if err := ul.Add(aUser, pw); nil != err {
		fmt.Fprintf(os.Stderr, "\n\tcan't update '%s': %v\n", aUser, err)
		os.Exit(1)
	}
	if _, err := ul.Store(); nil != err {
		fmt.Fprintf(os.Stderr, "\n\tcan't store modified list: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\tupdated user '%s' in list\n\n", aUser)
	os.Exit(0)
} // UpdateUser()

/* _EoF_ */
