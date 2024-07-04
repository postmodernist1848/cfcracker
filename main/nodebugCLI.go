//go:build !debug_cli

package main

import "cfcracker/client"

func debugCLI(_ string, _ string, _ *client.Client, _ *string, _ *string) {
}
