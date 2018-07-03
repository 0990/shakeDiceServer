package main

import (
	"testing"
)

func TestSlice(t *testing.T) {
	var intArr [4]*int
	a:=1
	intArr[1] = &a
	t.Log(intArr)
}
