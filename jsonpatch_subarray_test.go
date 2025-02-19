package jsonpatch

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var subArray1_current = `{"created":1739943104628936621,"updated":0,"index":"test","path":"test","data":{"subFields":[{"one":"one","two":"two","three":3}]}}`
var subArray1_target = `{"created":1739943104628936621,"updated":1739942662496282458,"index":"test","path":"test","data":{"subFields":[{"one":"one","two":"two","three":3},{"one":"one","two":"two","three":3}]}}`

var subArray2_current = `{"created":1739943104628936621,"updated":1739942662496282458,"index":"test","path":"test","data":{"subFields":[{"one":"one","two":"two","three":3},{"one":"one","two":"two","three":3}]}}`
var subArray2_target = `{"created":1739943104628936621,"updated":1739942662496282459,"index":"test","path":"test","data":{"subFields":[{"one":"one","two":"two","three":3},{"one":"one","two":"two","three":3},{"one":"one","two":"two","three":3}]}}`

func TestSubfieldArray(t *testing.T) {
	patch, e := CreatePatch([]byte(subArray1_current), []byte(subArray1_target))
	assert.NoError(t, e)
	assert.Equal(t, len(patch), 2, "the patch should update 2 fields")

	patch, e = CreatePatch([]byte(subArray2_current), []byte(subArray2_target))
	assert.NoError(t, e)
	log.Println("OP", patch)
	assert.Equal(t, len(patch), 2, "the patch should update 2 fields")
}
