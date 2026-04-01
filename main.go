package main

import "fmt"
import "diane/internal/agent"


func main() {
	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)
	body, err := client.CheckHealth()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
}
