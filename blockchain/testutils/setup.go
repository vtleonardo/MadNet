package testutils

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/testutils/cmd"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"testing"
)

var (
	configEndpoint      = "http://localhost:8545"
	ownerAccountAddress = "0x546f99f244b7b58b855330ae0e2bc1b30b41302f"
	password            = "abc123"
	configFinalityDelay = uint64(1)
)

func getEthereumDetails(accounts []accounts.Account) (*ethereum.Details, error) {

	//root := cmd.GetProjectRootPath()
	//assetKey := filepath.Join(root, "assets", "test", "keys")
	//assetPasscode := filepath.Join(root, "assets", "test", "passcodes.txt")
	//assetKey := filepath.Join(workingDir, "scripts", "generated", "keystores", "keys")
	//assetPasscode := filepath.Join(workingDir, "scripts", "generated", "keystores", "passcodes.txt")

	// TODO - use mock
	details, err := ethereum.NewEndpointWithAccount(
		configEndpoint,
		accounts,
		configFinalityDelay,
		500,
		0,
	)
	return details, err
}

func startHardHat(t *testing.T, ctx context.Context, validatorsCount int, workingDir string, accounts []accounts.Account) *ethereum.Details {

	log.Printf("Starting HardHat ...")
	err := cmd.RunHardHatNode()
	assert.Nilf(t, err, "Error starting hardhat node")

	err = cmd.WaitForHardHatNode(ctx)
	assert.Nilf(t, err, "Failed to wait for hardhat to be up and running")

	details, err := getEthereumDetails(accounts)
	assert.Nilf(t, err, "Failed to build Ethereum endpoint")
	assert.NotNilf(t, details, "Ethereum network should not be Nil")

	log.Printf("Deploying contracts ...")
	factoryAddress, err := cmd.RunDeploy(workingDir)
	if err != nil {
		details.Close()
		assert.Nilf(t, err, "Error deploying contracts: %v", err)
		return nil
	}
	addr := common.Address{}
	copy(addr[:], common.FromHex(factoryAddress))
	details.Contracts().Initialize(ctx, addr)

	// TODO - do I need this?
	validatorAddresses := make([]string, 0)
	knownAccounts := details.GetKnownAccounts()
	for _, acct := range knownAccounts[:validatorsCount] {
		validatorAddresses = append(validatorAddresses, acct.Address.String())
	}

	log.Printf("Registering %d validators ...", len(validatorAddresses))
	err = cmd.RunRegister(factoryAddress, nil)
	if err != nil {
		details.Close()
		assert.Nilf(t, err, "Error registering validators: %v")
	}
	//logger := logging.GetLogger("test").WithField("test", 0)

	//log.Printf("Funding accounts ...")
	//for _, account := range knownAccounts[1:] {
	//	txn, err := ethereum.TransferEther(details, logger, details.GetDefaultAccount().Address, account.Address, big.NewInt(100000000000000000))
	//	assert.Nilf(t, err, "Error in TrasferEther transaction")
	//	assert.NotNilf(t, txn, "Expected transaction not to be nil")
	//}
	return details
}

func GetEthereumNetwork(t *testing.T, cleanStart bool, validatorsCount int, workingDir string, accounts []accounts.Account) ethereum.Network {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isRunning, _ := cmd.IsHardHatRunning()
	if !isRunning {
		log.Printf("Hardhat is not running. Start new HardHat")
		details := startHardHat(t, ctx, validatorsCount, workingDir, accounts)
		assert.NotNilf(t, details, "Expected details to be not nil")
		return details
	}

	if cleanStart {
		err := cmd.StopHardHat()
		assert.Nilf(t, err, "Failed to stopHardHat")

		details := startHardHat(t, ctx, validatorsCount, workingDir, accounts)
		assert.NotNilf(t, details, "Expected details to be not nil")
		return details
	}

	//network, err := getEthereumDetails(workingDir)
	network, err := getEthereumDetails(accounts)
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

// GeneratePrivateKeys computes deterministic private keys for testing
func GeneratePrivateKeys(n int) []*ecdsa.PrivateKey {
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

// GenerateAccounts derives the associated addresses from private keys
func GenerateAccounts(privKeys []*ecdsa.PrivateKey) []accounts.Account {
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
		// TODO - don't think panic is the right solution here
		panic(err)
	}

	privateKeys := []*ecdsa.PrivateKey{pKey}
	randomPrivateKeys := GeneratePrivateKeys(n - 1)
	privateKeys = append(privateKeys, randomPrivateKeys...)
	accounts := GenerateAccounts(privateKeys)

	return privateKeys, accounts
}

func GetOwnerAccount() (*common.Address, *ecdsa.PrivateKey, error) {

	id, _ := uuid.Parse("6b2a0716-b444-46c3-a1e3-2936ddd8ecc5")
	pk, _ := crypto.HexToECDSA("6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1")
	addr := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
	key := &keystore.Key{
		Id:         id,
		Address:    addr,
		PrivateKey: pk,
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

func Init(workingDir string, n int) error {

	err := cmd.RunInit(workingDir, n)
	if err != nil {
		log.Fatal("----- INIT FAILED ----")
		return err
	}

	// TODO - change this to be global
	//SetCommandStdOut(&cmd)
	//err = cmd.Start()
	//if err != nil {
	//	return err
	//}

	return nil
}

// TODO - check this one
// sendHardhatCommand sends a command to the hardhat server via an RPC call
func sendHardhatCommand(command string, params ...interface{}) error {

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

	body := bytes.NewReader(buff.Bytes())
	resp, err := c.Post(
		configEndpoint,
		"application/json",
		body,
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

// MineBlocks mines a certain number of hardhat blocks
func MineBlocks(eth ethereum.Network, blocksToMine uint64) {
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)
	log.Printf("hardhat_mine %v blocks ", blocksToMine)
	err := sendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// advance to a certain block number
func AdvanceTo(eth ethereum.Network, target uint64) {
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

	err = sendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// SetNextBlockBaseFee The Base fee for the next hardhat block. Can be used to make tx stale.
func SetNextBlockBaseFee(target uint64) {
	log.Printf("Setting hardhat_setNextBlockBaseFeePerGas to %v", target)
	err := sendHardhatCommand("hardhat_setNextBlockBaseFeePerGas", "0x"+strconv.FormatUint(target, 16))
	if err != nil {
		panic(err)
	}
}

// Enable/disable hardhat autoMine TODO - check this one
func SetAutoMine(t *testing.T, eth ethereum.Network, autoMine bool) {
	log.Printf("Setting Automine to %v", autoMine)

	err := sendHardhatCommand("evm_setAutomine", autoMine)
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
	err := sendHardhatCommand("evm_setIntervalMining", intervalInMilliSeconds)
	if err != nil {
		panic(err)
	}
}
