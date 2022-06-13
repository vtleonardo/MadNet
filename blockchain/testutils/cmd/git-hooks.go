package cmd

import (
	"github.com/MadBase/MadNet/blockchain/testutils"
	"log"
)

func RunGitHooks() error {

	rootPath := testutils.GetProjectRootPath()
	_, err := executeCommand(rootPath, "git config core.hooksPath scripts/githooks 2>/dev/null")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	return nil
}
