package jsonpatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var errBadJSONDoc = fmt.Errorf("Invalid JSON Document")
var errBadMergeTypes = fmt.Errorf("Mismatched JSON Documents")

// Operation operation struct
type Operation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value,omitempty"`
}

// resemblesJSONArray indicates whether the byte-slice "appears" to be
// a JSON array or not.
// False-positives are possible, as this function does not check the internal
// structure of the array. It only checks that the outer syntax is present and
// correct.
func resemblesJSONArray(input []byte) bool {
	input = bytes.TrimSpace(input)

	hasPrefix := bytes.HasPrefix(input, []byte("["))
	hasSuffix := bytes.HasSuffix(input, []byte("]"))

	return hasPrefix && hasSuffix
}

// JSON returns a patch operation Json representation
func (j *Operation) JSON() string {
	b, _ := json.Marshal(j)
	return string(b)
}

// MarshalJSON for patch operations
func (j *Operation) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{")
	b.WriteString(fmt.Sprintf(`"op":"%s"`, j.Operation))
	b.WriteString(fmt.Sprintf(`,"path":"%s"`, j.Path))
	// Consider omitting Value for non-nullable operations.
	if j.Value != nil || j.Operation == "replace" || j.Operation == "add" {
		v, err := json.Marshal(j.Value)
		if err != nil {
			return nil, err
		}
		b.WriteString(`,"value":`)
		b.Write(v)
	}
	b.WriteString("}")
	return b.Bytes(), nil
}

// ByPath array of patch operation structs
type ByPath []Operation

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPath) Less(i, j int) bool { return a[i].Path < a[j].Path }

// NewPatch creates a patch operation struct
func NewPatch(operation, path string, value interface{}) Operation {
	return Operation{Operation: operation, Path: path, Value: value}
}

// CreatePatch creates a patch as specified in http://jsonpatch.com/
//
// 'a' is original, 'b' is the modified document. Both are to be given as json encoded content.
// The function will return an array of Operations
//
// An error will be returned if any of the two documents are invalid.
func CreatePatch(a, b []byte) ([]Operation, error) {
	if bytes.Equal(a, b) {
		return []Operation{}, nil
	}
	originalResemblesArray := resemblesJSONArray(a)
	modifiedResemblesArray := resemblesJSONArray(b)
	// Do both byte-slices seem like JSON arrays?
	if originalResemblesArray && modifiedResemblesArray {
		original := []json.RawMessage{}
		modified := []json.RawMessage{}

		d := json.NewDecoder(bytes.NewReader(a))
		d.UseNumber()
		err := d.Decode(&original)
		if err != nil {
			return nil, err
		}
		db := json.NewDecoder(bytes.NewReader(b))
		db.UseNumber()
		err = db.Decode(&modified)
		if err != nil {
			return nil, err
		}

		patch := []Operation{}
		path := ""

		keysModified := map[int]bool{}
		keysOriginal := map[int]bool{}
		for k := range original {
			keysOriginal[k] = true
		}

		if len(modified) == len(original) {
			// very specific case of a moving window of collections in ascending order
			diffCount := 0
			length := len(modified) - 1
			for key := range modified {
				// first element of the original cant be found in the modified
				if key < length && string(original[0]) == string(modified[key]) {
					diffCount++
					break
				}
				// last element of the modified cant be found in the original
				if key > 0 && string(modified[length]) == string(original[key]) {
					diffCount++
					break
				}
				// other then first original and last modified all elements are the same
				if key < length && string(original[key+1]) != string(modified[key]) {
					diffCount++
					break
				}
			}

			if diffCount == 0 {
				pFirst := makePath(path, 0)
				pLast := makePath(path, length)
				patch = append([]Operation{NewPatch("add", pLast, modified[0])}, patch...)
				patch = append([]Operation{NewPatch("remove", pFirst, nil)}, patch...)
				return patch, nil
			}
		}

		for key, bv := range modified {
			keysModified[key] = true
			p := makePath(path, key)
			_, found := keysOriginal[key]
			// value was added
			if !found {
				patch = append([]Operation{NewPatch("add", p, bv)}, patch...)
				continue
			}
			av := original[key]
			// If types have changed, replace completely
			if reflect.TypeOf(av) != reflect.TypeOf(bv) {
				patch = append([]Operation{NewPatch("replace", p, bv)}, patch...)
				continue
			}
			// Types are the same, compare values
			patch, err = diffObjects(av, bv, "/"+strconv.Itoa(key)+"/", patch)
			if err != nil {
				return nil, err
			}
		}
		// Now add all deleted values as nil
		for key := range original {
			_, found := keysModified[key]
			if !found {
				p := makePath(path, key)
				patch = append([]Operation{NewPatch("remove", p, nil)}, patch...)
			}
		}

		return patch, nil
	}

	// Are both byte-slices are not arrays? Then they are likely JSON objects...
	if !originalResemblesArray && !modifiedResemblesArray {
		return diffObjects(a, b, "", []Operation{})
	}

	// None of the above? Then return an error because of mismatched types.
	return nil, errBadMergeTypes
}

