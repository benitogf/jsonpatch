package jsonpatch

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var subArray1_current = `{"created":1739943104628936621,"updated":0,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3}]}}`
var subArray1_target = `{"created":1739943104628936621,"updated":1739942662496282458,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3}]}}`

var subArray2_current = `{"created":1739943104628936621,"updated":1739942662496282458,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3}]}}`
var subArray2_target = `{"created":1739943104628936621,"updated":1739942662496282459,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3}]}}`

var subArray3_current = `{"created":1739943104628936621,"updated":1739942662496282459,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3}]}}`
var subArray3_target = `{"created":1739943104628936621,"updated":1739942662496282460,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3}]}}`

var subArray4_current = `{"created":1739943104628936621,"updated":1739942662496282460,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3}]}}`
var subArray4_target = `{"created":1739943104628936621,"updated":1739942662496282461,"index":"test","path":"test","data":{"subFields":[{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3},{"one":"B","two":"two","three":3},{"one":"P","two":"two","three":3}]}}`

var subArray5_current = `{"created":1739954651885800613,"updated":1739954651890302236,"index":"game","path":"","data":{"id":"18258f8986f08d85","deck":1739954651885767932,"started":1739954651885767932,"ended":0,"burn":false,"burning":0,"overriden":0,"overrider":"","voided":0,"voider":"","edited":0,"editor":"","status":"ongoing","cards":[{"suit":1,"value":10,"owner":"P","show":false},{"suit":1,"value":10,"owner":"B","show":false},{"suit":1,"value":10,"owner":"P","show":false},{"suit":1,"value":10,"owner":"B","show":false}],"tally":{"player":0,"banker":0},"previousCards":null,"result":{"status":"","playerPair":false,"bankerPair":false,"natural8":false,"natural9":false,"lucky6_2":false,"lucky6_3":false},"fakeCards":false}}`
var subArray5__target = `{"created":1739954651885800613,"updated":1739954651891192258,"index":"game","path":"","data":{"id":"18258f8986f08d85","deck":1739954651885767932,"started":1739954651885767932,"ended":0,"burn":false,"burning":0,"overriden":0,"overrider":"","voided":0,"voider":"","edited":0,"editor":"","status":"ongoing","cards":[{"suit":1,"value":10,"owner":"P","show":true},{"suit":1,"value":10,"owner":"B","show":true},{"suit":1,"value":10,"owner":"P","show":true},{"suit":1,"value":10,"owner":"B","show":true},{"suit":1,"value":10,"owner":"P","show":false}],"tally":{"player":0,"banker":0},"previousCards":null,"result":{"status":"","playerPair":false,"bankerPair":false,"natural8":false,"natural9":false,"lucky6_2":false,"lucky6_3":false},"fakeCards":false}}`

type Card struct {
	Suit  int    `json:"suit"`
	Value int    `json:"value"`
	Owner string `json:"owner"`
	Show  bool   `json:"show"`
}

type GameData struct {
	Cards []Card `json:"cards"`
}

type Game struct {
	Data GameData `json:"data"`
}

func TestSubfieldArray(t *testing.T) {
	patch, e := CreatePatch([]byte(subArray1_current), []byte(subArray1_target))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should update 2 fields")

	patch, e = CreatePatch([]byte(subArray2_current), []byte(subArray2_target))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should update 2 fields")

	patch, e = CreatePatch([]byte(subArray3_current), []byte(subArray3_target))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should update 2 fields")

	patch, e = CreatePatch([]byte(subArray4_current), []byte(subArray4_target))
	assert.NoError(t, e)
	assert.Equal(t, 2, len(patch), "the patch should update 2 fields")

	patch, e = CreatePatch([]byte(subArray5_current), []byte(subArray5__target))
	assert.NoError(t, e)
	// log.Println("OP", patch)
	assert.Equal(t, 8, len(patch), "the patch should update 2 fields")

	_patch, err := json.Marshal(patch)
	require.NoError(t, err)
	obj, err := DecodePatch([]byte(_patch))
	require.NoError(t, err)
	patched, err := obj.Apply([]byte(subArray5_current))
	require.NoError(t, err)

	var _target = Game{}
	json.Unmarshal([]byte(subArray5__target), &_target)
	var _patched = Game{}
	json.Unmarshal([]byte(patched), &_patched)

	require.True(t, reflect.DeepEqual(_target, _patched))
}
