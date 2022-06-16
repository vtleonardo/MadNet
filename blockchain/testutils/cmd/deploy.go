package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func RunDeploy(workingDir string) (string, error) {

	bridgeDir := GetBridgePath()
	_, _, err := executeCommand(bridgeDir, "npx", "hardhat --network dev setHardhatIntervalMining --enable-auto-mine")
	if err != nil {
		return "", err
	}

	_, output, err := executeCommand(bridgeDir, "npx", "hardhat --network dev --show-stack-traces deployContracts --input-folder", filepath.Join(workingDir, "scripts", "generated"))
	if err != nil {
		return "", err
	}
	firstLogLine := strings.Split(string(output), "\n")[0]
	addressLine := strings.Split(firstLogLine, ":")
	factoryAddress := strings.TrimSpace(addressLine[len(addressLine)-1])

	err = ReplaceOwnerRegistryAddress(workingDir, factoryAddress)
	if err != nil {
		return "", err
	}
	err = ReplaceValidatorsRegistryAddress(workingDir, factoryAddress)
	if err != nil {
		return "", err
	}

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev fundValidators --config-path", filepath.Join(workingDir, "scripts", "generated", "config"))
	if err != nil {
		return "", err
	}

	_, isSet := os.LookupEnv("SKIP_REGISTRATION")
	if isSet {
		return "", nil
	}

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev setHardhatIntervalMining --interval 1000")
	if err != nil {
		return "", err
	}

	err = RunRegister(workingDir, factoryAddress)
	if err != nil {
		return "", err
	}

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev setMinEthereumBlocksPerSnapshot --block-num 10 --factory-address", factoryAddress)
	if err != nil {
		return "", err
	}

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev setHardhatIntervalMining")
	if err != nil {
		return "", err
	}

	generatedValidatorConfigFiles := filepath.Join(workingDir, "scripts", "generated", "config")
	files, _ := ioutil.ReadDir(generatedValidatorConfigFiles)
	err = RunValidator(workingDir, len(files))
	if err != nil {
		return "", err
	}

	return factoryAddress, nil
}
