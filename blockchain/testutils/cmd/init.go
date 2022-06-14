package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunInit(workingDir string, numbersOfValidator int) error {

	// Resources setup
	err := RunSetup(workingDir)
	if err != nil {
		return err
	}

	err = RunGitHooks()
	if err != nil {
		return err
	}

	// Ports
	listeningPort := 4242
	p2pPort := 4343
	discoveryPort := 4444
	localStatatePort := 8884

	// Validator instance check
	if numbersOfValidator < 4 || numbersOfValidator > 32 {
		return errors.New("number of possible validators can be from 4 up to 32")
	}

	rootPath := GetProjectRootPath()
	for i := 1; i < numbersOfValidator; i++ {

		passwordFilePath := filepath.Join(workingDir, "scripts", "base-files", "passwordFile")
		_, stdout, err := executeCommand(rootPath, "ethkey", "generate --passwordfile "+passwordFilePath)
		if err != nil {
			return err
		}
		address := string(stdout[:])
		address = strings.ReplaceAll(address, "Address: ", "")
		address = strings.ReplaceAll(address, "\n", "")

		// Remove keyfile.json
		err = os.Remove(filepath.Join(rootPath, "keyfile.json"))
		if err != nil {
			fmt.Print("Trying to remove keyfile.json that can or cannot be here. Not a problem")
		}

		// Generate private key
		privateKey, err := RandomHex(16)
		if err != nil {
			return err
		}

		// Validator configuration file
		err = ReplaceConfigurationFile(workingDir, address, privateKey, listeningPort, p2pPort, discoveryPort, localStatatePort, i)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(workingDir, "scripts", "generated", "keystores", "passcodes.txt"))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(fmt.Sprintf("%s=abc123", address))
		if err != nil {
			return err
		}

		// Genesis
		err = ReplaceGenesisBalance(workingDir)
		if err != nil {
			return err
		}

		listeningPort += 1
		p2pPort += 1
		discoveryPort += 1
		localStatatePort += 1
	}

	return nil
}
