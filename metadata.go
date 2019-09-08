/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mwat56/apachelogger"
)

const (
	// Name of the JSON section holding e.g. user-defined field definitions.
	mdFieldMetadata = "field_metadata"

	// Name of the JSON section holding the virtual library names to hide..
	mdHiddenVirtualLibraries = "virt_libs_hidden"

	// Name of the JSON section holding the virtual library definitions.
	mdVirtualLibraries = "virtual_libraries"
)

type (
	// TVirtualLibraryList is the `virtual_libraries` JSON metadata section.
	TVirtualLibraryList map[string]string
)

var (
	// cache of "field_metadata" list
	mdFieldsMetadata *map[string]interface{}

	// list of virtual libraries to hide
	mdHiddenVirtLibs map[string]interface{}

	// cache of all DB metadata preferences
	mdMetadataDbPrefs map[string]interface{}

	// virtual libraries list
	mdVirturalLibraryList TVirtualLibraryList

	// raw virtual libraries list
	mdVirtLibsRaw *map[string]interface{}
)

// `mdReadMetadataFile()` returns a map of the JSON data read.
func mdReadMetadataFile() error {
	if nil != mdMetadataDbPrefs {
		return nil // metadata already read
	}
	fName := CalibrePreferencesFile()
	srcFile, err := os.OpenFile(fName, os.O_RDONLY, 0)
	if nil != err {
		msg := fmt.Sprintf("os.OpenFile(%s): %v", fName, err)
		apachelogger.Log("mdReadMetadataFile", msg)
		return errors.New(msg)
	}
	defer srcFile.Close()

	var jsdata map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsdata); err != nil {
		msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		apachelogger.Log("mdReadMetadataFile", msg)
		return errors.New(msg)
	}

	// remove unneeded list entries:
	delete(jsdata, `column_icon_rules`)
	delete(jsdata, `cover_grid_icon_rules`)
	delete(jsdata, `gui_view_history`)
	delete(jsdata, `namespaced:CountPagesPlugin:settings`)
	delete(jsdata, `namespaced:FindDuplicatesPlugin:settings`)
	delete(jsdata, `news_to_be_synced`)
	delete(jsdata, `saved_searches`)
	delete(jsdata, `update_all_last_mod_dates_on_start`)
	delete(jsdata, `user_categories`)
	mdMetadataDbPrefs = jsdata

	return nil
} // mdReadMetadataFile()

// `mdReadFieldMetadata()`
func mdReadFieldMetadata() error {
	if nil != mdFieldsMetadata {
		return nil // field metadata already read
	}
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Log("mdReadFieldMetadata", msg)
		return errors.New(msg)
	}
	section, ok := mdMetadataDbPrefs[mdFieldMetadata]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", mdFieldMetadata)
		apachelogger.Log("mdReadFieldMetadata", msg)
		return errors.New(msg)
	}
	fmd := section.(map[string]interface{})
	mdFieldsMetadata = &fmd

	return nil
} // mdReadFieldMetadata()

// `mdGetFieldData()` returns a list of field definitions for `aField`.
func mdGetFieldData(aField string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if 0 == len(aField) {
		return result, nil
	}
	if err := mdReadFieldMetadata(); nil != err {
		msg := fmt.Sprintf("mdReadFieldMetadata(): %v", err)
		apachelogger.Log("mdGetFieldData", msg)
		return nil, errors.New(msg)
	}

	fmd := *mdFieldsMetadata
	fd, ok := fmd[aField]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", aField)
		apachelogger.Log("mdGetFieldData", msg)
		return nil, errors.New(msg)
	}
	result = fd.(map[string]interface{})

	return result, nil
} // mdGetFieldData()

// `mdReadHiddenVirtualLibraries()` reads the list ob hidden libraries to hide.
func mdReadHiddenVirtualLibraries() error {
	if nil != mdHiddenVirtLibs {
		return nil
	}
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Log("mdReadHiddenVirtualLibraries", msg)
		return errors.New(msg)
	}

	section, ok := mdMetadataDbPrefs[mdHiddenVirtualLibraries]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", mdHiddenVirtualLibraries)
		apachelogger.Log("mdReadHiddenVirtualLibraries", msg)
		return errors.New(msg)
	}

	hvl := section.([]interface{})
	if 0 == len(hvl) {
		return nil
	}
	mdHiddenVirtLibs = make(map[string]interface{}, len(hvl))
	for _, val := range hvl {
		lib := val.(string)
		mdHiddenVirtLibs[lib] = struct{}{}
	}

	return nil
} // mdReadHiddenVirtualLibraries()

