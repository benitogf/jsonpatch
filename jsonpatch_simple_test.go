package jsonpatch

import (
	"sort"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleA = `{"a":100, "b":200, "c":"hello"}`
var simpleB = `{"a":100, "b":200, "c":"goodbye"}`
var simpleC = `{"a":100, "b":100, "c":"hello"}`
var simpleD = `{"a":100, "b":200, "c":"hello", "d":"foo"}`
var simpleE = `{"a":100, "b":200}`
var simplef = `{"a":100, "b":100, "d":"foo"}`
var simpleG = `{"a":100, "b":null, "d":"foo"}`
var empty = `{}`

var collection = `[{"created":1564944548180294000,"updated":0,"index":"15b7ccc66f7878a8","data":"eyJuYW1lIjoibmFtZTEifQ=="},{"created":1564944548172292600,"updated":0,"index":"15b7ccc66efe6194","data":"eyJuYW1lIjoibmFtZTAifQ=="}]`
var emptyCollection = `[]`

var collectionWindowAscBefore = `[{
	"test":"1"
},
{
	"test":"2"
},
{
	"test":"3"
}]`

var collectionWindowAscAfter = `[{
	"test":"2"
},
{
	"test":"3"
},
{
	"test":"4"
}]`

var collectionWindowDscBefore = `[{
	"test":"3"
},
{
	"test":"2"
},
{
	"test":"1"
}]`

var collectionWindowDscAfter = `[{
	"test":"4"
},
{
	"test":"3"
},
{
	"test":"2"
}]`

var collectionOne = `[{
	"test":"1"
}]`

var collectionTwo = `[{
	"test":"2"
},
{
	"test":"1"
}]`

func TestMultipleRemove(t *testing.T) {
	patch, e := CreatePatch([]byte(collection), []byte(emptyCollection))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 2, "the patch should be the same lenght as the collection")
	assert.Equal(t, patch[0].Path, "/1", "the patch should have descending order by path")
}

func TestCollectionWindowAscMove(t *testing.T) {
	patch, e := CreatePatch([]byte(collectionWindowAscBefore), []byte(collectionWindowAscAfter))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should have one add and one remove")
	assert.Equal(t, "remove", patch[0].Operation, "the patch should remove on 0")
	assert.Equal(t, "/0", patch[0].Path, "the patch should remove on 0")
	assert.Equal(t, "add", patch[1].Operation, "the patch should add on the last position")
	newEntry, err := json.Marshal(&patch[1].Value)
	require.NoError(t, err)
	assert.Equal(t, `{"test":"4"}`, string(newEntry))
	assert.Equal(t, "/2", patch[1].Path, "the patch should add on the last position")
}

func TestCollectionWindowDscMove(t *testing.T) {
	patch, e := CreatePatch([]byte(collectionWindowDscBefore), []byte(collectionWindowDscAfter))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should have one add and one remove")
	assert.Equal(t, "add", patch[0].Operation, "the patch should add on 0")
	assert.Equal(t, "/0", patch[0].Path, "the patch should add on 0")
	newEntry, err := json.Marshal(&patch[0].Value)
	require.NoError(t, err)
	assert.Equal(t, `{"test":"4"}`, string(newEntry))
	assert.Equal(t, "remove", patch[1].Operation, "the patch should remove last position")
	assert.Equal(t, "/3", patch[1].Path, "the patch should have descending order by path")
}

func TestCollectionAdd(t *testing.T) {
	patch, e := CreatePatch([]byte(collectionOne), []byte(collectionTwo))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should have one add")
	assert.Equal(t, "replace", patch[0].Operation, "the patch should add on 0")
	assert.Equal(t, "/0/test", patch[0].Path, "the patch should add on 0")
	assert.Equal(t, "add", patch[1].Operation, "the patch should add on 1")
	assert.Equal(t, "/1", patch[1].Path, "the patch should add on 1")
	newEntry, err := json.Marshal(&patch[1].Value)
	require.NoError(t, err)
	assert.Equal(t, `{"test":"1"}`, string(newEntry))
	newEntry, err = json.Marshal(&patch[0].Value)
	require.NoError(t, err)
	assert.Equal(t, `"2"`, string(newEntry))
}

func TestOneNullReplace(t *testing.T) {
	patch, e := CreatePatch([]byte(simplef), []byte(simpleG))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 1, "they should be equal")
	change := patch[0]
	assert.Equal(t, change.Operation, "replace", "they should be equal")
	assert.Equal(t, change.Path, "/b", "they should be equal")
	assert.Equal(t, change.Value, nil, "they should be equal")
}

func TestSame(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(simpleA))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 0, "they should be equal")
}

func TestOneStringReplace(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(simpleB))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 1, "they should be equal")
	change := patch[0]
	assert.Equal(t, change.Operation, "replace", "they should be equal")
	assert.Equal(t, change.Path, "/c", "they should be equal")
	assert.Equal(t, change.Value, "goodbye", "they should be equal")
}

func TestOneIntReplace(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(simpleC))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 1, "they should be equal")
	change := patch[0]
	assert.Equal(t, change.Operation, "replace", "they should be equal")
	assert.Equal(t, change.Path, "/b", "they should be equal")
	var expected = json.Number("100")
	assert.Equal(t, change.Value, expected, "they should be equal")
}

func TestOneAdd(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(simpleD))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 1, "they should be equal")
	change := patch[0]
	assert.Equal(t, change.Operation, "add", "they should be equal")
	assert.Equal(t, change.Path, "/d", "they should be equal")
	assert.Equal(t, change.Value, "foo", "they should be equal")
}

func TestOneRemove(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(simpleE))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 1, "they should be equal")
	change := patch[0]
	assert.Equal(t, change.Operation, "remove", "they should be equal")
	assert.Equal(t, change.Path, "/c", "they should be equal")
	assert.Equal(t, change.Value, nil, "they should be equal")
}

func TestVsEmpty(t *testing.T) {
	patch, e := CreatePatch([]byte(simpleA), []byte(empty))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 3, "they should be equal")
	sort.Sort(ByPath(patch))
	change := patch[0]
	assert.Equal(t, change.Operation, "remove", "they should be equal")
	assert.Equal(t, change.Path, "/a", "they should be equal")

	change = patch[1]
	assert.Equal(t, change.Operation, "remove", "they should be equal")
	assert.Equal(t, change.Path, "/b", "they should be equal")

	change = patch[2]
	assert.Equal(t, change.Operation, "remove", "they should be equal")
	assert.Equal(t, change.Path, "/c", "they should be equal")
}

func BenchmarkBigArrays(b *testing.B) {
	var a1, a2 []interface{}
	a1 = make([]interface{}, 100)
	a2 = make([]interface{}, 101)

	for i := 0; i < 100; i++ {
		a1[i] = i
		a2[i+1] = i
	}
	for i := 0; i < b.N; i++ {
		compareArray(a1, a2, "/")
	}
}

func BenchmarkBigArrays2(b *testing.B) {
	var a1, a2 []interface{}
	a1 = make([]interface{}, 100)
	a2 = make([]interface{}, 101)

	for i := 0; i < 100; i++ {
		a1[i] = i
		a2[i] = i
	}
	for i := 0; i < b.N; i++ {
		compareArray(a1, a2, "/")
	}
}
