package testutils

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/logging"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	os "os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	scriptDeploy            = "deploy"
	scriptRegisterValidator = "register_test"
	scriptStartHardHatNode  = "hardhat_node"
	scriptInit              = "init"
	hardHatProcessId        = "HARDHAT_PROCESS_ID"
)

func getEthereumDetails() (*ethereum.Details, error) {
	details, err := ethereum.NewEndpoint(
		getConfigurationProperty("endpoint"),
		"../../assets/test/keys",
		"../../assets/test/passcodes.txt",
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F",
		1,
		500,
		0,
	)
	return details, err
}

func waitForHardHatNode(ctx context.Context) error {
	c := http.Client{}
	msg := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "eth_chainId",
		Params:  make([]byte, 0),
	}

	params, err := json.Marshal(make([]string, 0))
	if err != nil {
		log.Printf("could not run hardhat node: %v", err)
		return err
	}
	msg.Params = params

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(msg)
	if err != nil {
		log.Printf("Error creating a buffer json encoder: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			body := bytes.NewReader(buff.Bytes())
			_, err := c.Post(
				getConfigurationProperty("endpoint"),
				"application/json",
				body,
			)
			if err != nil {
				continue
			}
			log.Printf("HardHat node started correctly")
			return nil
		}
	}
}

func isHardHatRunning() (bool, error) {
	var client = http.Client{Timeout: 2 * time.Second}
	resp, err := client.Head(getConfigurationProperty("endpoint"))
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true, nil
	}

	return false, nil
}

func startHardHat(t *testing.T, ctx context.Context) *ethereum.Details {

	log.Printf("Starting HardHat ...")
	err := runScriptHardHatNode()
	assert.Nilf(t, err, "Error starting hardhat node")

	err = waitForHardHatNode(ctx)
	assert.Nilf(t, err, "Failed to wait for hardhat to be up and running")

	details, err := getEthereumDetails()
	assert.Nilf(t, err, "Failed to build Ethereum endpoint")
	assert.NotNilf(t, details, "Ethereum network should not be Nil")

	log.Printf("Deploying contracts ...")
	err = runScriptDeployContracts(details, ctx, getConfigurationProperty("factoryAddress"))
	if err != nil {
		details.Close()
		assert.Nilf(t, err, "Error deploying contracts: %v")
	}

	log.Printf("Registering validators ...")
	validatorAddresses := make([]string, 0)
	knownAccounts := details.GetKnownAccounts()
	for _, acct := range knownAccounts {
		validatorAddresses = append(validatorAddresses, acct.Address.String())
	}
	err = runScriptRegisterValidators(details, validatorAddresses)
	if err != nil {
		details.Close()
		assert.Nilf(t, err, "Error registering validators: %v")
	}
	logger := logging.GetLogger("test").WithField("test", 0)

	log.Printf("Funding accounts ...")
	for _, account := range knownAccounts[1:] {
		watcher := transaction.WatcherFromNetwork(details)
		watcher.StartLoop()

		txn, err := ethereum.TransferEther(details, logger, details.GetDefaultAccount().Address, account.Address, big.NewInt(100000000000000000))
		assert.Nilf(t, err, "Error in TrasferEther transaction")
		assert.NotNilf(t, txn, "Expected transaction not to be nil")
	}
	return details
}

func stopHardHat() error {
	log.Printf("Stopping HardHat running instance ...")
	isRunning, _ := isHardHatRunning()
	if !isRunning {
		return nil
	}

	pid, _ := strconv.Atoi(os.Getenv(hardHatProcessId))
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Error finding HardHat pid: %v", err)
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Printf("Error waiting sending SIGTERM signal to HardHat process: %v", err)
		return err
	}

	_, err = process.Wait()
	if err != nil {
		log.Printf("Error waiting HardHat process to stop: %v", err)
		return err
	}

	log.Printf("HardHat node has been stopped")
	return nil
}

