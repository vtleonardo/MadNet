package cmd

import (
	"log"
)

func RunGitHooks() error {

	rootPath := GetProjectRootPath()
	_, _, err := executeCommand(rootPath, "git", "config core.hooksPath scripts/githooks")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}
	return nil
}
