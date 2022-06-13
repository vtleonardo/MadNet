package cmd

import (
	"fmt"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"path/filepath"
)

func RunValidator(validatorIndex int) error {

	rootDir := testutils.GetProjectRootPath()
	validatorConfigPath := filepath.Join(rootDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	_, err := executeCommand(rootDir, "./madnet --config", validatorConfigPath, "validator")

	if err != nil {
		return err
	}
	return nil
}
