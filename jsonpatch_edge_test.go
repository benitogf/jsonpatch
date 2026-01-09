package jsonpatch

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePatchEmptyObjects(t *testing.T) {
	patch, err := CreatePatch([]byte(`{}`), []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 0, len(patch))
}

func TestCreatePatchEmptyArrays(t *testing.T) {
	patch, err := CreatePatch([]byte(`[]`), []byte(`[]`))
	require.NoError(t, err)
	assert.Equal(t, 0, len(patch))
}

func TestCreatePatchEmptyToNonEmpty(t *testing.T) {
	patch, err := CreatePatch([]byte(`{}`), []byte(`{"a":1}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "add", patch[0].Operation)
	assert.Equal(t, "/a", patch[0].Path)
}

func TestCreatePatchNonEmptyToEmpty(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":1}`), []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "remove", patch[0].Operation)
	assert.Equal(t, "/a", patch[0].Path)
}

func TestCreatePatchNullValue(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":1}`), []byte(`{"a":null}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "replace", patch[0].Operation)
	assert.Nil(t, patch[0].Value)
}

func TestCreatePatchNullToValue(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":null}`), []byte(`{"a":1}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "replace", patch[0].Operation)
}

func TestCreatePatchNestedNull(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":{"b":null}}`), []byte(`{"a":{"b":1}}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "replace", patch[0].Operation)
	assert.Equal(t, "/a/b", patch[0].Path)
}

func TestCreatePatchArrayDuplicates(t *testing.T) {
	// Top-level arrays with primitives require object elements for diffObjects
	patch, err := CreatePatch([]byte(`[{"a":1},{"a":1},{"a":1}]`), []byte(`[{"a":1},{"a":1},{"a":2}]`))
	require.NoError(t, err)
	assert.True(t, len(patch) > 0)
}

func TestCreatePatchArrayWithNulls(t *testing.T) {
	// Top-level arrays need object elements
	patch, err := CreatePatch([]byte(`[{"a":null},{"a":1},{"a":2}]`), []byte(`[{"a":null},{"a":1},{"a":3}]`))
	require.NoError(t, err)
	assert.True(t, len(patch) > 0)
}

func TestCreatePatchDeepNesting(t *testing.T) {
	a := `{"a":{"b":{"c":{"d":{"e":{"f":1}}}}}}`
	b := `{"a":{"b":{"c":{"d":{"e":{"f":2}}}}}}`
	patch, err := CreatePatch([]byte(a), []byte(b))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "/a/b/c/d/e/f", patch[0].Path)
}

func TestCreatePatchSpecialCharactersInKeys(t *testing.T) {
	a := `{"a/b":1,"~c":2}`
	b := `{"a/b":2,"~c":3}`
	patch, err := CreatePatch([]byte(a), []byte(b))
	require.NoError(t, err)
	assert.Equal(t, 2, len(patch))
}

func TestCreatePatchBooleanValues(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":true}`), []byte(`{"a":false}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
	assert.Equal(t, "replace", patch[0].Operation)
	assert.Equal(t, false, patch[0].Value)
}

func TestCreatePatchNumberTypes(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":1}`), []byte(`{"a":1.5}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
}

func TestCreatePatchLargeNumbers(t *testing.T) {
	patch, err := CreatePatch([]byte(`{"a":9999999999999999}`), []byte(`{"a":9999999999999998}`))
	require.NoError(t, err)
	assert.Equal(t, 1, len(patch))
}

func TestCreatePatchMixedArrayTypes(t *testing.T) {
	// Test with object wrapper
	a := `{"arr":[1,"hello",true,null,{"a":1}]}`
	b := `{"arr":[1,"world",false,null,{"a":2}]}`
	patch, err := CreatePatch([]byte(a), []byte(b))
	require.NoError(t, err)
	assert.True(t, len(patch) > 0)
}

func TestCreatePatchTypeMismatch(t *testing.T) {
	_, err := CreatePatch([]byte(`{}`), []byte(`[]`))
	assert.Error(t, err)
}

func TestCreatePatchInvalidJSON(t *testing.T) {
	_, err := CreatePatch([]byte(`{invalid}`), []byte(`{}`))
	assert.Error(t, err)
}

func TestApplyPatchEmptyPatch(t *testing.T) {
	doc := []byte(`{"a":1}`)
	patch, _ := DecodePatch([]byte(`[]`))
	result, err := patch.Apply(doc)
	require.NoError(t, err)
	assert.True(t, Equal(doc, result))
}

func TestApplyPatchToEmptyObject(t *testing.T) {
	doc := []byte(`{}`)
	patch, _ := DecodePatch([]byte(`[{"op":"add","path":"/a","value":1}]`))
	result, err := patch.Apply(doc)
	require.NoError(t, err)
	assert.True(t, Equal([]byte(`{"a":1}`), result))
}

func TestApplyPatchToEmptyArray(t *testing.T) {
	doc := []byte(`[]`)
	patch, _ := DecodePatch([]byte(`[{"op":"add","path":"/0","value":1}]`))
	result, err := patch.Apply(doc)
	require.NoError(t, err)
	assert.True(t, Equal([]byte(`[1]`), result))
}

