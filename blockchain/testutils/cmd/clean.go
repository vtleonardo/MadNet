package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunClean(workingDir string) error {

	//Remove content
	rootPath := GetProjectRootPath()
	err := os.Remove(filepath.Join(rootPath, "keyfile.json"))
	if err != nil {
		fmt.Print("Trying to remove a file that can or cannot be here. Not a problem")
	}
	return nil
}
