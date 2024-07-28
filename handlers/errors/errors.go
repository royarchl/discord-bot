package errors

import (
	"log"
)

func CheckNilErr(e error) {
	if e != nil {
		log.Fatalf("Error message: %v", e)
	}
}
