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

	"github.com/mwat56/apachelogger"
)

type (
	// Structure of the `virtual_libraries` JSON section
	tVirtLibJSON map[string]string

	// Structure to hold a virtual library definition.
	tVirtLibStruct struct {
		Def string // Calibre's definitions
		SQL string // SQL: WHERE clause
	}

	// TvirtLibMap is a list of virt.lib. definitions
	TvirtLibMap map[string]tVirtLibStruct
)

const (
	// name of the JSON section holding the virtual library definitions
	virtlibJSONsection = "virtual_libraries"
)

// `virtlibReadJSONmetadata()` reads `aFilename` and returns a map of
// the JSON data read.
//
//	aFilename The path/filename of Calibre's metadata JSON file.
func virtlibReadJSONmetadata() (*map[string]interface{}, error) {
	fName := CalibrePreferencesPath()
	srcFile, err := os.OpenFile(fName, os.O_RDONLY, 0)
	if nil != err {
		msg := fmt.Sprintf("os.OpenFile(%s): %v", fName, err)
		apachelogger.Log("virtlib.virtlibReadJSONmetadata", msg)
		return nil, err
	}
	defer srcFile.Close()

	var jsdata map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsdata); err != nil {
		msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		apachelogger.Log("virtlib.virtlibReadJSONmetadata", msg)
		return nil, err
	}

	// remove unneeded list entries:
	delete(jsdata, `books view split pane state`)
	delete(jsdata, `column_icon_rules`)
	delete(jsdata, `column_color_rules`)
	delete(jsdata, `cover_grid_icon_rules`)
	delete(jsdata, `field_metadata`) // ? keep ?
	delete(jsdata, `gui_view_history`)
	delete(jsdata, `namespaced:CountPagesPlugin:settings`)
	delete(jsdata, `namespaced:FindDuplicatesPlugin:settings`)
	delete(jsdata, `news_to_be_synced`)
	delete(jsdata, `saved_searches`)
	delete(jsdata, `update_all_last_mod_dates_on_start`)
	delete(jsdata, `user_categories`)

	return &jsdata, nil
} // virtlibReadJSONmetadata()

// `virtlibGetLibDefs()` reads `aFilename` and returns a map of
// virtual library definitions.
func virtlibGetLibDefs() (*tVirtLibJSON, error) {
	jsdata, err := virtlibReadJSONmetadata()
	if nil != err {
		msg := fmt.Sprintf("virtlibReadJSONmetadata(): %v", err)
		apachelogger.Log("virtlib.virtlibGetLibDefs", msg)
		return nil, err
	}
	section, ok := (*jsdata)[virtlibJSONsection]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", virtlibJSONsection)
		apachelogger.Log("virtlib.virtlibGetLibDefs", msg)
		return nil, errors.New(msg)
	}

	m := section.(map[string]interface{})
	result := make(tVirtLibJSON, len(m))
	for key, value := range m {
		if definition, ok := value.(string); ok {
			result[key] = definition
		} else {
			msg := fmt.Sprintf("json.value.(string): wrong type %v", value)
			apachelogger.Log("virtlib.virtlibGetLibDefs", msg)
		}
	}

	return &result, nil
} // virtlibGetLibDefs()

// GetVirtLibList reads `aFilename` and returns a list of virtual
// library definitions and SQL code to access them.
func GetVirtLibList() (*TvirtLibMap, error) {
	jsList, err := virtlibGetLibDefs()
	if nil != err {
		msg := fmt.Sprintf("virtlibGetLibDefs(): %v", err)
		apachelogger.Log("virtlib.GetVirtLibList", msg)
		return nil, err
	}
	result := make(TvirtLibMap, len(*jsList))
	for key, value := range *jsList {
		vl := NewSearch(value).Parse()
		result[key] = tVirtLibStruct{
			Def: value,
			SQL: vl.Where(),
		}
	}

	return &result, nil
} // GetVirtLibList()

/* _EoF_ */
