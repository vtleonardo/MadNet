package cmd

import (
	"fmt"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"path/filepath"
)

func RunValidator(validatorIndex int) error {

	rootDir := testutils.GetProjectRootPath()
	validatorConfigPath := append(rootDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	err := executeCommand(filepath.Join(rootDir...), "./madnet --config", filepath.Join(validatorConfigPath...), "validator")

	if err != nil {
		return err
	}
	return nil
}
