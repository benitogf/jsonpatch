package jsonpatch

import (
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

func BenchmarkCreatePatchSimpleObject(b *testing.B) {
	a := []byte(`{"a":100,"b":200,"c":"hello"}`)
	bb := []byte(`{"a":100,"b":200,"c":"goodbye"}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkCreatePatchNestedObject(b *testing.B) {
	a := []byte(`{"a":{"b":{"c":{"d":1,"e":"hello"}}}}`)
	bb := []byte(`{"a":{"b":{"c":{"d":2,"e":"world"}}}}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkCreatePatchLargeObject(b *testing.B) {
	obj := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		obj[fmt.Sprintf("key%d", i)] = i
	}
	a, _ := json.Marshal(obj)

	obj["key50"] = "changed"
	obj["key99"] = "modified"
	bb, _ := json.Marshal(obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkCreatePatchSmallArray(b *testing.B) {
	a := []byte(`[1,2,3,4,5]`)
	bb := []byte(`[1,2,4,5,6]`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkCreatePatchMediumArray(b *testing.B) {
	arr1 := make([]int, 50)
	arr2 := make([]int, 50)
	for i := 0; i < 50; i++ {
		arr1[i] = i
		arr2[i] = i
	}
	arr2[25] = 999
	arr2[49] = 888

	a, _ := json.Marshal(arr1)
	bb, _ := json.Marshal(arr2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkCreatePatchArrayWithObjects(b *testing.B) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	arr1 := make([]Item, 20)
	arr2 := make([]Item, 20)
	for i := 0; i < 20; i++ {
		arr1[i] = Item{ID: i, Name: fmt.Sprintf("item%d", i)}
		arr2[i] = Item{ID: i, Name: fmt.Sprintf("item%d", i)}
	}
	arr2[10].Name = "modified"
	arr2[15].Name = "changed"

	a, _ := json.Marshal(arr1)
	bb, _ := json.Marshal(arr2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePatch(a, bb)
	}
}

func BenchmarkApplyPatchSimple(b *testing.B) {
	doc := []byte(`{"a":100,"b":200,"c":"hello"}`)
	patchJSON := []byte(`[{"op":"replace","path":"/c","value":"goodbye"}]`)
	patch, _ := DecodePatch(patchJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch.Apply(doc)
	}
}

func BenchmarkApplyPatchMultipleOps(b *testing.B) {
	doc := []byte(`{"a":1,"b":2,"c":3,"d":4,"e":5}`)
	patchJSON := []byte(`[
		{"op":"replace","path":"/a","value":10},
		{"op":"replace","path":"/b","value":20},
		{"op":"add","path":"/f","value":6},
		{"op":"remove","path":"/c"}
	]`)
	patch, _ := DecodePatch(patchJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch.Apply(doc)
	}
}

func BenchmarkApplyPatchArray(b *testing.B) {
	doc := []byte(`{"items":[1,2,3,4,5]}`)
	patchJSON := []byte(`[
		{"op":"add","path":"/items/0","value":0},
		{"op":"remove","path":"/items/3"},
		{"op":"replace","path":"/items/1","value":99}
	]`)
	patch, _ := DecodePatch(patchJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch.Apply(doc)
	}
}

func BenchmarkApplyPatchNested(b *testing.B) {
	doc := []byte(`{"a":{"b":{"c":{"d":1}}}}`)
	patchJSON := []byte(`[{"op":"replace","path":"/a/b/c/d","value":2}]`)
	patch, _ := DecodePatch(patchJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch.Apply(doc)
	}
}

func BenchmarkDecodePatch(b *testing.B) {
	patchJSON := []byte(`[
		{"op":"add","path":"/a","value":1},
		{"op":"remove","path":"/b"},
		{"op":"replace","path":"/c","value":"hello"},
		{"op":"move","from":"/d","path":"/e"},
		{"op":"copy","from":"/f","path":"/g"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodePatch(patchJSON)
	}
}

func BenchmarkEqual(b *testing.B) {
	a := []byte(`{"a":1,"b":{"c":2,"d":[1,2,3]},"e":"hello"}`)
	bb := []byte(`{"a":1,"b":{"c":2,"d":[1,2,3]},"e":"hello"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Equal(a, bb)
	}
}

func BenchmarkEqualDifferent(b *testing.B) {
	a := []byte(`{"a":1,"b":{"c":2,"d":[1,2,3]},"e":"hello"}`)
	bb := []byte(`{"a":1,"b":{"c":2,"d":[1,2,4]},"e":"hello"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Equal(a, bb)
	}
}

func BenchmarkMatchesValueString(b *testing.B) {
	a := "hello world"
	bb := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchesValue(a, bb)
	}
}

func BenchmarkMatchesValueMap(b *testing.B) {
	a := map[string]interface{}{"a": 1, "b": "hello", "c": true}
	bb := map[string]interface{}{"a": 1, "b": "hello", "c": true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchesValue(a, bb)
	}
}

func BenchmarkMakePath(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		makePath("/a/b", 123)
	}
}

func BenchmarkMakePathString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		makePath("/a/b", "key")
	}
}

func BenchmarkHashValue(b *testing.B) {
	v := map[string]interface{}{"a": 1, "b": "hello"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashValue(v)
	}
}
