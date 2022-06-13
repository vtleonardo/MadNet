package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func executeCommand(dir string, command ...string) error {
	args := strings.Split(strings.Join(command, " "), " ")
	cmd := exec.Cmd{
		Args:   args,
		Dir:    dir,
		Stdin:  os.Stdout,
		Stdout: os.Stdin,
		Stderr: os.Stderr,
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("Could not execute command: %v", args)
		return err
	}

	return nil
}

func executeCommandWithOutput(dir string, command ...string) (string, error) {
	args := strings.Split(strings.Join(command, " "), " ")
	cmd := exec.Cmd{
		Args:   args,
		Dir:    dir,
		Stdin:  os.Stdout,
		Stdout: os.Stdin,
		Stderr: os.Stderr,
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("Could not execute command: %v", args)
		return "nil", err
	}

	stdout, err := cmd.Output()
	return string(stdout), err

}

func CreateTempFolder() string {
	// create tmp folder
	file, err := ioutil.TempFile("dir", "prefix")
	if err != nil {
		log.Fatal(err)
	}

	return file.Name() // For example "dir/prefix054003078"
}

func CopyFileToFolder(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
