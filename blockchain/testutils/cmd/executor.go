package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// TODO - double check github action will pick this up

// SetCommandStdOut If ENABLE_SCRIPT_LOG env variable is set as 'true' the command will show scripts logs
func SetCommandStdOut(cmd *exec.Cmd) {

	flagValue, found := os.LookupEnv("ENABLE_SCRIPT_LOG")
	enabled, err := strconv.ParseBool(flagValue)

	if err == nil && found && enabled {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
}

func executeCommand(dir, command string, args ...string) (*exec.Cmd, []byte, error) {
	cmdArgs := strings.Split(strings.Join(args, " "), " ")

	cmd := exec.Command(command, cmdArgs...)
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("Error executing command: %v in dir: %v. %v", command, dir, err)
		return &exec.Cmd{}, nil, err
	}

	return cmd, output, err

}

func runCommand(dir, command string, args ...string) (*exec.Cmd, []byte, error) {
	cmdArgs := strings.Split(strings.Join(args, " "), " ")

	cmd := exec.Command(command, cmdArgs...)
	cmd.Dir = dir
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error executing command: %v in dir: %v. %v", command, dir, err)
		return &exec.Cmd{}, nil, err
	}
	return cmd, nil, err

}

// TODO - make it wait()
// CreateTempFolder creates a test working folder in the OS temporary resources folder
func CreateTempFolder() (string, error) {
	file, err := ioutil.TempDir("", "unittest")
	if err != nil {
		return "", err
	}

	return file, nil
}

// TODO - make it wait()
func CopyFileToFolder(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	_, err = os.Create(dst)
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// GetProjectRootPath returns the project root path
func GetProjectRootPath() string {

	rootPath := []string{string(os.PathSeparator)}

	cmd := exec.Command("go", "list", "-m", "-f", "'{{.Dir}}'", "github.com/MadBase/MadNet")
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting project root path: %v", err)
		return ""
	}
	path := string(stdout)
	path = strings.ReplaceAll(path, "'", "")
	path = strings.ReplaceAll(path, "\n", "")

	pathNodes := strings.Split(path, string(os.PathSeparator))
	for _, pathNode := range pathNodes {
		rootPath = append(rootPath, pathNode)
	}

	return filepath.Join(rootPath...)
}

// GetBridgePath return the bridge folder path
func GetBridgePath() string {
	rootPath := GetProjectRootPath()
	bridgePath := filepath.Join(rootPath, "bridge")

	return bridgePath
}

//RandomHex hexdump -n 16 -e '4/4 "%08X" 1 "\n"' /dev/urandom
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

func ReplaceConfigurationFile(workingDir, address, privateKey string, listeningPort, p2pPort, discoveryPort, localStatePort, index int) error {
	baseConfigFile, err := os.ReadFile(filepath.Join(workingDir, "scripts", "base-files", "baseConfig"))
	if err != nil {
		log.Fatalf("Error reading base configuration file - %v", err)
		return err
	}
	fileContent := string(baseConfigFile)
	validatorFileName := "validator" + strconv.Itoa(index)

	regex := regexp.MustCompile(`defaultAccount = .*`)
	result := regex.ReplaceAllString(fileContent, "defaultAccount = \""+address+"\"")
	regex = regexp.MustCompile(`rewardAccount = .*`)
	result = regex.ReplaceAllString(result, "rewardAccount = \""+address+"\"")
	regex = regexp.MustCompile(`listeningAddress = .*`)
	result = regex.ReplaceAllString(result, "listeningAddress = \"0.0.0.0:"+strconv.Itoa(listeningPort)+"\"")
	regex = regexp.MustCompile(`p2pListeningAddress = .*`)
	result = regex.ReplaceAllString(result, "p2pListeningAddress = \"0.0.0.0:"+strconv.Itoa(p2pPort)+"\"")
	regex = regexp.MustCompile(`discoveryListeningAddress = .*`)
	result = regex.ReplaceAllString(result, "discoveryListeningAddress = \"0.0.0.0:"+strconv.Itoa(discoveryPort)+"\"")
	regex = regexp.MustCompile(`localStateListeningAddress = .*`)
	result = regex.ReplaceAllString(result, "localStateListeningAddress = \"0.0.0.0:"+strconv.Itoa(localStatePort)+"\"")
	regex = regexp.MustCompile(`passcodes = .*`)
	result = regex.ReplaceAllString(result, "passcodes = \""+filepath.Join("scripts", "generated", "keystores", "passcodes.txt")+"\"") // TODO - check file path root project or working dir
	regex = regexp.MustCompile(`keystore = .*`)
	result = regex.ReplaceAllString(result, "keystore = \""+filepath.Join("scripts", "generated", "keystores", "keys")+"\"")
	regex = regexp.MustCompile(`stateDB = .*`)
	result = regex.ReplaceAllString(result, "stateDB = \""+filepath.Join("scripts", "generated", "stateDBs", validatorFileName)+"\"")
	regex = regexp.MustCompile(`monitorDB = .*`)
	result = regex.ReplaceAllString(result, "monitorDB = \""+filepath.Join("scripts", "generated", "monitorDBs", validatorFileName)+"\"")
	regex = regexp.MustCompile(`privateKey = .*`)
	result = regex.ReplaceAllString(result, "privateKey = \""+privateKey+"\"")

	f, err := os.Create(filepath.Join(workingDir, "scripts", "generated", "config", validatorFileName+".toml"))
	if err != nil {
		log.Fatalf("Error creating validator configuration file - %v", err)
		return err
	}
	_, err = fmt.Fprintf(f, "%s", result)
	if err != nil {
		log.Fatalf("Error writing on validator configuration file - %v", err)
		return err
	}
	defer f.Close()
	return nil
}

func ReplaceGenesisBalance(workingDir string) error {
	genesisFilePath := filepath.Join(workingDir, "scripts", "base-files", "genesis.json")
	genesisConfigFile, err := os.ReadFile(genesisFilePath)
	if err != nil {
		log.Fatalf("Error reading base configuration file - %v", err)
		return err
	}
	fileContent := string(genesisConfigFile)
	regex := regexp.MustCompile(`balance.*`)
	result := regex.ReplaceAllString(fileContent, "balance\": \"10000000000000000000000\" }")

	f, err := os.Create(genesisFilePath)
	if err != nil {
		log.Fatalf("Error creating modified genesis.json file - %v", err)
		return err
	}
	_, err = fmt.Fprintf(f, "%s", result)
	if err != nil {
		log.Fatalf("Error writing on new genesis.json file - %v", err)
		return err
	}
	defer f.Close()
	return nil
}
