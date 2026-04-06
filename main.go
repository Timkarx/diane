package main

import (
	"diane/internal/agent"
	"fmt"
	"os"
)


func main() {
	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)
	res, err := client.Prompt("What gurantees does c++ give about function arg evaluation?")
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(res.AsPlainText())
	os.Exit(0)
}
