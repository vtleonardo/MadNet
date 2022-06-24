package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	password = "abc123"
)

func RunInit(workingDir string, numbersOfValidator int) ([]string, error) {

	// Ports
	listeningPort := 4242
	p2pPort := 4343
	discoveryPort := 4444
	localStataPort := 8884

	// Validator instance check
	if numbersOfValidator < 4 || numbersOfValidator > 32 {
		return nil, errors.New("number of possible validators can be from 4 up to 32")
	}

	// Build validator configuration files
	rootPath := GetProjectRootPath()
	tempFile := temporaryFile()
	passcodeFilePath := filepath.Join(workingDir, "keystores", "passcodes.txt")
	passcodesFile, err := os.Create(passcodeFilePath)
	if err != nil {
		return nil, err
	}
	defer passcodesFile.Close()

	validatorAddresses := make([]string, 0)
	for i := 0; i < numbersOfValidator; i++ {

		_, stdout, err := executeCommand(rootPath, "ethkey", "generate --passwordfile "+tempFile)
		if err != nil {
			return nil, err
		}
		address := string(stdout[:])
		address = strings.ReplaceAll(address, "Address: ", "")
		address = strings.ReplaceAll(address, "\n", "")
		validatorAddresses = append(validatorAddresses, address)

		// Generate private key
		privateKey, err := RandomHex(16) // TODO - is this right?
		if err != nil {
			return nil, err
		}

		// Validator configuration file
		err = ReplaceConfigurationFile(workingDir, address, privateKey, listeningPort, p2pPort, discoveryPort, localStataPort, i)
		if err != nil {
			return nil, err
		}

		// Passcode file
		passcodesFile, err := os.OpenFile(passcodeFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		_, err = passcodesFile.WriteString(fmt.Sprintf("%s=%s\n", address, password))
		if err != nil {
			return nil, err
		}

		// Genesis
		err = ReplaceGenesisBalance(workingDir)
		if err != nil {
			return nil, err
		}

		// Keyfile.json
		_, err = CopyFileToFolder(filepath.Join(rootPath, "keyfile.json"), filepath.Join(workingDir, "keystores", "keys", address))
		if err != nil {
			fmt.Print("Error copying keyfile.json into generated folder")
		}
		err = os.Remove(filepath.Join(rootPath, "keyfile.json"))
		if err != nil {
			fmt.Print("Trying to remove keyfile.json ")
		}

		listeningPort += 1
		p2pPort += 1
		discoveryPort += 1
		localStataPort += 1
	}

	return validatorAddresses, nil
}

func temporaryFile() string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(password)
	if err != nil {
		log.Fatal(err)
	}
	return f.Name()
}
