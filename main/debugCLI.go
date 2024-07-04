//go:build debug_cli

package main

import (
	"cfcracker/client"
	"log"
	"os"
)

func debugCLI(handleOrEmail string, password string, c *client.Client, sourcePath *string, createConfigPath *string) {
	log.Println(handleOrEmail)
	log.Println(password)
	log.Println(c)
	log.Println(*sourcePath)
	log.Println(*createConfigPath)
	os.Exit(0)
}
