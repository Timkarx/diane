package main

import (
	"diane/internal/agent"
)


func main() {
	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)
	client.EvaluateInput("What gurantees does c++ give about function arg evaluation?")
}
