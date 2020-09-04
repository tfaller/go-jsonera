// Package jsonera can find and track changes of a json document
package jsonera

import (
	"github.com/tfaller/go-jsonvisitor"
)

const (
	// eraPrefixProp is the property name prefix in an
	// era document to signal the current era number
	eraPrefixProp = "."

	// eraPrefixObj is the property name prefix in an
	// era document to signal sub-properties
	eraPrefixObj = "_"
)

// EraDocument is a complete era representation
type EraDocument struct {
	// Doc is the current known version of the document
	Doc interface{} `json:"doc"`

	// Era represents the document, but instead of the actual
	// properties, it holds the era numbers to indicate when a
	// specific property was changed
	Era interface{} `json:"era"`

	// DocEra is the current era number
	DocEra uint32 `json:"docEra"`
}

// ChangeMode represents the kind of change
type ChangeMode int

const (
	// ChangeEqual represents unchanged property
	ChangeEqual ChangeMode = iota

	// ChangeNew represents that the property is new
	ChangeNew

	// ChangeUpdate represents that the property was updated
	ChangeUpdate

	// ChangeDelete represents that the property got deleted
	ChangeDelete
)

var changeModeNames = []string{"ChangeEqual", "ChangeNew", "ChangeUpdate", "ChangeDelete"}

func (c ChangeMode) String() string {
	return changeModeNames[c]
}

// Change is a changed property
type Change struct {
	// Path is the complete path of the property
	Path []string

	// Era is the previous era number
	Era uint64

	// Mode states which kind of change it was
	Mode ChangeMode
}

type valType int

const (
	valTypeBasic valType = iota
	valTypeObject
	valTypeArray
)

// jsonObj is a unmarshalled json object
// Yes, type alias, instead of type declaration!
type jsonObj = map[string]interface{}

// UpdateDoc updates the document and the era information
// The result of this function is all found changes
func (e *EraDocument) UpdateDoc(newDoc interface{}) []Change {
	newEraTree, changes := buildNewEra(newDoc, e.Doc, e.Era.(jsonObj), e.DocEra+1)

	if len(changes) > 0 {
		// the doc actually changed
		e.Doc = newDoc
		e.Era = newEraTree
		e.DocEra++
	}

	return changes
}

// NewEraDocument creates and initializes a new EraDocument
func NewEraDocument(doc interface{}) *EraDocument {
	e := &EraDocument{Doc: jsonvisitor.Undefined, Era: jsonObj{}, DocEra: 0}
	e.UpdateDoc(doc)
	return e
}

// buildNewEra find all changes between an old and a new document.
// It also creates a new era document.
func buildNewEra(newDoc, oldDoc interface{}, oldEraObj jsonObj, newEra uint32) (jsonObj, []Change) {
	changes, newEraObj := []Change{}, jsonObj{}

	// pack the docs into a document -> makes tracking of
	// schema changes of the document simple
	newDoc, oldDoc = jsonObj{"": newDoc}, jsonObj{"": oldDoc}

	buildSubEra(nil,
		newDoc, oldDoc,
		newEraObj, oldEraObj,
		newEra, &changes,
	)

	return newEraObj, changes
}

// buildSubEra builds the new era document and finds changes of a part of a document
// The function calls itself recursively if a property is an array or object
func buildSubEra(path []string, newVal, oldVal interface{}, newEraObj, oldEraObj jsonObj, newEra uint32, changes *[]Change) bool {
	startPathLen := len(path)
	changedSchema := false

	// walk over all properties of this array or object
	jsonvisitor.PairVisitWithPath(path, newVal, oldVal, func(path []string, newVal, oldVal interface{}) bool {
		pathLen := len(path)
		if pathLen == startPathLen {
			// ignore the object/array itself, but visit the children
			return true
		}

		change := ChangeEqual
		newType, oldType := getValType(newVal), getValType(oldVal)

		if newVal == jsonvisitor.Undefined {
			// if no new value exists the prop got delete
			change = ChangeDelete
		} else if oldVal == jsonvisitor.Undefined {
			// if no old value exists the prop is new
			change = ChangeNew
		} else if newType != oldType {
			// we changed between basic, object or array
			// so this is for sure a change
			change = ChangeUpdate
		} else if newType == valTypeBasic {
			change = cmpBasicJSONVal(newVal, oldVal)
		}

		// the array/object changed itself,
		// if it was a new or deleted property
		changedSchema = changedSchema ||
			change == ChangeNew ||
			change == ChangeDelete

		name := path[pathLen-1]
		eraName := eraPrefixProp + name
		eraObjName := eraPrefixObj + name

		newSubEraObj := jsonObj{}
		if newType != valTypeBasic {
			// only add era of sub elements
			// if we are an type that actually holds sub properties
			newEraObj[eraObjName] = newSubEraObj
		}

		// visit children
		oldSubEraObj, _ := oldEraObj[eraObjName].(jsonObj)
		subSchemaChanged := buildSubEra(path,
			newVal, oldVal,
			newSubEraObj, oldSubEraObj,
			newEra, changes,
		)

		if subSchemaChanged && change == ChangeEqual {
			// a property got deleted or added
			change = ChangeUpdate
		}

		oldEraVal, oldEraValExist := oldEraObj[eraName].(float64)
		if !oldEraValExist {
			oldEraVal = float64(newEra)
		}
		newEraVal := oldEraVal

		if change != ChangeEqual {
			newEraVal = float64(newEra)

			*changes = append(*changes, Change{
				// copy to prevent side effects
				Path: copyStringSlice(path[1:]),
				Era:  uint64(oldEraVal),
				Mode: change,
			})
		}

		if change != ChangeDelete {
			// Set era number for this property.
			// Don't set era if the prop was deleted.
			newEraObj[eraName] = float64(newEraVal)
		}

		// don't visit recursively deeper
		return false
	})

	return changedSchema
}

func copyStringSlice(slice []string) []string {
	dst := make([]string, len(slice))
	copy(dst, slice)
	return dst
}

// normalizeNumber normalisizes any number to float64.
// This is because JSON can only have that type of number.
// If the supplied value is not a number, the result is unchanged
// An int64 or uint64 could overflow the safe values of float64
// but this is not checked here!
func normalizeNumber(num interface{}) interface{} {

	switch num.(type) {
	case int:
		return float64(num.(int))
	case int8:
		return float64(num.(int8))
	case int16:
		return float64(num.(int16))
	case int32:
		return float64(num.(int32))
	case int64:
		return float64(num.(int64))
	case uint:
		return float64(num.(uint))
	case uint8:
		return float64(num.(uint8))
	case uint16:
		return float64(num.(uint16))
	case uint32:
		return float64(num.(uint32))
	case uint64:
		return float64(num.(uint64))
	case float32:
		return float64(num.(float32))
	}

	// already float64 or not a number
	return num
}

// cmpBasicJSONVal compares basic JSON values ... string, number, bool, nil.
// Only basic values are allowed to be passed as func arguments!
// Values like arrays, objects or pointer produce undefined behavior and might panic!
func cmpBasicJSONVal(a, b interface{}) ChangeMode {
	// normalize potential numbers for comparisions
	a, b = normalizeNumber(a), normalizeNumber(b)

	// if the values are string|bool|number|nil
	// the basic "==" operator works as expected
	if a == b {
		return ChangeEqual
	}
	return ChangeUpdate
}

func getValType(val interface{}) valType {
	switch val.(type) {
	case map[string]interface{}:
		return valTypeObject
	case []interface{}:
		return valTypeArray
	}
	return valTypeBasic
}
