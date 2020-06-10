/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
               All rights reserved
           EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * This file provides functions to maintain a user/password file.
 */

import (
	"github.com/mwat56/passlist"
)

// ListUsers reads `aFilename` and lists all users stored in there.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aFilename` name of the password file to use.
func ListUsers(aFilename string) {
	passlist.ListUsers(aFilename)
} // ListUsers()

// UserAdd reads a password for `aUser` from the commandline
// and adds it to `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
//	`aUser` the username to add to the password file.
//	`aFilename` name of the password file to use.
func UserAdd(aUser, aFilename string) {
	passlist.AddUser(aUser, aFilename)
} // UserAdd()

// UserCheck reads a password for `aUser` from the commandline
// and compares it with the one stored in `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
//	`aUser` the username to check in the password file.
//	`aFilename` name of the password file to use.
func UserCheck(aUser, aFilename string) {
	passlist.CheckUser(aUser, aFilename)
} // UserCheck()

// UserDelete removes the entry for `aUser` from the password
// list `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
//	`aUser` the username to remove from the password file.
//	`aFilename` name of the password file to use.
func UserDelete(aUser, aFilename string) {
	passlist.DeleteUser(aUser, aFilename)
} // UserDelete()

// UserUpdate reads a password for `aUser` from the commandline
// and updates the entry in the password list `aFilename`.
//
// NOTE: This function does not return but terminates the program
// with error code `0` (zero) if successful, or `1` (one) otherwise.
//
// `aUser` the username to remove from the password file.
// `aFilename` name of the password file to use.
func UserUpdate(aUser, aFilename string) {
	passlist.UpdateUser(aUser, aFilename)
} // UserUpdate()

/* _EoF_ */