func GetEthereumNetwork(t *testing.T, cleanStart bool) ethereum.Network {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isRunning, _ := isHardHatRunning()
	if !isRunning {
		log.Printf("Hardhat is not running. Start new HardHat")
		details := startHardHat(t, ctx)
		assert.NotNilf(t, details, "Expected details to be not nil")
		return details
	}

	if cleanStart {
		err := stopHardHat()
		assert.Nilf(t, err, "Failed to stopHardHat")

		details := startHardHat(t, ctx)
		assert.NotNilf(t, details, "Expected details to be not nil")
		return details
	}

	network, err := getEthereumDetails()
	assert.Nilf(t, err, "Failed to build Ethereum endpoint")
	assert.NotNilf(t, network, "Ethereum network should not be Nil")

	return network
}

// ========================================================
// ========================================================
// ========================================================
// ========================================================
// ========================================================
// ========================================================
// ========================================================
// ========================================================
// ========================================================

// SetupPrivateKeys computes deterministic private keys for testing
func SetupPrivateKeys(n int) []*ecdsa.PrivateKey {
	if (n < 1) || (n >= 256) {
		panic("invalid number for accounts")
	}
	secp256k1N, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	baseBytes := make([]byte, 32)
	baseBytes[0] = 255
	baseBytes[31] = 255
	privKeyArray := []*ecdsa.PrivateKey{}
	for k := 0; k < n; k++ {
		privKeyBytes := utils.CopySlice(baseBytes)
		privKeyBytes[1] = uint8(k)
		privKeyBig := new(big.Int).SetBytes(privKeyBytes)
		privKeyBig.Mod(privKeyBig, secp256k1N)
		privKeyBytes = privKeyBig.Bytes()
		privKey, err := crypto.ToECDSA(privKeyBytes)
		if err != nil {
			panic(err)
		}
		privKeyArray = append(privKeyArray, privKey)
	}
	return privKeyArray
}

// SetupAccounts derives the associated addresses from private keys
func SetupAccounts(privKeys []*ecdsa.PrivateKey) []accounts.Account {
	accountsArray := []accounts.Account{}
	for _, pk := range privKeys {
		commonAddr := crypto.PubkeyToAddress(pk.PublicKey)
		accountValue := accounts.Account{Address: commonAddr}
		accountsArray = append(accountsArray, accountValue)
	}
	return accountsArray
}

func InitializePrivateKeysAndAccounts(n int) ([]*ecdsa.PrivateKey, []accounts.Account) {
	_, pKey, err := GetOwnerAccount()
	if err != nil {
		panic(err)
	}

	//t.Logf("owner: %v, pvKey: %v", account.Address.String(), key.PrivateKey)
	privateKeys := []*ecdsa.PrivateKey{pKey}
	randomPrivateKeys := SetupPrivateKeys(n - 1)
	privateKeys = append(privateKeys, randomPrivateKeys...)
	accounts := SetupAccounts(privateKeys)

	return privateKeys, accounts
}

// GetConfigurationKeyValueFromFile given a file following the pattern `key = value`, it returns the value of `key`
func GetConfigurationKeyValueFromFile(filePath string, key string) (string, error) {
	rootPath := getProjectRootPath()
	rootPath = append(rootPath, filePath)
	fileFullPath := filepath.Join(rootPath...)

	f, err := os.Open(fileFullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var value string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			value = scanner.Text()
			break
		}
	}

	splits := strings.Split(value, "=")
	return strings.Trim(splits[1], " \""), nil
}

func getConfigurationProperty(key string) string {
	value, err := GetConfigurationKeyValueFromFile("scripts/base-files/baseConfig", key)
	if err != nil {
		return ""
	}
	return value
}

