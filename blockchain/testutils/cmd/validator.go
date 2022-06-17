package cmd

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
)

func RunValidator(workingDir string, validatorIndex int) error {

	rootDir := GetProjectRootPath()
	//validatorConfigPath := filepath.Join(workingDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	// TODO - this will be runCommand() once fixed
	configurationFileDir := filepath.Join(workingDir, "scripts", "generated", "config")
	files, err := ioutil.ReadDir(configurationFileDir)
	for _, file := range files {
		src := filepath.Join(configurationFileDir, file.Name())
		dst := filepath.Join(rootDir, "scripts", "aaa", file.Name())
		_, err := CopyFileToFolder(src, dst)
		if err != nil {
			log.Fatalf("Error copying config file to working directory", err)
			return err
		}
	}
	_, _, err = executeCommand(rootDir, "make", "build")

	for i := 0; i < validatorIndex; i++ {
		validatorI := "validatorI" + strconv.Itoa(i)
		_, _, err = executeCommand(rootDir, "./madnet", "--config", filepath.Join("scripts", "aaa", validatorI), "validator")
	}

	if err != nil {
		return err
	}
	return nil
}
