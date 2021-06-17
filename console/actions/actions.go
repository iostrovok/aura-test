package main

/*
	Application for simple test the service to create and destroy.
	HOST & URL are defined in ../helpers/helpers.go
*/

import (
	"fmt"
	"net/http"

	"github.com/iostrovok/aura-test/console/helpers"
)

func main() {
	client := &http.Client{}
	id := helpers.CreateSession(client)
	if id == "" {
		return
	}

	code, err := helpers.DestroySession(client, id)
	fmt.Printf("code: %d, err: %+v\n", code, err)

	code, err = helpers.DestroySession(client, id)
	fmt.Printf("code: %d, err: %+v\n", code, err)

	helpers.List()
}