func GetOwnerAccount() (*common.Address, *ecdsa.PrivateKey, error) {
	rootPath := getProjectRootPath()

	// Account
	acctAddress, err := GetConfigurationKeyValueFromFile("scripts/base-files/owner.toml", "defaultAccount")
	if err != nil {
		log.Printf("Error retrieving default account address. %v", err)
		return nil, nil, err
	}
	acctAddressLowerCase := strings.ToLower(acctAddress)

	// Password
	passwordPath := append(rootPath, "scripts")
	passwordPath = append(passwordPath, "base-files")
	passwordPath = append(passwordPath, "passwordFile")
	passwordFullPath := filepath.Join(passwordPath...)

	passwordFileContent, err := ioutil.ReadFile(passwordFullPath)
	if err != nil {
		log.Printf("Error opening password file. %v", err)
		return nil, nil, err
	}
	password := string(passwordFileContent)

	// Wallet
	walletPath := append(rootPath, "scripts")
	walletPath = append(walletPath, "base-files")
	walletPath = append(walletPath, acctAddressLowerCase)
	walletFullPath := filepath.Join(walletPath...)

	jsonBytes, err := ioutil.ReadFile(walletFullPath)
	if err != nil {
		log.Printf("Error opening %v file.  %v", acctAddressLowerCase, err)
		return nil, nil, err
	}

	key, err := keystore.DecryptKey(jsonBytes, password)
	if err != nil {
		log.Printf("Error decrypting jsonBytes. %v", err)
		return nil, nil, err
	}

	return &key.Address, key.PrivateKey, nil
}

//func Setup(finalityDelay uint64, numAccounts int, registryAddress common.Address) (ethereum.Network, *logrus.Logger, error) {
//	logger := logging.GetLogger("test")
//	logger.SetLevel(logrus.TraceLevel)
//	ecdsaPrivateKeys, _ := InitializePrivateKeysAndAccounts(numAccounts)
//	eth, err := ethereum.NewSimulator(
//		ecdsaPrivateKeys,
//		6,
//		10*time.Second,
//		30*time.Second,
//		0,
//		big.NewInt(math.MaxInt64),
//		50,
//		math.MaxInt64)
//	if err != nil {
//		return nil, logger, err
//	}
//
//	eth.SetFinalityDelay(finalityDelay)
//	knownSelectors := transaction.NewKnownSelectors()
//	transaction := transaction.NewWatcher(eth.GetClient(), knownSelectors, 5)
//	transaction.SetNumOfConfirmationBlocks(finalityDelay)
//
//	//todo: redeploy and get the registryAddress here
//	err = eth.Contracts().LookupContracts(context.Background(), registryAddress)
//	if err != nil {
//		return nil, logger, err
//	}
//	return eth, logger, nil
//}

