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
	err = constructSource(source)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
