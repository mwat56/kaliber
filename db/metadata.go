/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * This file provides functions to access Calibre`s metadata store.
 */

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/mwat56/apachelogger"
)

const (
	// Name of the JSON section holding the names of book fields.
	mdBookDisplayFields = `book_display_fields`

	// Name of the JSON section holding e.g. user-defined field definitions.
	mdFieldMetadata = `field_metadata`

	// Name of the JSON section holding the virtual library names to hide..
	mdHiddenVirtualLibraries = `virt_libs_hidden`

	// Name of the JSON section holding the virtual library definitions.
	mdVirtualLibraries = `virtual_libraries`
)

type (
	// tBookDisplayFieldsList is the `book_display_fields` metadata list.
	tBookDisplayFieldsList map[string]bool

	tInterfaceList map[string]interface{}

	// TVirtLibList is the `virtual_libraries` JSON metadata section
	// indexed by virt.library name.
	TVirtLibList map[string]string
)

var (
	// cache of "book_display_fields" list
	mdBookDisplayFieldsList    tBookDisplayFieldsList
	mdBookDisplayFieldsListMtx = new(sync.RWMutex)

	// cache of "field_metadata" list
	mdFieldsMetadataList    *tInterfaceList // map[string]interface{}
	mdFieldsMetadataListMtx = new(sync.RWMutex)

	// list of virtual libraries to hide
	mdHiddenVirtLibs    *tInterfaceList // map[string]interface{}
	mdHiddenVirtLibsMtx = new(sync.RWMutex)

	// cache of all DB metadata preferences
	mdMetadataDbPrefs    *tInterfaceList // map[string]interface{}
	mdMetadataDbPrefsMtx = new(sync.RWMutex)

	// virtual libraries list
	mdVirtLibList    TVirtLibList
	mdVirtLibListMtx = new(sync.RWMutex)

	// raw virtual libraries list
	mdVirtLibsRaw    *tInterfaceList // map[string]interface{}
	mdVirtLibsRawMtx = new(sync.RWMutex)
)

// `mdGetFieldData()` returns a list of field definitions for `aField`.
func mdGetFieldData(aField string) (rList tInterfaceList /* map[string]interface{} */, rErr error) {
	if 0 == len(aField) {
		return
	}
	if rErr = mdReadFieldMetadata(); nil != rErr {
		msg := fmt.Sprintf("mdReadFieldMetadata(): %v", rErr)
		rErr = errors.New(msg)
		return
	}
	mdFieldsMetadataListMtx.RLock()
	defer mdFieldsMetadataListMtx.RUnlock()

	fmd := *mdFieldsMetadataList
	fd, ok := fmd[aField]
	if !ok {
		return nil, errors.New("no such JSON section: " + aField)
	}
	lst := fd.(map[string]interface{})
	rList = tInterfaceList(lst)

	return
} // mdGetFieldData()

// `mdReadBookDisplayFields()`
func mdReadBookDisplayFields() error {
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		return errors.New(msg)
	}

	section, ok := mdGetMetadataDbPref(mdBookDisplayFields)
	if !ok {
		return errors.New("no such JSON section: " + mdBookDisplayFields)
	}

	mdBookDisplayFieldsListMtx.Lock()
	defer mdBookDisplayFieldsListMtx.Unlock()

	if nil != mdBookDisplayFieldsList {
		return nil // field metadata already read
	}

	data := section.([]interface{})
	mdBookDisplayFieldsList = make(tBookDisplayFieldsList, len(data))
	for _, raw := range data {
		entry := raw.([]interface{})
		field := entry[0].(string)
		display := entry[1].(bool)
		mdBookDisplayFieldsList[field] = display
	}

	return nil
} // mdReadBookDisplayFields()

// `mdReadFieldMetadata()`
func mdReadFieldMetadata() error {
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		return errors.New(msg)
	}

	section, ok := mdGetMetadataDbPref(mdFieldMetadata)
	if !ok {
		return errors.New("no such JSON section: " + mdFieldMetadata)
	}

	mdFieldsMetadataListMtx.Lock()
	defer mdFieldsMetadataListMtx.Unlock()

	if nil != mdFieldsMetadataList {
		return nil // field metadata already read
	}

	fmd := section.(map[string]interface{})
	msi := tInterfaceList(fmd)
	mdFieldsMetadataList = &msi

	return nil
} // mdReadFieldMetadata()

// `mdReadHiddenVirtualLibraries()` reads the list ob hidden libraries to hide.
func mdReadHiddenVirtualLibraries() error {
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Err("mdReadHiddenVirtualLibraries", msg)
		return errors.New(msg)
	}

	section, ok := mdGetMetadataDbPref(mdHiddenVirtualLibraries)
	if !ok {
		msg := "no such JSON section: " + mdHiddenVirtualLibraries
		apachelogger.Err("mdReadHiddenVirtualLibraries", msg)
		return errors.New(msg)
	}

	mdHiddenVirtLibsMtx.Lock()
	defer mdHiddenVirtLibsMtx.Unlock()

	if nil != mdHiddenVirtLibs {
		return nil
	}

	hvl := section.([]interface{})
	if 0 == len(hvl) {
		return nil
	}
	result := make(tInterfaceList /* map[string]interface{} */, len(hvl))
	for _, val := range hvl {
		lib := val.(string)
		result[lib] = struct{}{}
	}
	mdHiddenVirtLibs = &result

	return nil
} // mdReadHiddenVirtualLibraries()

// `mdGetMetadataDbPref()` returns the preferences section indexed by `aKey`.
func mdGetMetadataDbPref(aKey string) (rSection interface{}, rOK bool) {
	mdMetadataDbPrefsMtx.RLock()
	defer mdMetadataDbPrefsMtx.RUnlock()

	rSection, rOK = (*mdMetadataDbPrefs)[aKey]

	return
} // mdGetMetadataDbPref()

// `mdReadMetadataFile()` returns a map of the JSON data read.
func mdReadMetadataFile() error {
	mdMetadataDbPrefsMtx.Lock()
	defer mdMetadataDbPrefsMtx.Unlock()

	if nil != mdMetadataDbPrefs {
		return nil // metadata already read
	}

	fName := CalibrePreferencesFile()
	srcFile, err := os.OpenFile(fName, os.O_RDONLY, 0)
	if nil != err {
		msg := fmt.Sprintf("os.OpenFile(%s): %v", fName, err)
		return errors.New(msg)
	}
	defer srcFile.Close()

	var jsData tInterfaceList // map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsData); err != nil {
		msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		return errors.New(msg)
	}

	// remove unneeded list entries:
	delete(jsData, `column_icon_rules`)
	delete(jsData, `cover_grid_icon_rules`)
	delete(jsData, `gui_view_history`)
	delete(jsData, `namespaced:CountPagesPlugin:settings`)
	delete(jsData, `namespaced:FindDuplicatesPlugin:settings`)
	delete(jsData, `news_to_be_synced`)
	delete(jsData, `saved_searches`)
	delete(jsData, `update_all_last_mod_dates_on_start`)
	delete(jsData, `user_categories`)
	mdMetadataDbPrefs = &jsData

	return nil
} // mdReadMetadataFile()

// `mdReadVirtualLibraries()` reads the raw virt.library definitions.
func mdReadVirtualLibraries() error {
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Err("mdReadVirtualLibraries()", msg)
		return errors.New(msg)
	}

	section, ok := mdGetMetadataDbPref(mdVirtualLibraries)
	if !ok {
		msg := "no such JSON section: " + mdVirtualLibraries
		apachelogger.Err("mdReadVirtualLibraries()", msg)
		return errors.New(msg)
	}

	mdVirtLibsRawMtx.Lock()
	defer mdVirtLibsRawMtx.Unlock()

	if nil != mdVirtLibsRaw {
		return nil
	}

	vlr := section.(map[string]interface{})
	msi := tInterfaceList(vlr)
	mdVirtLibsRaw = &msi

	return nil
} // mdReadVirtualLibraries()

// `mdVirtLibDefinitions()` returns a map of virtual library definitions.
func mdVirtLibDefinitions() (*TVirtLibList, error) {
	if err := mdReadVirtualLibraries(); nil != err {
		msg := fmt.Sprintf("mdReadVirtualLibraries(): %v", err)
		apachelogger.Err("mdVirtualLibDefinitions()", msg)
		return nil, errors.New(msg)
	}
	if err := mdReadHiddenVirtualLibraries(); nil != err {
		msg := fmt.Sprintf("mdReadHiddenVirtualLibraries(): %v", err)
		apachelogger.Err("mdVirtualLibDefinitions()", msg)
		return nil, errors.New(msg)
	}

	mdVirtLibsRawMtx.RLock()
	defer mdVirtLibsRawMtx.RUnlock()

	m := *mdVirtLibsRaw
	result := make(TVirtLibList, len(m))
	for key, value := range m {
		//FIXME MUTEX
		if nil != mdHiddenVirtLibs {
			if _, ok := (*mdHiddenVirtLibs)[key]; ok {
				continue
			}
		}
		if definition, ok := value.(string); ok {
			result[key] = definition
		} else {
			msg := fmt.Sprintf("json.value.(string): wrong type %v", value)
			apachelogger.Err("mdVirtualLibDefinitions", msg)
		}
	}

	return &result, nil
} // mdVirtualLibDefinitions()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// BookFieldVisible returns whether `aFieldname` should be visible or not.
//
// If `aFieldname` can't be found the function returns `true` and an error,
// otherwise the (boolean) `visible` value and `nil`.
//
//	`aFieldname` The name of the field/column to check.
func BookFieldVisible(aFieldname string) (bool, error) {
	if err := mdReadBookDisplayFields(); nil != err {
		msg := fmt.Sprintf("mdReadBookDisplayFields(): %v", err)
		apachelogger.Err("md.BookFieldVisible()", msg)
		return true, errors.New(msg)
	}
	mdBookDisplayFieldsListMtx.RLock()
	defer mdBookDisplayFieldsListMtx.RUnlock()

	if result, ok := mdBookDisplayFieldsList[aFieldname]; ok {
		return result, nil
	}

	msg := "field name doesn't exist: " + aFieldname
	apachelogger.Err("md.BookFieldVisible()", msg)

	return true, errors.New(msg)
} // BookFieldVisible()

// MetaFieldValue returns the value of `aField` of `aSection`.
//
//	`aSection` Name of the field's metadata section.
//	`aField` Name of the data field within `aSection`.
func MetaFieldValue(aSection, aField string) (interface{}, error) {
	if (0 == len(aSection)) || (0 == len(aField)) {
		msg := fmt.Sprintf(`md.MetaFieldValue(): empty arguments ("%s". "%s")`, aSection, aField)
		apachelogger.Err("md.MetaFieldValue", msg)
		return nil, errors.New(msg)
	}

	fmd, err := mdGetFieldData(aSection)
	if nil != err {
		msg := fmt.Sprintf("mdGetFieldData(): %v", err)
		apachelogger.Err("md.MetaFieldValue", msg)
		return nil, errors.New(msg)
	}

	result, ok := fmd[aField]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s[%s]", aSection, aField)
		apachelogger.Err("md.MetaFieldValue", msg)
		return nil, errors.New(msg)
	}

	return result, nil
} // MetaFieldValue()

// VirtLibOptions returns the SELECT/OPTIONs of the virtual libraries.
//
//	`aSelected` Name of the currently selected library.
func VirtLibOptions(aSelected string) string {
	_, err := VirtualLibraryList()
	if nil != err {
		msg := fmt.Sprintf("md.VirtualLibraryList(): %v", err)
		apachelogger.Err("md.VirtLibOptions", msg)
		return ""
	}
	mdVirtLibListMtx.RLock()
	defer mdVirtLibListMtx.RUnlock()

	list := make([]string, 0, len(mdVirtLibList)+1)
	if (0 == len(aSelected)) || ("-" == aSelected) {
		list = append(list, `<option value="-" SELECTED> – </option>`)
		aSelected = ""
	} else {
		list = append(list, `<option value="-"> – </option>`)
	}
	for key := range mdVirtLibList {
		option := `<option value="` + key + `"`
		if key == aSelected {
			option += ` SELECTED`
			aSelected = ""
		}
		option += `>` + key + `</option>`
		list = append(list, option)
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i]) < strings.ToLower(list[j])
	})

	return strings.Join(list, "\n")
} // VirtLibOptions()

var (
	// RegEx to find `.*` in a virt.lib. definition
	mdDotStarRE = regexp.MustCompile(`(\.?\*)`)
)

// VirtualLibraryList returns a list of virtual library definitions
// and SQL code to access them.
func VirtualLibraryList() (TVirtLibList, error) {
	mdVirtLibListMtx.Lock()
	defer mdVirtLibListMtx.Unlock()

	if nil != mdVirtLibList {
		return mdVirtLibList, nil
	}

	jsList, err := mdVirtLibDefinitions()
	if nil != err {
		msg := fmt.Sprintf("mdVirtLibDefinitions(): %v", err)
		apachelogger.Err("md.VirtualLibraryList()", msg)
		return nil, err
	}

	mdVirtLibList = make(TVirtLibList, len(*jsList))
	for key, value := range *jsList {

		//TODO check for libraries to hide

		mdVirtLibList[key] = mdDotStarRE.ReplaceAllLiteralString(value, "%")
	}

	return mdVirtLibList, nil
} // VirtualLibraryList()

/* _EoF_ */