func TestApplyPatchNullValue(t *testing.T) {
	doc := []byte(`{"a":1}`)
	patch, _ := DecodePatch([]byte(`[{"op":"replace","path":"/a","value":null}]`))
	result, err := patch.Apply(doc)
	require.NoError(t, err)
	assert.True(t, Equal([]byte(`{"a":null}`), result))
}

func TestEqualIdentical(t *testing.T) {
	a := []byte(`{"a":1,"b":[1,2,3]}`)
	assert.True(t, Equal(a, a))
}

func TestEqualDifferentOrder(t *testing.T) {
	a := []byte(`{"a":1,"b":2}`)
	b := []byte(`{"b":2,"a":1}`)
	assert.True(t, Equal(a, b))
}

func TestEqualDifferentValues(t *testing.T) {
	a := []byte(`{"a":1}`)
	b := []byte(`{"a":2}`)
	assert.False(t, Equal(a, b))
}

func TestEqualDifferentKeys(t *testing.T) {
	a := []byte(`{"a":1}`)
	b := []byte(`{"b":1}`)
	assert.False(t, Equal(a, b))
}

func TestEqualExtraKey(t *testing.T) {
	a := []byte(`{"a":1}`)
	b := []byte(`{"a":1,"b":2}`)
	assert.False(t, Equal(a, b))
}

func TestEqualArrays(t *testing.T) {
	a := []byte(`[1,2,3]`)
	b := []byte(`[1,2,3]`)
	assert.True(t, Equal(a, b))
}

func TestEqualArraysDifferentOrder(t *testing.T) {
	a := []byte(`[1,2,3]`)
	b := []byte(`[3,2,1]`)
	assert.False(t, Equal(a, b))
}

func TestMatchesValueNil(t *testing.T) {
	assert.True(t, matchesValue(nil, nil))
	assert.False(t, matchesValue(nil, 1))
	assert.False(t, matchesValue(1, nil))
}

func TestMatchesValueDifferentTypes(t *testing.T) {
	assert.False(t, matchesValue("1", 1))
	assert.False(t, matchesValue(true, "true"))
	assert.False(t, matchesValue([]interface{}{1}, map[string]interface{}{"a": 1}))
}

func TestSameType(t *testing.T) {
	assert.True(t, sameType("a", "b"))
	assert.True(t, sameType(true, false))
	assert.True(t, sameType(nil, nil))
	assert.False(t, sameType("a", 1))
	assert.False(t, sameType(nil, "a"))
	// Note: sameType uses json.Number, not int
	assert.True(t, sameType(json.Number("1"), json.Number("2")))
}

func TestSortDescending(t *testing.T) {
	s := []int{1, 5, 3, 2, 4}
	sortDescending(s)
	assert.Equal(t, []int{5, 4, 3, 2, 1}, s)
}

func TestSortDescendingEmpty(t *testing.T) {
	s := []int{}
	sortDescending(s)
	assert.Equal(t, []int{}, s)
}

func TestSortDescendingSingle(t *testing.T) {
	s := []int{1}
	sortDescending(s)
	assert.Equal(t, []int{1}, s)
}

func TestSortAscending(t *testing.T) {
	s := []int{5, 1, 3, 2, 4}
	sortAscending(s)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, s)
}

func TestHashValue(t *testing.T) {
	h1 := hashValue(map[string]interface{}{"a": 1})
	h2 := hashValue(map[string]interface{}{"a": 1})
	assert.Equal(t, h1, h2)

	h3 := hashValue(map[string]interface{}{"a": 2})
	assert.NotEqual(t, h1, h3)
}

func TestRoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		original string
		modified string
	}{
		{"simple_replace", `{"a":1}`, `{"a":2}`},
		{"add_key", `{"a":1}`, `{"a":1,"b":2}`},
		{"remove_key", `{"a":1,"b":2}`, `{"a":1}`},
		{"nested_change", `{"a":{"b":1}}`, `{"a":{"b":2}}`},
		{"array_change", `{"a":[1,2,3]}`, `{"a":[1,2,4]}`},
		{"array_add", `{"a":[1,2]}`, `{"a":[1,2,3]}`},
		{"array_remove", `{"a":[1,2,3]}`, `{"a":[1,2]}`},
		{"complex", `{"a":1,"b":{"c":[1,2,3]}}`, `{"a":2,"b":{"c":[1,3],"d":4}}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patch, err := CreatePatch([]byte(tc.original), []byte(tc.modified))
			require.NoError(t, err)

			patchBytes, err := MarshalPatch(patch)
			require.NoError(t, err)

			decoded, err := DecodePatch(patchBytes)
			require.NoError(t, err)

			result, err := decoded.Apply([]byte(tc.original))
			require.NoError(t, err)

			assert.True(t, Equal([]byte(tc.modified), result),
				"Expected %s but got %s", tc.modified, string(result))
		})
	}
}

func MarshalPatch(ops []Operation) ([]byte, error) {
	return json.Marshal(ops)
}
