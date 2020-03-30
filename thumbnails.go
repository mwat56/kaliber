/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/mwat56/apachelogger"
	"github.com/mwat56/kaliber/db"
	"github.com/nfnt/resize"
)

/*
 * This file provides functions for thumbnail generation and maintenance.
 */

// `goThumbCleanup()` removes orphaned thumbnails.
//
//	`aDB` The DB handle to access the `Calibre` database.
func goThumbCleanup(aDB *db.TDataBase) {
	bd := db.CalibreCachePath()
	dirNames, err := filepath.Glob(bd + "/*")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", bd, err)
		apachelogger.Err("goThumbCleanup()", msg)
		return
	}
	for _, numDir := range dirNames {
		checkThumbBase(numDir, aDB)
	}

	// just mark that function as `used`:
	_ = thumbnailRemove(db.NewDocument())
} // goThumbCleanup()

// `checkThumbBase()`
//
//	`aDirectory` The thumbnail directory to check.
//	`aDB` The DB handle to access the `Calibre` database.
func checkThumbBase(aDirectory string, aDB *db.TDataBase) {
	subDirs, err := filepath.Glob(aDirectory + "/*")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", aDirectory+"/*", err)
		apachelogger.Err("checkThumbBase()", msg)
		return
	}
	for _, subDir := range subDirs {
		checkThumbDir(subDir, aDB)
	}
} // checkThumbBase()

// `checkThumbDir()` checks `aDirectory` for orphaned thumbnail files.
//
//	`aDirectory` The thumbnail directory to check.
//	`aDB` The DB handle to access the `Calibre` database.
func checkThumbDir(aDirectory string, aDB *db.TDataBase) {
	fileDirs, err := filepath.Glob(aDirectory + "/*.jpg")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", aDirectory+"/*.jpg", err)
		apachelogger.Err("checkThumbDir()", msg)
		return
	}
	for _, fName := range fileDirs {
		checkThumbFile(fName, aDB)
	}
} // checkThumbDir()

// `checkThumbFile()` deletes orphaned thumbnail files.
//
//	`aFilename` The thumbnail file to check.
//	`aDB` The DB handle to access the `Calibre` database.
func checkThumbFile(aFilename string, aDB *db.TDataBase) {
	var msg string
	baseName := path.Base(aFilename)
	docID, err := strconv.Atoi(baseName[:len(baseName)-4])
	if nil != err {
		msg = fmt.Sprintf("strconv.Atoi(%s): %v", baseName[:len(baseName)-4], err)
		apachelogger.Err("checkThumbFile()", msg)
		return
	}

	doc := aDB.QueryDocument(context.Background(), docID)
	if nil == doc {
		// remove thumbnail for non-existing document
		if err = os.Remove(aFilename); nil != err {
			msg = fmt.Sprintf("os.Remove(%s): %v", aFilename, err)
			apachelogger.Err("checkThumbFile()", msg)
		}
		return
	}

	cFile, err := doc.CoverFile()
	if nil != err {
		msg = fmt.Sprintf("doc.CoverFile(%d): %v", docID, err)
		apachelogger.Err("checkThumbFile()", msg)
		return
	}

	tFI, err := os.Stat(aFilename)
	if nil != err {
		msg = fmt.Sprintf("os.Stat(%s): %v", aFilename, err)
		apachelogger.Err("checkThumbFile()", msg)
		return
	}

	cFI, err := os.Stat(cFile)
	if nil != err {
		msg = fmt.Sprintf("os.Stat(%s): %v", cFile, err)
		apachelogger.Err("checkThumbFile()", msg)
		return
	}

	if tFI.ModTime().Before(cFI.ModTime()) {
		// remove outdated thumbnail
		if err = os.Remove(aFilename); nil != err {
			msg = fmt.Sprintf("os.Remove(%s): %v", aFilename, err)
			apachelogger.Err("checkThumbFile()", msg)
		}
		if err = makeThumbnail(cFile, aFilename); nil != err {
			msg = fmt.Sprintf("makeThumbnail(%s): %v", aFilename, err)
			apachelogger.Err("checkThumbFile()", msg)
		}
	}
} // checkThumbFile()

// `makeThumbDir()` creates the directory for the document's thumbnail.
//
// The directory is created with filemode `0775` (`drwxrwxr-x`).
//
//	`aDoc` The document for which to make a thumbnail directory.
func makeThumbDir(aDoc *db.TDocument) error {
	fMode := os.ModeDir | 0775
	fName := thumbnailName(aDoc)
	dName := filepath.Dir(fName)

	return os.MkdirAll(filepath.FromSlash(dName), fMode)
} // makeThumbDir()

// `makeThumbnail()` generates a thumbnail for `aSrcName` and stores it
// in `aDstName`.
//
//	`aSrcName` The filename of a document's cover image.
//	`aDstName` The name of the generated thumbnail file.
func makeThumbnail(aSrcName, aDstName string) error {
	var (
		sImg         image.Image
		err          error
		dFile, sFile *os.File
	)

	if sFile, err = os.OpenFile(aSrcName, os.O_RDONLY, 0); /* #nosec G304 */ nil != err {
		return err
	}
	defer sFile.Close()

	if sImg, _, err = image.Decode(sFile); nil != err {
		return err
	}
	_ = sFile.Close()

	dImg := makeThumbPrim(sImg)
	if dFile, err = os.OpenFile(aDstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640); /* #nosec G302 */ nil != err {
		return err
	}
	defer func() {
		_ = dFile.Close()
		if nil != err {
			_ = os.Remove(aDstName)
		}
	}()
	err = jpeg.Encode(dFile, dImg, &jpeg.Options{Quality: 100})

	return err
} // makeThumbnail()

