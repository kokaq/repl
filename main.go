/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/kokaq/repl/cmd"
)

func main() {
	cmd.NewReplClient(":9001").Start()
}
