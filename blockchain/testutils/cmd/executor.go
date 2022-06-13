package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// TODO - double check github action will pick this up

// SetCommandStdOut If ENABLE_SCRIPT_LOG env variable is set as 'true' the command will show scripts logs
func SetCommandStdOut(cmd *exec.Cmd) {

	flagValue, found := os.LookupEnv("ENABLE_SCRIPT_LOG")
	enabled, err := strconv.ParseBool(flagValue)

	if err == nil && found && enabled {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
}

func executeCommand(dir string, command ...string) (exec.Cmd, error) {
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
		return exec.Cmd{}, err
	}

	return cmd, nil
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
