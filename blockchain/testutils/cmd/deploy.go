package cmd

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts"
	"os"
	"path/filepath"
	"strings"
)

func RunDeploy(workingDir string, accountPrivateKeyMap map[accounts.Account]*ecdsa.PrivateKey) (string, error) {

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

	// Replace filename

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev fundValidators --config-path", filepath.Join(workingDir, "scripts", "generated", "config"))
	if err != nil {
		return "", err
	}

	_, isSet := os.LookupEnv("SKIP_REGISTRATION")
	if isSet {
		return "", nil
	}

	_, _, err = executeCommand(bridgeDir, "npx", "hardhat --network dev setHardhatIntervalMining --interval 100")
	if err != nil {
		return "", err
	}

	var validatorsAddressList []string
	for k := range accountPrivateKeyMap {
		if k.Address.String() != "0x546F99F244b7B58B855330AE0E2BC1b30b41302F" {
			validatorsAddressList = append(validatorsAddressList, k.Address.String())
		}
	}

	err = RunRegister(factoryAddress, validatorsAddressList)
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

	//generatedValidatorConfigFiles := filepath.Join(workingDir, "scripts", "generated", "config")
	//files, _ := ioutil.ReadDir(generatedValidatorConfigFiles)
	//err = RunValidator(workingDir, len(files))
	//if err != nil {
	//	return "", err
	//}

	return factoryAddress, nil
}
