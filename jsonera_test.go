package jsonera

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBuildNewEra(t *testing.T) {
	testCases := []struct {
		newDoc  string
		oldDoc  string
		newEra  string
		oldEra  string
		changes []Change
	}{
		// no changes
		{
			newDoc:  `{}`,
			oldDoc:  `{}`,
			newEra:  `{".": 2, "_": {}}`,
			oldEra:  `{".": 2, "_": {}}`,
			changes: []Change{},
		},
		// simple value change
		{
			newDoc: `{"a": 1}`,
			oldDoc: `{"a": 0}`,
			newEra: `{".": 1, "_": {".a": 1}}`,
			oldEra: `{}`,
			changes: []Change{
				{Path: []string{"a"}, Era: 1, Mode: ChangeUpdate},
			},
		},
		// change from array to object
		{
			newDoc: `{"a": {}}`,
			oldDoc: `{"a": []}`,
			newEra: `{".": 1, "_": {".a": 1, "_a": {}}}`,
			oldEra: `{}`,
			changes: []Change{
				{Path: []string{"a"}, Era: 1, Mode: ChangeUpdate},
			},
		},
		// simple new value
		{
			newDoc: `{"a": 0}`,
			oldDoc: `{}`,
			newEra: `{".": 1, "_": {".a": 1}}`,
			oldEra: `{}`,
			changes: []Change{
				{Path: []string{"a"}, Era: 1, Mode: ChangeNew},
				{Path: []string{}, Era: 1, Mode: ChangeUpdate},
			},
		},
		// deletion of top-level property
		{
			newDoc: `{}`,
			oldDoc: `{"a": [0]}`,
			newEra: `{".": 1, "_": {}}`,
			oldEra: `{}`,
			changes: []Change{
				{Path: []string{"a", "0"}, Era: 1, Mode: ChangeDelete},
				{Path: []string{"a"}, Era: 1, Mode: ChangeDelete},
				{Path: []string{}, Era: 1, Mode: ChangeUpdate},
			},
		},
		// deletion of sub-level property
		{
			newDoc: `{"a": []}`,
			oldDoc: `{"a": [0]}`,
			newEra: `{".": 2, "_": {".a": 1, "_a": {}}}`,
			oldEra: `{".": 2, "_": {".a": 2, "_a": {".0": 2}}}`,
			changes: []Change{
				{Path: []string{"a", "0"}, Era: 2, Mode: ChangeDelete},
				{Path: []string{"a"}, Era: 2, Mode: ChangeUpdate},
			},
		},
	}

	for i, test := range testCases {
		var valA, valB, oldEra, newEra interface{}

		unmarshalJSONString(test.newDoc, &valA, t)
		unmarshalJSONString(test.oldDoc, &valB, t)
		unmarshalJSONString(test.newEra, &newEra, t)
		unmarshalJSONString(test.oldEra, &oldEra, t)

		computedNewEra, computedChanges := buildNewEra(valA, valB, oldEra.(map[string]interface{}), 1)

		if !reflect.DeepEqual(newEra, computedNewEra) {
			t.Errorf("Test %v has wrong era: %v <-> %v", i, newEra, computedNewEra)
		}

		if !reflect.DeepEqual(test.changes, computedChanges) {
			t.Errorf("Test %v has wrong changes: %v <-> %v", i, test.changes, computedChanges)
		}
	}
}

func TestEraDocument(t *testing.T) {
	var doc, docUpdate interface{}
	rawDoc := `{}`
	rawDocUpdate := `{"a": 1}`
	unmarshalJSONString(rawDoc, &doc, t)
	unmarshalJSONString(rawDocUpdate, &docUpdate, t)

	eraDoc := NewEraDocument(doc)
	if eraDoc.DocEra != 1 {
		t.Error("invalid initial era")
	}

	// update with same data
	changes := eraDoc.UpdateDoc(doc)
	if len(changes) != 0 {
		t.Error("Found changed but there where actually no changes")
	}
	if eraDoc.DocEra != 1 {
		t.Error("DocEra change even no changes where found")
	}

	// update with changed data
	changes = eraDoc.UpdateDoc(docUpdate)
	if len(changes) != 2 {
		t.Error("Wrong number of found changes")
	}
	if eraDoc.DocEra != 2 {
		t.Error("DocEra has not expected value")
	}
}

func TestNormalizeNumber(t *testing.T) {
	cases := []interface{}{
		// floats
		float32(1), float64(1),
		// signed ints
		int(1), int8(1), int16(1), int32(1), int64(1),
		// unsigned ints
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1),
	}

	for i, c := range cases {
		if normalizeNumber(c) != 1.0 {
			t.Errorf("Test %v has wrong number", i)
		}
	}
}

func TestChangeName(t *testing.T) {
	tests := []struct {
		mode ChangeMode
		name string
	}{
		{ChangeEqual, "ChangeEqual"},
		{ChangeNew, "ChangeNew"},
		{ChangeUpdate, "ChangeUpdate"},
		{ChangeDelete, "ChangeDelete"},
	}

	for i, test := range tests {
		if test.mode.String() != test.name {
			t.Errorf("Test %v failed", i)
		}
	}
}

func TestCmpBasicJSONVal(t *testing.T) {
	if cmpBasicJSONVal(nil, nil) != ChangeEqual {
		t.Error("nil must be equal to nil")
	}
}

func unmarshalJSONString(str string, target interface{}, t *testing.T) {
	err := json.Unmarshal([]byte(str), target)
	if err != nil {
		t.Fatal(err)
	}
}
