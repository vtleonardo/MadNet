package cmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func RunRegister(workingDir, factoryAddress string) error {

	bridgeDir := GetBridgePath()
	keys := filepath.Join(workingDir, "scripts", "generated", "keystores", "keys")

	// Build validator names
	files, err := ioutil.ReadDir(keys)
	validators := make([]string, 0)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "0x546f99f244b") {
			continue
		}
		validators = append(validators, file.Name())
	}

	// Register validator
	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev --show-stack-traces registerValidators --factory-address", factoryAddress, strings.Join(validators, " "))
	if err != nil {
		return err
	}

	return nil
}
