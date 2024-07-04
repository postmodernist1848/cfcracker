//go:build debug_cli

package main

import (
	"log"
	"os"
)

func debugCLI(as ...interface{}) {
	for _, a := range as {
		log.Println(a)
	}
	os.Exit(0)
}