// `mdReadVirtualLibraries()` reads the raw virt.library definitions.
func mdReadVirtualLibraries() error {
	if nil != mdVirtLibsRaw {
		return nil
	}
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Log("mdReadVirtualLibraries", msg)
		return errors.New(msg)
	}

	section, ok := mdMetadataDbPrefs[mdVirtualLibraries]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", mdVirtualLibraries)
		apachelogger.Log("mdReadVirtualLibraries", msg)
		return errors.New(msg)
	}
	vlr := section.(map[string]interface{})
	mdVirtLibsRaw = &vlr

	return nil
} // mdReadVirtualLibraries()

// `mdVirtualLibDefinitions()` returns a map of virtual library definitions.
func mdVirtualLibDefinitions() (*TVirtualLibraryList, error) {
	if err := mdReadVirtualLibraries(); nil != err {
		msg := fmt.Sprintf("mdReadVirtualLibraries(): %v", err)
		apachelogger.Log("mdVirtualLibDefinitions", msg)
		return nil, errors.New(msg)
	}
	if err := mdReadHiddenVirtualLibraries(); nil != err {
		msg := fmt.Sprintf("mdReadHiddenVirtualLibraries(): %v", err)
		apachelogger.Log("mdVirtualLibDefinitions", msg)
		return nil, errors.New(msg)
	}

	m := *mdVirtLibsRaw
	result := make(TVirtualLibraryList, len(m))
	for key, value := range m {
		if nil != mdHiddenVirtLibs {
			if _, ok := mdHiddenVirtLibs[key]; ok {
				continue
			}
		}
		if definition, ok := value.(string); ok {
			result[key] = definition
		} else {
			msg := fmt.Sprintf("json.value.(string): wrong type %v", value)
			apachelogger.Log("mdVirtualLibDefinitions", msg)
		}
	}

	return &result, nil
} // mdVirtualLibDefinitions()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// GetMetaFieldValue returns the value of `aField` of `aSection`.
//
//	aSection Name of the field metadata section.
//	aField Name of the data field within `aSection`.
func GetMetaFieldValue(aSection, aField string) (interface{}, error) {
	if (0 == len(aSection)) || (0 == len(aField)) {
		msg := fmt.Sprintf(`GetMetaFieldValue(): empty arguments ("%s". "%s")`, aSection, aField)
		return nil, errors.New(msg)
	}

	fmd, err := mdGetFieldData(aSection)
	if nil != err {
		msg := fmt.Sprintf("mdGetFieldData(): %v", err)
		apachelogger.Log("md.GetMetaFieldValue", msg)
		return nil, errors.New(msg)
	}

	result, ok := fmd[aField]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s[%s]", aSection, aField)
		apachelogger.Log("md.GetMetaFieldValue", msg)
		return nil, errors.New(msg)
	}

	return result, nil
} // GetMetaFieldValue()

var (
	// RegEx to find `.*` in a virt.lib. definition
	mdDotStarRE = regexp.MustCompile(`(\.?\*)`)
)

// GetVirtualLibraryList returns a list of virtual library definitions
// and SQL code to access them.
func GetVirtualLibraryList() (TVirtualLibraryList, error) {
	if nil != mdVirturalLibraryList {
		return mdVirturalLibraryList, nil
	}
	jsList, err := mdVirtualLibDefinitions()
	if nil != err {
		msg := fmt.Sprintf("mdVirtualLibDefinitions(): %v", err)
		apachelogger.Log("md.GetVirtLibList", msg)
		return nil, err
	}

	mdVirturalLibraryList = make(TVirtualLibraryList, len(*jsList))
	for key, value := range *jsList {

		//TODO check for libraries to hide

		mdVirturalLibraryList[key] = mdDotStarRE.ReplaceAllLiteralString(value, "%")
	}

	return mdVirturalLibraryList, nil
} // GetVirtualLibraryList()

// GetVirtLibOptions returns the SELECT/OPTIONs of virtual libraries.
//
//	aSelected Name of the currently selected library.
func GetVirtLibOptions(aSelected string) string {
	_, err := GetVirtualLibraryList()
	if nil != err {
		msg := fmt.Sprintf("GetVirtLibList(): %v", err)
		apachelogger.Log("md.GetVirtLibOptions", msg)
		return ""
	}

	list := make([]string, 0, len(mdVirturalLibraryList)+1)
	if (0 == len(aSelected)) || ("-" == aSelected) {
		list = append(list, `<option value="-" SELECTED> – </option>`)
		aSelected = ""
	} else {
		list = append(list, `<option value="-"> – </option>`)
	}
	for key := range mdVirturalLibraryList {
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
} // GetVirtLibOptions()

/* _EoF_ */