var (
	// The image's width of the generated thumbnail.
	thThumbwidth uint = 320
)

// Thumbnail will downscale the provided image to max width and height
// preserving the original aspect ratio using the interpolation function
// `resize.Bilinear`.
// It will return original image, without processing it, if original sizes
// are already smaller than provided constraints.
func makeThumbPrim(img image.Image) image.Image {
	origBounds := img.Bounds()
	origWidth, origHeight := uint(origBounds.Dx()), uint(origBounds.Dy())
	newWidth, newHeight := origWidth, origHeight

	// Preserve aspect ratio
	if origWidth > thThumbwidth {
		newHeight = origHeight * thThumbwidth / origWidth
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = thThumbwidth
	}

	return resize.Resize(newWidth, newHeight, img, resize.Bilinear)
} // makeThumbPrim()

// Thumbnail generates a thumbnail of the document's cover.
//
//	`aDoc` The document to check the thumbnail for.
func Thumbnail(aDoc *db.TDocument) (string, error) {
	var (
		err      error
		sName    string
		dFI, sFI os.FileInfo
	)

	// Get the path/filename of the document's cover:
	if sName, err = aDoc.CoverFile(); nil != err {
		return "", err
	}
	if sFI, err = os.Stat(sName); nil != err {
		return "", err
	}
	if !sFI.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", sName)
	}

	dName := thumbnailName(aDoc)
	if dFI, err = os.Stat(dName); nil == err {
		if dFI.ModTime().After(sFI.ModTime()) {
			// dest file exists and is younger than the original cover file
			return dName, nil
		}
	}
	if err = makeThumbDir(aDoc); nil != err {
		return "", err
	}
	if err = makeThumbnail(sName, dName); nil != err {
		return "", err
	}

	return dName, nil
} // Thumbnail()

// `thumbnailName()` returns the name of the thumbnail file of `aDoc`.
//
//	`aDoc` The document for which to compute the thumbnail name.
func thumbnailName(aDoc *db.TDocument) string {
	name := fmt.Sprintf("%06d", aDoc.ID)

	return filepath.Join(db.CalibreCachePath(), name[:4], name+`.jpg`)
} // thumbnailName()

// `thumbnailRemove()` deletes the thumbnail of `aDoc`.
//
// Note that this function is only needed for during testing.
//
//	`aDoc` The document to remove the thumbnail for.
func thumbnailRemove(aDoc *db.TDocument) error {
	fName := thumbnailName(aDoc)
	err := os.Remove(fName)
	if nil == err {
		return nil
	}
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		return nil
	}

	return err
} // thumbnailRemove()

// ThumbnailUpdate creates thumbnails for all existing documents.
func ThumbnailUpdate() {
	// Since this maintenance tasks does not depend on a certain
	// page request we can use `context.Background()` here and
	// open a separate DB connetion.
	ctx := context.Background()
	dbHandle, err := db.OpenDatabase(ctx)
	if nil != err {
		msg := fmt.Sprintf("OpenDatabase(): %v", err)
		apachelogger.Err("ThumbnailUpdate()", msg)
		return
	}

	docList, err := dbHandle.QueryIDs(ctx)
	if nil != err {
		return
	}
	for _, doc := range *docList {
		if _, err = Thumbnail(&doc); nil != err {
			msg := fmt.Sprintf("Thumbnail(%d): %v", doc.ID, err)
			apachelogger.Err("ThumbnailUpdate()", msg)
		}
	}

	// Delete/update all orphaned/outdated thumbnails:
	go goThumbCleanup(dbHandle)
} // ThumbnailUpdate()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// SetThumbWidth set the new width for generated thumbnails.
//
// If `aWidth` is smaller than `64` it's increased to `64`.
//
//	`aWidth` The new thumbnail width to use.
func SetThumbWidth(aWidth uint) uint {
	if 64 > aWidth {
		aWidth = 64
	}
	thThumbwidth = aWidth

	return thThumbwidth
} // SetThumbWidth()

// ThumbWidth returns the configured width of generated thumbnails.
func ThumbWidth() uint {
	return thThumbwidth
} // ThumbWidth()

/* _EoF_ */
