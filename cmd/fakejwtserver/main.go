package main

import "github.com/stackitcloud/fake-jwt-server/cmd/fakejwtserver/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
