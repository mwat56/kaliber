/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/mwat56/apachelogger"
)

type (
	// structure of the `virtual_libraries` JSON section
	tVirtLibMap map[string]string
)

const (
	// Calibre's metadata/preferences store
	calibrePreferencesFile = "metadata_db_prefs_backup.json"

	// name of the JSON section holding the virtual library definitions
	virtLibSection = "virtual_libraries"
)

// `readJSONmetaDataFile()` reads `aFilename` and returns a map of
// the JSON data read.
func readJSONmetaDataFile(aFilename string) (*map[string]interface{}, error) {
	srcFile, err := os.OpenFile(aFilename, os.O_RDONLY, 0)
	if nil != err {
		msg := fmt.Sprintf("os.OpenFile(%s): %v", aFilename, err)
		apachelogger.Log("virtlib.readMetaDataFile", msg)
		return nil, err
	}
	defer srcFile.Close()

	var jsdata map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsdata); err != nil {
		msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		apachelogger.Log("virtlib.readMetaDataFile", msg)
		return nil, err
	}

	return &jsdata, nil
} // readJSONmetaDataFile()

// `readJSONvirtualLibs()` reads `aFilename` and returns a map of
// virtual library definitions.
func readJSONvirtualLibs(aFilename string) (*tVirtLibMap, error) {
	jsdata, err := readJSONmetaDataFile(aFilename)
	if nil != err {
		msg := fmt.Sprintf("readJSONmetaDataFile(%s): %v", aFilename, err)
		apachelogger.Log("virtlib.readJSONvirtualLibs", msg)
		return nil, err
	}
	section, ok := (*jsdata)[virtLibSection]
	if !ok {
		msg := fmt.Sprintf("no such JSON section: %s", virtLibSection)
		apachelogger.Log("virtlib.readJSONvirtualLibs", msg)
		return nil, errors.New(msg)
	}

	m := section.(map[string]interface{})
	result := make(tVirtLibMap, len(m))
	for key, value := range m {
		if definition, ok := value.(string); ok {
			result[key] = definition
		} else {
			msg := fmt.Sprintf("json.value.(string): wrong type %v", value)
			apachelogger.Log("virtlib.readJSONvirtualLibs", msg)
		}
	}

	return &result, nil
} // readJSONvirtualLibs()

/* _EoF_ */
