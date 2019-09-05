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

	// Name of the JSON section holding the virtual library definitions.
	mdVirtualLibraries = "virtual_libraries"
)

type (
	// Structure of the `virtual_libraries` JSON section.
	tMdLibList map[string]string

	// TmdVirtLibStruct is a structure to hold a virtual library definition.
	TmdVirtLibStruct struct {
		Def string // Calibre's definitions
		SQL string // SQL: WHERE clause
	}
)

var (
	// cache of "field_metadata" list
	mdFieldsMetadata *map[string]interface{}

	// cache of all DB metadata preferemces
	mdMetadataDbPrefs *map[string]interface{}

	// raw virtual libraries list
	mdVirtLibsRaw *map[string]interface{}

	// virtual libraries list
	mdVirtLibList map[string]TmdVirtLibStruct
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
		apachelogger.Log("md.mdReadMetadataFile", msg)
		return errors.New(msg)
	}
	defer srcFile.Close()

	var jsdata map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsdata); err != nil {
		msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		apachelogger.Log("md.mdReadMetadataFile", msg)
		return errors.New(msg)
	}

	// remove unneeded list entries:
	delete(jsdata, `books view split pane state`)
	delete(jsdata, `column_icon_rules`)
	delete(jsdata, `column_color_rules`)
	delete(jsdata, `cover_grid_icon_rules`)
	delete(jsdata, `gui_view_history`)
	delete(jsdata, `namespaced:CountPagesPlugin:settings`)
	delete(jsdata, `namespaced:FindDuplicatesPlugin:settings`)
	delete(jsdata, `news_to_be_synced`)
	delete(jsdata, `saved_searches`)
	delete(jsdata, `update_all_last_mod_dates_on_start`)
	delete(jsdata, `user_categories`)
	mdMetadataDbPrefs = &jsdata

	return nil
} // mdReadMetadataFile()

// `mdReadFieldMetadata()`
func mdReadFieldMetadata() error {
	if nil != mdFieldsMetadata {
		return nil // field metadata already read
	}
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Log("md.mdReadFieldMetadata", msg)
		return errors.New(msg)
	}
	section, ok := (*mdMetadataDbPrefs)[mdFieldMetadata]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", mdFieldMetadata)
		apachelogger.Log("md.mdReadFieldMetadata", msg)
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
		apachelogger.Log("md.mdGetFieldData", msg)
		return nil, errors.New(msg)
	}

	fmd := *mdFieldsMetadata
	fd, ok := fmd[aField]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", aField)
		apachelogger.Log("md.mdGetFieldData", msg)
		return nil, errors.New(msg)
	}
	result = fd.(map[string]interface{})

	return result, nil
} // mdGetFieldData()

// `mdReadVirtLibs()` reads the raw virt.library definitions.
func mdReadVirtLibs() error {
	if nil != mdVirtLibsRaw {
		return nil
	}
	if err := mdReadMetadataFile(); nil != err {
		msg := fmt.Sprintf("mdReadMetadataFile(): %v", err)
		apachelogger.Log("md.mdReadVirtLibs", msg)
		return errors.New(msg)
	}

	section, ok := (*mdMetadataDbPrefs)[mdVirtualLibraries]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", mdVirtualLibraries)
		apachelogger.Log("md.mdReadVirtLibs", msg)
		return errors.New(msg)
	}
	vlr := section.(map[string]interface{})
	mdVirtLibsRaw = &vlr

	return nil
} // mdReadVirtLibs()

// `mdGetLibDefs()` returns a map of virtual library definitions.
func mdGetLibDefs() (*tMdLibList, error) {
	if err := mdReadVirtLibs(); nil != err {
		msg := fmt.Sprintf("mdReadVirtLibs(): %v", err)
		apachelogger.Log("md.mdGetLibDefs", msg)
		return nil, errors.New(msg)
	}

	m := *mdVirtLibsRaw
	result := make(tMdLibList, len(m))
	for key, value := range m {
		if definition, ok := value.(string); ok {
			result[key] = definition
		} else {
			msg := fmt.Sprintf("json.value.(string): wrong type %v", value)
			apachelogger.Log("md.mdGetLibDefs", msg)
		}
	}

	return &result, nil
} // mdGetLibDefs()

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
	dotStarRE = regexp.MustCompile(`(\.?\*)`)
)

// GetVirtLibList returns a list of virtual library definitions
// and SQL code to access them.
func GetVirtLibList() (map[string]TmdVirtLibStruct, error) {
	if nil != mdVirtLibList {
		return mdVirtLibList, nil
	}
	jsList, err := mdGetLibDefs()
	if nil != err {
		msg := fmt.Sprintf("mdGetLibDefs(): %v", err)
		apachelogger.Log("md.GetVirtLibList", msg)
		return nil, err
	}

	mdVirtLibList := make(map[string]TmdVirtLibStruct, len(*jsList))
	for key, value := range *jsList {
		vl := NewSearch(value).Parse()

		mdVirtLibList[key] = TmdVirtLibStruct{
			Def: dotStarRE.ReplaceAllLiteralString(value, "%"),
			SQL: vl.Where(),
		}
	}

	return mdVirtLibList, nil
} // GetVirtLibList()

// GetVirtLibOptions returns the SELECT/OPTIONs of virtual libraries.
//
//	aSelected Name of the currently selected library.
func GetVirtLibOptions(aSelected string) string {
	_, err := GetVirtLibList()
	if nil != err {
		msg := fmt.Sprintf("GetVirtLibList(): %v", err)
		apachelogger.Log("md.GetVirtLibOptions", msg)
		return ""
	}

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
} // GetVirtLibOptions()

/* _EoF_ */