func runScriptHardHatNode() error {

	rootPath := getProjectRootPath()
	scriptPath := getMainScriptPath()

	cmd := exec.Cmd{
		Path: scriptPath,
		Args: []string{scriptPath, scriptStartHardHatNode},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Start()
	if err != nil {
		log.Printf("Could not execute %s script: %v", scriptStartHardHatNode, err)
		return err
	}

	err = os.Setenv(hardHatProcessId, strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		log.Printf("Error setting environment variable: %v", err)
		return err
	}

	return nil
}

func RunScriptInitializeValidator(n int) error {

	rootPath := getProjectRootPath()
	scriptPath := getMainScriptPath()

	cmd := exec.Cmd{
		Path: scriptPath,
		Args: []string{scriptPath, scriptInit, strconv.Itoa(n)},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Start()
	if err != nil {
		log.Printf("Could not execute %s script: %v", scriptInit, err)
		return err
	}

	return nil
}

func runScriptDeployContracts(eth *ethereum.Details, ctx context.Context, factoryAddress string) error {

	rootPath := getProjectRootPath()
	scriptPath := getMainScriptPath()

	err := os.Setenv("SKIP_REGISTRATION", "1")
	if err != nil {
		log.Printf("Error setting environment variable: %v", err)
		return err
	}

	cmd := exec.Cmd{
		Path: scriptPath,
		Args: []string{scriptPath, scriptDeploy},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err = cmd.Run()
	if err != nil {
		log.Printf("Could not execute %s script: %v", scriptDeploy, err)
		return err
	}

	addr := common.Address{}
	copy(addr[:], common.FromHex(factoryAddress))
	eth.Contracts().Initialize(ctx, addr)

	return nil
}

func runScriptRegisterValidators(eth *ethereum.Details, validatorAddresses []string) error {

	rootPath := getProjectRootPath()
	scriptPath := getMainScriptPath()

	args := []string{
		scriptPath,
		scriptRegisterValidator,
		eth.Contracts().ContractFactoryAddress().String(),
	}
	args = append(args, validatorAddresses...)

	cmd := exec.Cmd{
		Path: scriptPath,
		Args: args,
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Run()
	if err != nil {
		log.Printf("Could not execute %s script: %v", scriptRegisterValidator, err)
		return err
	}

	return nil
}

// send a command to the hardhat server via an RPC call
func SendHardhatCommand(command string, params ...interface{}) error {

	commandJson := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  command,
		Params:  make([]byte, 0),
	}

	paramsJson, err := json.Marshal(params)
	if err != nil {
		return err
	}

	commandJson.Params = paramsJson

	c := http.Client{}
	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(commandJson)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buff.Bytes())

	resp, err := c.Post(
		"http://127.0.0.1:8545",
		"application/json",
		reader,
	)

	if err != nil {
		return err
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// mine a certain number of hardhat blocks
func MineBlocks(t *testing.T, eth ethereum.Network, blocksToMine uint64) {
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)
	log.Printf("hardhat_mine %v blocks ", blocksToMine)
	err := SendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// advance to a certain block number
func AdvanceTo(t *testing.T, eth ethereum.Network, target uint64) {
	currentBlock, err := eth.GetCurrentHeight(context.Background())
	if err != nil {
		panic(err)
	}
	if target < currentBlock {
		return
	}
	blocksToMine := target - currentBlock
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)

	log.Printf("hardhat_mine %v blocks to target height %v", blocksToMine, target)

	err = SendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// The the Base fee for the next hardhat block. Can be used to make tx stale.
func SetNextBlockBaseFee(t *testing.T, eth ethereum.Network, target uint64) {
	log.Printf("Setting hardhat_setNextBlockBaseFeePerGas to %v", target)

	err := SendHardhatCommand("hardhat_setNextBlockBaseFeePerGas", "0x"+strconv.FormatUint(target, 16))
	if err != nil {
		panic(err)
	}
}

// Enable/disable hardhat autoMine
func SetAutoMine(t *testing.T, eth ethereum.Network, autoMine bool) {
	log.Printf("Setting Automine to %v", autoMine)

	err := SendHardhatCommand("evm_setAutomine", autoMine)
	if err != nil {
		panic(err)
	}
}

// Set the interval between hardhat blocks. In case interval is 0, we enter in
// manual mode and blocks can only be mined explicitly by calling `MineBlocks`.
// This function disables autoMine.
func SetBlockInterval(t *testing.T, eth ethereum.Network, intervalInMilliSeconds uint64) {
	SetAutoMine(t, eth, false)
	log.Printf("Setting block interval to %v seconds", intervalInMilliSeconds)
	err := SendHardhatCommand("evm_setIntervalMining", intervalInMilliSeconds)
	if err != nil {
		panic(err)
	}
}

// getProjectRootPath returns the project root path
func getProjectRootPath() []string {

	rootPath := []string{string(os.PathSeparator)}

	cmd := exec.Command("go", "list", "-m", "-f", "'{{.Dir}}'", "github.com/MadBase/MadNet")
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting project root path: %v", err)
		return rootPath
	}

	path := string(stdout)
	path = strings.ReplaceAll(path, "'", "")
	path = strings.ReplaceAll(path, "\n", "")

	pathNodes := strings.Split(path, string(os.PathSeparator))
	for _, pathNode := range pathNodes {
		rootPath = append(rootPath, pathNode)
	}

	return rootPath
}

// getMainScriptPath return the path of the main.sh script
func getMainScriptPath() string {
	rootPath := getProjectRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)

	return scriptPathJoined
}

// setCommandStdOut If ENABLE_SCRIPT_LOG env variable is set as 'true' the command will show scripts logs
func setCommandStdOut(cmd *exec.Cmd) {

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
