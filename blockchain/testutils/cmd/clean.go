package cmd

import (
	"github.com/MadBase/MadNet/blockchain/testutils"
	"os"
	"path/filepath"
)

func RunClean(workingDir string) error {

	// TODO - not needed?
	// Remove folder content
	//rootPath := testutils.GetProjectRootPath()
	//generatedFolder := append(rootPath,"scripts","generated")
	//err := os.RemoveAll(filepath.Join(generatedFolder...))
	//if err != nil {
	//	return err
	//}

	// Create folders
	folders := []string{
		filepath.Join("scripts", "generated", "monitorDBs"),
		filepath.Join("scripts", "generated", "config"),
		filepath.Join("scripts", "generated", "keystores", "keys"),
		filepath.Join("scripts", "generated", "keystores", "passcodes.txt"),
	}
	for _, folder := range folders {
		if err := os.Mkdir(filepath.Join(workingDir, folder), os.ModePerm); err != nil {
			return err
		}
	}

	// Copy config files
	rootPath := testutils.GetProjectRootPath()
	srcGenesis := append(rootPath, "scripts", "base-file", "genesis.json")
	_, err := CopyFileToFolder(filepath.Join(srcGenesis...), filepath.Join(workingDir, "scripts", "generated"))
	if err != nil {
		return err
	}
	srcKey := append(rootPath, "scripts", "base-file", "0x546f99f244b7b58b855330ae0e2bc1b30b41302f")
	_, err = CopyFileToFolder(filepath.Join(srcKey...), filepath.Join(workingDir, "scripts", "generated", "keystores", "keys"))
	if err != nil {
		return err
	}

	return nil
}
