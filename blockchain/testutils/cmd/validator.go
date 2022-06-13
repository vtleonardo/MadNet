package cmd

import (
	"fmt"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"path/filepath"
)

func RunValidator(validatorIndex int) error {

	//./madnet --config ./scripts/generated/config/validator$1.toml validator
	rootDir := filepath.Join(testutils.GetProjectRootPath()...)

	validatorConfig := append(rootDir, "scripts", "generated", "config", fmt.Sprintf("validator%d.toml", validatorIndex))

	err := executeCommand(rootDir, "./madnet --config", filepath.Join(validatorConfig...), "validator")
	if err != nil {
		return err
	}
	return nil
}
