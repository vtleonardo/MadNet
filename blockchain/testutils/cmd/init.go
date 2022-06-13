package cmd

import (
	"errors"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"os"
	"path/filepath"
)

func RunInit(workingDir string, numbersOfValidator int) error {

	err := RunGitHooks()
	if err != nil {
		return err
	}

	LA := 4242
	PA := 4343
	DA := 4444
	LSA := 8884

	//# Check that number of validators is valid
	//if ! [[ $1 =~ $re ]] || [[ $1 -lt 4 ]] || [[ $1 -gt 32 ]]; then
	//echo -e "Invalid number of validators [4-32]"
	//exit 1
	//fi
	if numbersOfValidator < 4 || numbersOfValidator > 32 {
		return errors.New("number of possible validators can be from 4 up to 32")
	}

	// TODO - Not necessary
	//if [ -f "./scripts/generated/genesis.json" ]; then
	//echo -e "Generated files already exist, run clean"
	//exit 1
	//fi
	//CLEAN_UP

	rootPath := testutils.GetProjectRootPath()
	for i := 1; i < numbersOfValidator; i++ {
		// TODO change to working direcotir password file
		cmd, err := executeCommand(rootPath, "ethkey generate --passwordfile ./scripts/base-files/passwordFile | cut -d' ' -f2")
		if err != nil {
			return err
		}
		stdout, err := cmd.Output()
		ADDRESS := stdout

		cmd, err = executeCommand(rootPath, "hexdump -n 16 -e '4/4 \"%08X\" 1 \"\\n\"' /dev/urandom")
		if err != nil {
			return err
		}
		stdout, err = cmd.Output()
		PK := stdout

		// TODO - use temp working dir and validator number
		sedCommand := fmt.Sprintf(`
			sed -e 's/defaultAccount = .*/defaultAccount = \"'"%s"'\"/' ./scripts/base-files/baseConfig |
			sed -e 's/rewardAccount = .*/rewardAccount = \"'"%s"'\"/' |
			sed -e 's/listeningAddress = .*/listeningAddress = \"0.0.0.0:'"%s"'\"/' |
			sed -e 's/p2pListeningAddress = .*/p2pListeningAddress = \"0.0.0.0:'"%s"'\"/' |
			sed -e 's/discoveryListeningAddress = .*/discoveryListeningAddress = \"0.0.0.0:'"%s"'\"/' |
			sed -e 's/localStateListeningAddress = .*/localStateListeningAddress = \"0.0.0.0:'"%s"'\"/' |
			sed -e 's/passcodes = .*/passcodes = \"scripts\/generated\/keystores\/passcodes.txt\"/' |
			sed -e 's/keystore = .*/keystore = \"scripts\/generated\/keystores\/keys\"/' |
			sed -e 's/stateDB = .*/stateDB = \"scripts\/generated\/stateDBs\/validator'"%d"'\/\"/' |
			sed -e 's/monitorDB = .*/monitorDB = \"scripts\/generated\/monitorDBs\/validator'"%d"'\/\"/' |
			sed -e 's/privateKey = .*/privateKey = \"'"%s"'\"/' > ./scripts/generated/config/validator%d.toml`,
			ADDRESS, ADDRESS, LA, PA, DA, LSA, i, i, PK, i)
		_, err = executeCommand(rootPath, sedCommand)
		if err != nil {
			return err
		}

		//echo "$ADDRESS=abc123" >> ./scripts/generated/keystores/passcodes.txt
		f, err := os.Create(filepath.Join(workingDir, "scripts", "generated", "keystores", "passcodes.txt"))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(fmt.Sprintf("%s=abc123", ADDRESS))
		if err != nil {
			return err
		}
		// TODO - does this exists?
		//mv ./keyfile.json ./scripts/generated/keystores/keys/$ADDRESS

		// Genesis
		genesisPath := filepath.Join(workingDir, "scripts", "generated", "genesis.json")
		jqCommand := fmt.Sprintf(`
			jq '.alloc += {"'"$(echo %s | cut -c3-)"'": {balance:"10000000000000000000000"}}' %s > 	%s.tmp && mv %s.tmp %s
		`, ADDRESS, genesisPath, genesisPath, genesisPath, genesisPath)
		_, err = executeCommand(rootPath, jqCommand)
		if err != nil {
			return err
		}

		LA = LA + 1
		PA = PA + 1
		DA = DA + 1
		LSA = LSA + 1
	}

	return nil
}
