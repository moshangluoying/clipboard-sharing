package main

import (
	"log"
)

func init() {
	err := enableProcessPrivileges()
	if err != nil {
		log.Printf("Warning: Failed to enable process privileges: %v\n", err)
	}
}

func main() {
	err := run()
	if err != nil {
		log.Println(err.Error())
	}
}
