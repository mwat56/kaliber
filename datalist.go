/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

/*
 * This file provides an object whose properties are to be inserted
 * into templates.
 */

type (
	// TemplateData is a list of values to be injected into a template.
	TemplateData map[string]interface{}
)

// Set inserts `aValue` identified by `aKey` to the list.
//
// If there's already a list entry with `aKey` its current value
// gets replaced by `aValue`.
//
// `aKey` is the values's identifier (as used as placeholder in the template).
//
// `aValue` contains the data entry's value.
func (dl *TemplateData) Set(aKey string, aValue interface{}) *TemplateData {
	(*dl)[aKey] = aValue

	return dl
} // Set()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// NewTemplateData returns a new (empty) TDataList instance.
func NewTemplateData() *TemplateData {
	result := make(TemplateData, 32)

	return &result
} // NewTemplateData()

/* _EoF_ */
