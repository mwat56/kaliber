/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/mwat56/apachelogger"
	"github.com/nfnt/resize"
)

/*
 * This file provides functions for thumbnail generation.
 */

var (
	// The image's width of the generated thumbnail.
	thumbwidth uint = 320
)

// `goThumbCleaup()` removes orphaned thumbnails.
func goThumbCleanup() {
	bd := CalibreCachePath()
	dirNames, err := filepath.Glob(bd + "/*")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", bd, err)
		apachelogger.Log("goThumbCleaup()", msg)
		return
	}
	for _, numDir := range dirNames {
		checkThumbBase(numDir)
	}
} // goThumbCleanup()

// `checkThumbBase()`
func checkThumbBase(aDirectory string) {
	subDirs, err := filepath.Glob(aDirectory + "/*")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", aDirectory+"/*", err)
		apachelogger.Log("checkThumbBase()", msg)
		return
	}
	for _, subDir := range subDirs {
		checkThumbDir(subDir)
	}
} // checkThumbBase()

// `checkThumbDir()`
func checkThumbDir(aDirectory string) {
	fileDirs, err := filepath.Glob(aDirectory + "/*.jpg")
	if nil != err {
		msg := fmt.Sprintf("filepath.Glob(%s): %v", aDirectory+"/*.jpg", err)
		apachelogger.Log("checkThumbDir()", msg)
		return
	}
	for _, fName := range fileDirs {
		checkThumbFile(fName)
	}
} // checkThumbDir()

// `checkThumbFile()` deletes orphaned thumbnail files.
func checkThumbFile(aFilename string) {
	baseName := path.Base(aFilename)
	docID, err := strconv.Atoi(baseName[:len(baseName)-4])
	if nil != err {
		msg := fmt.Sprintf("strconv.Atoi(%s): %v", baseName[:len(baseName)-4], err)
		apachelogger.Log("checkThumbFile()", msg)
		return
	}

	doc := QueryDocument(docID)
	if nil == doc {
		// remove thumbnail for non-existing document
		if err = os.Remove(aFilename); nil != err {
			msg := fmt.Sprintf("os.Remove(%s): %v", aFilename, err)
			apachelogger.Log("checkThumbFile()", msg)
		}
		return
	}
	cFile, err := doc.CoverFile()
	if nil != err {
		msg := fmt.Sprintf("doc.CoverFile(%d): %v", docID, err)
		apachelogger.Log("checkThumbFile()", msg)
		return
	}

	tFI, err := os.Stat(aFilename)
	if nil != err {
		msg := fmt.Sprintf("os.Stat(%s): %v", aFilename, err)
		apachelogger.Log("checkThumbFile()", msg)
		return
	}

	cFI, err := os.Stat(cFile)
	if nil != err {
		msg := fmt.Sprintf("os.Stat(%s): %v", cFile, err)
		apachelogger.Log("checkThumbFile()", msg)
		return
	}

	if tFI.ModTime().Before(cFI.ModTime()) {
		// remove outdated thumbnail
		if err = os.Remove(aFilename); nil != err {
			msg := fmt.Sprintf("os.Remove(%s): %v", aFilename, err)
			apachelogger.Log("checkThumbFile()", msg)
		}
		if err = makeThumbnail(cFile, aFilename); nil != err {
			msg := fmt.Sprintf("makeThumbnail(%s): %v", aFilename, err)
			apachelogger.Log("checkThumbFile()", msg)
		}
	}
} // checkThumbFile()

// `makeThumbDir()` creates the directory for the document's thumbnail.
//
// The directory is created with filemode `0775` (`drwxrwxr-x`).
func makeThumbDir(aDoc *TDocument) error {
	fmode := os.ModeDir | 0775
	fName := ThumbnailName(aDoc)
	dName := filepath.Dir(fName)

	return os.MkdirAll(filepath.FromSlash(dName), fmode)
} // makeThumbDir()

// `makeThumbnail()` generates a thumbnail for `aSrcName` and stores it
// in `aDstName`.
func makeThumbnail(aSrcName, aDstName string) error {
	var (
		sImg         image.Image
		err          error
		dFile, sFile *os.File
	)

	if sFile, err = os.Open(aSrcName); /* #nosec G304 */ nil != err {
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

// Thumbnail will downscale provided image to max width and height preserving
// original aspect ratio and using the interpolation function interp.
// It will return original image, without processing it, if original sizes
// are already smaller than provided constraints.
func makeThumbPrim(img image.Image) image.Image {
	origBounds := img.Bounds()
	origWidth := uint(origBounds.Dx())
	origHeight := uint(origBounds.Dy())
	newWidth, newHeight := origWidth, origHeight

	// Preserve aspect ratio
	if origWidth > thumbwidth {
		newHeight = origHeight * thumbwidth / origWidth
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = thumbwidth
	}

	return resize.Resize(newWidth, newHeight, img, resize.Bilinear)
} // makeThumbPrim()

// SetThumbWidth set the new width for generated thumbnails.
//
// If `aWidth` is smaller than `64` it's increased.
func SetThumbWidth(aWidth uint) uint {
	if 64 > aWidth {
		aWidth = 64
	}
	thumbwidth = aWidth

	return thumbwidth
} // SetThumbWidth()

// Thumbnail generates a thumbnail of the document's cover.
func Thumbnail(aDoc *TDocument) (string, error) {
	var (
		err      error
		sName    string
		dFI, sFI os.FileInfo
	)

	// Get the path/filename of the document's cover:
	if sName, err = aDoc.coverAbs(false); nil != err {
		return "", err
	}
	if sFI, err = os.Stat(sName); nil != err {
		return "", err
	}
	if !sFI.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", sName)
	}

	dName := ThumbnailName(aDoc)
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

// ThumbnailName returns the name of thumbnail file of `aDoc`.
func ThumbnailName(aDoc *TDocument) string {
	name := fmt.Sprintf("%06d", aDoc.ID)

	return filepath.Join(CalibreCachePath(), name[:4], name+`.jpg`)
} // ThumbnailName

// ThumbnailRemove deletes the thumbnail of `aDoc`.
func ThumbnailRemove(aDoc *TDocument) error {
	fName := ThumbnailName(aDoc)
	err := os.Remove(fName)
	if nil == err {
		return nil
	}
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		return nil
	}

	return err
} // ThumbnailRemove

// ThumbnailUpdate creates thumbnails for all existing documents
func ThumbnailUpdate() {
	docList, err := QueryIDs()
	if nil != err {
		return
	}
	for _, doc := range *docList {
		// here we ignore all errors but hope for the best
		if _, err = Thumbnail(&doc); nil != err {
			msg := fmt.Sprintf("Thumbnail(%d): %v", doc.ID, err)
			apachelogger.Log("ThumbnailUpdate()", msg)
		}
	}

	// delete/update all orphaned/outdated thumbnails:
	go goThumbCleanup()
} // ThumbnailUpdate()

// ThumbWidth returns the configured width of generated thumbnails.
func ThumbWidth() uint {
	return thumbwidth
} // ThumbWidth()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

/* _EoF_ */
