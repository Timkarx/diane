package main

import (
	"diane/internal/agent"
)


func main() {
	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)
	client.EvaluateInput("Hello how are you?")
}
