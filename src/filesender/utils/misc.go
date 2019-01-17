package utils

import (
	"fmt"
	"os"
)

//
func OutputAndExit(data string) {

	fmt.Println(data)
	os.Exit(-1)
}
