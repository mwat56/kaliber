/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type (
	// structure of the `virtual_libraries` JSON section
	tVirtLibMap map[string]string
)

const (
	// Calibre's metadata/preferences store
	virtLibFile = "metadata_db_prefs_backup.json"

	virtLibSection = "virtual_libraries"
)

// `readJSONmetaDataFile()` reads `aFilename`
func readJSONmetaDataFile(aFilename string) (*tVirtLibMap, error) {
	srcFile, err := os.OpenFile(aFilename, os.O_RDONLY, 0)
	if nil != err {
		log.Println("os.OpenFile():", aFilename, err)
		// msg := fmt.Sprintf("os.OpenFile(%s): %v", aFilename, err)
		// apachelogger.Log("virtlib.readMetaDataFile()", msg)
		return nil, err
	}
	defer srcFile.Close()

	var jsdata map[string]interface{}
	dec := json.NewDecoder(srcFile)
	if err := dec.Decode(&jsdata); err != nil {
		log.Println("json.NewDecoder.Decode():", err)
		// msg := fmt.Sprintf("json.NewDecoder.Decode(): %v", err)
		// apachelogger.Log("virtlib.readMetaDataFile()", msg)
		return nil, err
	}

	section, ok := jsdata[virtLibSection]
	if !ok {
		err = fmt.Errorf("no such JSON section: %s", virtLibSection)
		log.Println("virtlib.readMetaDataFile():", err)
		return nil, err
	}
	m := section.(map[string]interface{})
	result := make(tVirtLibMap, len(m))
	for key, value := range m {
		switch vv := value.(type) {
		case string:
			result[key] = vv
		default:
			log.Println("virtlib.readMetaDataFile.range: wrong type ", vv)
		}
	}

	return &result, nil
} // readJSONmetaDataFile()

/* _EoF_ */
