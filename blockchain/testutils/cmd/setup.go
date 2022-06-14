package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func RunSetup(workingDir string) error {

	// Create directories
	folders := []string{
		filepath.Join("scripts", "base-files"),
		filepath.Join("scripts", "generated", "monitorDBs"),
		filepath.Join("scripts", "generated", "config"),
		filepath.Join("scripts", "generated", "keystores"),
		filepath.Join("assets", "test", "keys"),
	}
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(workingDir, folder), os.ModePerm); err != nil {
			fmt.Printf("Error creating configuration folders: %v", err)
			return err
		}
	}

	// Copy configuration files
	rootPath := GetProjectRootPath()
	configurationFileDir := filepath.Join(rootPath, "scripts", "base-files")
	files, err := ioutil.ReadDir(configurationFileDir)
	if err != nil {
		log.Fatalf("Error reading configuaration file dir path: %s", configurationFileDir)
		return err
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
		src := filepath.Join(configurationFileDir, file.Name())
		dst := filepath.Join(workingDir, "scripts", "base-files", file.Name())
		_, err = CopyFileToFolder(src, dst)
		if err != nil {
			log.Fatalf("Error copying config file to working directory", err)
			return err
		}
	}

	// Copy asset files
	assetFileDir := filepath.Join(rootPath, "assets", "test", "keys")
	files, err = ioutil.ReadDir(assetFileDir)
	if err != nil {
		log.Fatalf("Error reading asset file dir path: %s", assetFileDir)
		return err
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
		src := filepath.Join(assetFileDir, file.Name())
		dst := filepath.Join(workingDir, "assets", "test", "keys", file.Name())
		_, err = CopyFileToFolder(src, dst)
		if err != nil {
			log.Fatalf("Error copying assets file to working directory: %v", err)
			return err
		}
	}
	_, err = CopyFileToFolder(filepath.Join(rootPath, "assets", "test", "blockheaders.txt"), filepath.Join(workingDir, "assets", "test", "blockheaders.txt"))
	if err != nil {
		log.Fatalf("Error reading asset blockheaders: %s", assetFileDir)
		return err
	}
	_, err = CopyFileToFolder(filepath.Join(rootPath, "assets", "test", "passcodes.txt"), filepath.Join(workingDir, "assets", "test", "passcodes.txt"))
	if err != nil {
		log.Fatalf("Error reading asset passcodes: %s", assetFileDir)
		return err
	}

	return nil
}
