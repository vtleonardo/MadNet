package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func RunDeploy(workingDir string) error {

	factoryAddress := "0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc" // TODO - how to calculate this
	//rootPath := GetProjectRootPath()
	bridgeDir := GetBridgePath()

	_, _, err := runCommand(bridgeDir, "npx", "hardhat setHardhatIntervalMining --network dev --enable-auto-mine")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//npx hardhat --network "$NETWORK" --show-stack-traces deployContracts --input-folder ../scripts/generated
	_, _, err = runCommand(bridgeDir, "npx", "hardhat --show-stack-traces deployContracts --input-folder", workingDir)
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//addr="$(grep -Pzo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"
	// TODO - check how to create factoryAddress variable

	//export FACTORY_ADDRESS=$addr
	//if [[ -z "${FACTORY_ADDRESS}" ]]; then
	//echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
	//exit 1
	//fi
	// TODO - unnecessary check

	//for filePath in $(ls ../scripts/generated/config | xargs); do
	//   sed -e "s/registryAddress = .*/registryAddress = $FACTORY_ADDRESS/" "../scripts/generated/config/$filePath" > "../scripts/generated/config/$filePath".bk &&\
	//   mv "../scripts/generated/config/$filePath".bk "../scripts/generated/config/$filePath"
	//done
	// TODO - get the right path
	//err = filepath.Walk("GENERATED CONFIG FILE TO BE DEFINED IN CURRENT WORKDIR", func(path string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		log.Fatalf(err.Error())
	//		return err
	//	}
	//	fmt.Printf("File Name: %s\n", info.Name())
	//	return nil
	//})
	//if err != nil {
	//	return err
	//}
	// TODO - I think un necessary
	//for filePath in $(ls ../scripts/generated/config | xargs); do
	//sed -e "s/registryAddress = .*/registryAddress = AAAAAAAAA/" "../scripts/generated/config/validator1.toml" > "../scripts/generated/config/$filePath".bk &&\
	//mv "../scripts/generated/config/$filePath".bk "../scripts/generated/config/$filePath"
	//done
	//
	//cp ../scripts/base-files/owner.toml ../scripts/generated/owner.toml
	//sed -e "s/registryAddress = .*/registryAddress = $FACTORY_ADDRESS/" "../scripts/generated/owner.toml" > "../scripts/generated/owner.toml".bk &&\
	//mv "../scripts/generated/owner.toml".bk "../scripts/generated/owner.toml"

	// npx hardhat fundValidators --network $NETWORK
	_, _, err = runCommand(bridgeDir, "npx", "hardhat --network dev fundValidators")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//
	//if [[ ! -z "${SKIP_REGISTRATION}" ]]; then
	//echo "SKIPPING VALIDATOR REGISTRATION"
	//exit 0
	//fi
	_, isSet := os.LookupEnv("SKIP_REGISTRATION")
	if isSet {
		return nil
	}

	//
	//FACTORY_ADDRESS="$(echo "$addr" | sed -e 's/^"//' -e 's/"$//')"
	//
	//if [[ -z "${FACTORY_ADDRESS}" ]]; then
	//echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Not starting the registration!"
	//exit 1
	//fi
	//
	//cd $BRIDGE_DIR
	//cd $CURRENT_WD
	//npx hardhat setHardhatIntervalMining --network $NETWORK --interval 1000
	_, _, err = runCommand(bridgeDir, "npx", "hardhat setHardhatIntervalMining --interval 1000")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//./scripts/main.sh register
	err = RunRegister(workingDir)
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//
	//cd $BRIDGE_DIR
	//npx hardhat --network $NETWORK setMinEthereumBlocksPerSnapshot --factory-address $FACTORY_ADDRESS --block-num 10
	_, _, err = runCommand(bridgeDir, "npx", "hardhat setMinEthereumBlocksPerSnapshot --block-num 10 --factory-address", factoryAddress)
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//npx hardhat setHardhatIntervalMining --network $NETWORK
	_, _, err = runCommand(bridgeDir, "npx", "hardhat setHardhatIntervalMining")
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	//cd $CURRENT_WD
	//
	//if [[ -n "${AUTO_START_VALIDATORS}" ]]; then
	//if command -v gnome-terminal &>/dev/null; then
	//i=1
	//for filePath in $(ls ./scripts/generated/config | xargs); do
	//gnome-terminal --tab --title="Validator $i" -- bash -c "./scripts/main.sh validator $i"
	//i=$((i + 1))
	//done
	//exit 0
	//fi
	//echo -e "failed to auto start validators terminals, manually open a terminal for each validator and execute"
	//fi
	generatedValidatorConfigFiles := filepath.Join(workingDir, "scripts", "generated", "config")
	files, _ := ioutil.ReadDir(generatedValidatorConfigFiles)
	err = RunValidator(workingDir, len(files))
	if err != nil {
		log.Printf("Could not execute script: %v", err)
		return err
	}

	return nil
}
