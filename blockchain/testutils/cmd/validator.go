package cmd

import (
	"fmt"
	"path/filepath"
)

func RunValidator(workingDir string, validatorIndex int) error {

	rootDir := GetProjectRootPath()
	validatorConfigPath := filepath.Join(workingDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	_, _, err := runCommand(rootDir, "./madnet", "--config", validatorConfigPath, "validator")

	if err != nil {
		return err
	}
	return nil
}
