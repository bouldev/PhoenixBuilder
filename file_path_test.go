package main

import "testing"
import (
	"fmt"
	"os"
	"path/filepath"
)

func TestFile(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(ex, exPath)
}
