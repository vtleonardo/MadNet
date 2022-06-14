package cmd

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func RunRegister() error {

	bridgeDir := GetBridgePath()
	rootDir := GetProjectRootPath()
	factoryAddress := "0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc" // TODO - how to calculate this
	// TODO - get the right path
	keys := filepath.Join(rootDir, "scripts", "generated", "keystores", "keys")

	// Build validator names
	files, err := ioutil.ReadDir(keys)
	validators := make([]string, 0)
	if err != nil {
		return err
	}
	for _, file := range files {
		validators = append(validators, file.Name())
	}

	// Register validator
	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev --show-stack-traces registerValidators --factory-address", factoryAddress, strings.Join(validators, " "))
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	return nil
}
