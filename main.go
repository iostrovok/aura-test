package main

import (
	"context"

	"github.com/iostrovok/aura-test/server"
)

func main() {
	server.Start(context.Background())
}
