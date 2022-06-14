package cmd

import (
	"fmt"
	"path/filepath"
)

func RunValidator(validatorIndex int) error {

	rootDir := GetProjectRootPath()
	validatorConfigPath := filepath.Join(rootDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	_, _, err := executeCommand(rootDir, "./madnet", "--config", validatorConfigPath, "validator")

	if err != nil {
		return err
	}
	return nil
}
