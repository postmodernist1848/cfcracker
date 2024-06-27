package main

import (
	"os"
	"testing"
)

func TestConstructSource(t *testing.T) {
	source, err := os.ReadFile("solutions/watermelon.cpp")
	if err != nil {
		panic(err)
	}
	parts, err := intoParts(source)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	testCases := TestCases{
		[][]int{},
		false,
	}

	constructed := constructFromParts(parts, testCases)
	err = os.WriteFile("solutions/out.cpp", []byte(constructed), 0666)
	if err != nil {
		t.Errorf("could not write to file: %v", err)
	}
	// TODO: compile and test
}
