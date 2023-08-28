package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"buildpack"
)

func main() {
	cmd := exec.Command(buildpack.BinaryName, os.Args[1:]...)
	fmt.Printf("running %s\n", cmd.String())

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		log.Fatal(err)
	}

}