func diffObjects(a, b []byte, key string, patch []Operation) ([]Operation, error) {
	aI := map[string]interface{}{}
	bI := map[string]interface{}{}
	d := json.NewDecoder(bytes.NewReader(a))
	d.UseNumber()
	err := d.Decode(&aI)
	if err != nil {
		return nil, err
	}
	db := json.NewDecoder(bytes.NewReader(b))
	db.UseNumber()
	err = db.Decode(&bI)
	if err != nil {
		return nil, err
	}

	return diff(aI, bI, key, patch)
}

// Returns true if the values matches (must be json types)
// The types of the values must match, otherwise it will always return false
// If two map[string]interface{} are given, all elements must match.
func matchesValue(av, bv interface{}) bool {
	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
		return false
	}
	switch at := av.(type) {
	case string:
		bt := bv.(string)
		if bt == at {
			return true
		}
	case json.Number:
		bt := bv.(json.Number)
		if bt == at {
			return true
		}
	case bool:
		bt := bv.(bool)
		if bt == at {
			return true
		}
	case map[string]interface{}:
		bt := bv.(map[string]interface{})
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	case []interface{}:
		bt := bv.([]interface{})
		if len(bt) != len(at) {
			return false
		}
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	}
	return false
}

// From http://tools.ietf.org/html/rfc6901#section-4 :
//
// Evaluation of each reference token begins by decoding any escaped
// character sequence.  This is performed by first transforming any
// occurrence of the sequence '~1' to '/', and then transforming any
// occurrence of the sequence '~0' to '~'.
//   TODO decode support:
//   var rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")

var rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")

func makePath(path string, newPart interface{}) string {
	key := rfc6901Encoder.Replace(fmt.Sprintf("%v", newPart))
	if path == "" {
		return "/" + key
	}
	if strings.HasSuffix(path, "/") {
		return path + key
	}
	return path + "/" + key
}

// diff returns the (recursive) difference between a and b as an array of Operations.
func diff(a, b map[string]interface{}, path string, patch []Operation) ([]Operation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			patch = append([]Operation{NewPatch("add", p, bv)}, patch...)
			continue
		}
		// If types have changed, replace completely
		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
			patch = append([]Operation{NewPatch("replace", p, bv)}, patch...)
			continue
		}
		// Types are the same, compare values
		var err error
		patch, err = handleValues(av, bv, p, patch)
		if err != nil {
			return nil, err
		}
	}
	// Now add all deleted values as nil
	for key := range a {
		_, found := b[key]
		if !found {
			p := makePath(path, key)

			patch = append([]Operation{NewPatch("remove", p, nil)}, patch...)
		}
	}
	return patch, nil
}

func handleValues(av, bv interface{}, p string, patch []Operation) ([]Operation, error) {
	var err error
	switch at := av.(type) {
	case map[string]interface{}:
		bt := bv.(map[string]interface{})
		patch, err = diff(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case string, json.Number, bool:
		if !matchesValue(av, bv) {
			patch = append([]Operation{NewPatch("replace", p, bv)}, patch...)
		}
	case []interface{}:
		bt, ok := bv.([]interface{})
		if !ok {
			// array replaced by non-array
			patch = append([]Operation{NewPatch("replace", p, bv)}, patch...)
		} else if len(at) != len(bt) {
			// arrays are not the same length
			patch = append(patch, compareArray(at, bt, p)...)

		} else {
			for i := range bt {
				patch, err = handleValues(at[i], bt[i], makePath(p, i), patch)
				if err != nil {
					return nil, err
				}
			}
		}
	case nil:
		switch bv.(type) {
		case nil:
			// Both nil, fine.
		default:
			patch = append([]Operation{NewPatch("add", p, bv)}, patch...)
		}
	default:
		panic(fmt.Sprintf("Unknown type:%T ", av))
	}
	return patch, nil
}

func compareArray(av, bv []interface{}, p string) []Operation {
	retval := []Operation{}
	//	var err error
	for i, v := range av {
		found := false
		for _, v2 := range bv {
			if reflect.DeepEqual(v, v2) {
				found = true
				break
			}
		}
		if !found {
			retval = append([]Operation{NewPatch("remove", makePath(p, i), nil)}, retval...)
		}
	}

	for i, v := range bv {
		found := false
		for _, v2 := range av {
			if reflect.DeepEqual(v, v2) {
				found = true
				break
			}
		}
		if !found {
			retval = append([]Operation{NewPatch("add", makePath(p, i), v)}, retval...)
		}
	}

	return retval
}
