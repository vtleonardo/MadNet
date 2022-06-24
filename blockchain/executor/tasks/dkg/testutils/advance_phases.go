//go:build integration

package testutils

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/blockchain/monitor/events"
	testutils "github.com/MadBase/MadNet/blockchain/testutils"
	"github.com/MadBase/MadNet/blockchain/testutils/cmd"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strings"
	"testing"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	Eth              ethereum.Network
	DKGStates        []*state.DkgState
	ecdsaPrivateKeys []*ecdsa.PrivateKey

	regTasks                     []*dkg.RegisterTask
	DispMissingRegTasks          []*dkg.DisputeMissingRegistrationTask
	ShareDistTasks               []*dkg.ShareDistributionTask
	DisputeMissingShareDistTasks []*dkg.DisputeMissingShareDistributionTask
	DisputeShareDistTasks        []*dkg.DisputeShareDistributionTask
	KeyshareSubmissionTasks      []*dkg.KeyShareSubmissionTask
	DisputeMissingKeyshareTasks  []*dkg.DisputeMissingKeySharesTask
	MpkSubmissionTasks           []*dkg.MPKSubmissionTask
	GpkjSubmissionTasks          []*dkg.GPKjSubmissionTask
	DisputeMissingGPKjTasks      []*dkg.DisputeMissingGPKjTask
	DisputeGPKjTasks             []*dkg.DisputeGPKjTask
	CompletionTasks              []*dkg.CompletionTask
}

func SetETHDKGPhaseLength(length uint16, eth ethereum.Network, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {

	// Shorten ethdkg phase for testing purposes
	ethdkgABI, err := abi.JSON(strings.NewReader(bindings.ETHDKGMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := ethdkgABI.Pack("setPhaseLength", uint16(length))
	if err != nil {
		return nil, nil, err
	}

	txn, err := eth.Contracts().ContractFactory().CallAny(callOpts, eth.Contracts().EthdkgAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(ethdkg, setPhaseLength(...))")
	}

	watcher := transaction.WatcherFromNetwork(eth)
	c, err := watcher.Subscribe(ctx, txn)
	testutils.MineBlocks(eth, 12)
	rcpt, err := watcher.Wait(ctx, c)

	if err != nil {
		return nil, nil, err
	}
	if rcpt == nil {
		return nil, nil, errors.New("non existent receipt for tx ContractFactory.CallAny(ethdkg, setPhaseLength(...))")
	}
	return txn, rcpt, nil
}

func InitializeETHDKG(eth ethereum.Network, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {
	// Shorten ethdkg phase for testing purposes
	validatorPoolABI, err := abi.JSON(strings.NewReader(bindings.ValidatorPoolMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := validatorPoolABI.Pack("initializeETHDKG")
	if err != nil {
		return nil, nil, err
	}

	txn, err := eth.Contracts().ContractFactory().CallAny(callOpts, eth.Contracts().ValidatorPoolAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(validatorPool, initializeETHDKG())")
	}

	watcher := transaction.WatcherFromNetwork(eth)
	c, err := watcher.Subscribe(ctx, txn)
	testutils.MineBlocks(eth, 12)
	rcpt, err := watcher.Wait(ctx, c)

	if err != nil {
		return nil, nil, err
	}
	if rcpt == nil {
		return nil, nil, errors.New("non existent receipt for tx ContractFactory.CallAny(validatorPool, initializeETHDKG())")
	}

	return txn, rcpt, nil
}

func StartFromRegistrationOpenPhase(t *testing.T, n int, unregisteredValidators int, phaseLength uint16, workingDir string) *TestSuite {
	//ecdsaPrivateKeys, accountList := testutils.InitializePrivateKeysAndAccounts(n)
	log.Printf("********** 0 - Testing with %d validators", n)

	accountMap := testutils.InitializePrivateKeysAndAccountsMap(n)

	log.Printf("*** running init")
	err := cmd.RunInit(workingDir, n, accountMap)
	assert.Nil(t, err)

	//eth := testutils.GetEthereumNetwork(t, ecdsaPrivateKeys, 1000*time.Millisecond)
	//eth := testutils.GetEthereumNetwork(t, false, n, workingDir)
	eth := testutils.GetEthereumNetwork(t, false, n, workingDir, accountMap)
	assert.NotNil(t, eth)
	log.Printf("********** 8 - Got ETH network")
	log.Printf("********** 9 - eth default account (owner) address: %s", eth.GetDefaultAccount().Address.String())

	ctx := context.Background()
	var owner = eth.GetDefaultAccount()

	accountList := make([]accounts.Account, 0)
	ecdsaPrivateKeys := make([]*ecdsa.PrivateKey, 0)
	for a, k := range accountMap {
		accountList = append(accountList, a)
		ecdsaPrivateKeys = append(ecdsaPrivateKeys, k)
	}

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)
	log.Printf("********** 10 - Owner opts %v", ownerOpts)

	log.Printf("********** 11 - at this point account map looks like this")
	for v := range accountMap {
		log.Printf("   %s", v.Address.String())
	}
	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(phaseLength, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// init ETHDKG on ValidatorPool, through ContractFactory
	_, rcpt, err := InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	validatorAddresses, err := utils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	assert.Nil(t, err)

	phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(state.RegistrationOpen), phase)

	// TODO - remove owner
	//valCount, err := eth.Contracts().ValidatorPool().GetValidatorsCount(callOpts)
	//assert.Nil(t, err)
	//assert.Equal(t, uint64(n), valCount.Uint64())

	// Do Register task
	regTasks := make([]*dkg.RegisterTask, n)
	dispMissingRegTasks := make([]*dkg.DisputeMissingRegistrationTask, n)
	dkgStates := make([]*state.DkgState, n)
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accountList[idx].Address.String())
		// Set Registration success to true
		state, regTask, dispMissingRegTask := events.UpdateStateOnRegistrationOpened(
			accountList[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64(),
			true,
			validatorAddresses,
		)

		dkgStates[idx] = state
		regTasks[idx] = regTask
		dispMissingRegTasks[idx] = dispMissingRegTask

		err = regTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)

		if idx >= n-unregisteredValidators {
			continue
		}

		nVal, err := eth.Contracts().Ethdkg().GetNumParticipants(callOpts)
		assert.Nil(t, err)
		assert.Equal(t, uint64(idx), nVal.Uint64())

		err = regTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)
		assert.True(t, regTasks[idx].Success)
	}

	// simulate receiving AddressRegistered event
	for i := 0; i < n; i++ {
		state := dkgStates[i]

		if i >= n-unregisteredValidators {
			continue
		}

		for j := 0; j < n; j++ {
			dkgStates[j].OnAddressRegistered(state.Account.Address, i+1, state.Nonce, state.TransportPublicKey)
		}
	}

	shareDistributionTasks := make([]*dkg.ShareDistributionTask, n)
	disputeMissingShareDistributionTasks := make([]*dkg.DisputeMissingShareDistributionTask, n)
	disputeShareDistTasks := make([]*dkg.DisputeShareDistributionTask, n)

	if unregisteredValidators == 0 {
		height, err := eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		for idx := 0; idx < n; idx++ {
			shareDistributionTask, disputeMissingShareDistributionTask, disputeShareDistTask := events.UpdateStateOnRegistrationComplete(dkgStates[idx], height)

			shareDistributionTasks[idx] = shareDistributionTask
			disputeMissingShareDistributionTasks[idx] = disputeMissingShareDistributionTask
			disputeShareDistTasks[idx] = disputeShareDistTask
		}

		// skip all the way to ShareDistribution phase
		testutils.AdvanceTo(eth, shareDistributionTasks[0].Start)
	} else {
		// this means some validators did not register, and the next phase is DisputeMissingRegistration
		testutils.AdvanceTo(eth, dkgStates[0].PhaseStart+dkgStates[0].PhaseLength)
	}

	return &TestSuite{
		Eth:                          eth,
		DKGStates:                    dkgStates,
		ecdsaPrivateKeys:             ecdsaPrivateKeys,
		regTasks:                     regTasks,
		DispMissingRegTasks:          dispMissingRegTasks,
		ShareDistTasks:               shareDistributionTasks,
		DisputeMissingShareDistTasks: disputeMissingShareDistributionTasks,
		DisputeShareDistTasks:        disputeShareDistTasks,
	}
}

func StartFromShareDistributionPhase(t *testing.T, n int, undistributedSharesIdx []int, badSharesIdx []int, phaseLength uint16, workingDir string) *TestSuite {
	suite := StartFromRegistrationOpenPhase(t, n, 0, phaseLength, workingDir)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	callOpts, err := suite.Eth.GetCallOpts(ctx, suite.Eth.GetDefaultAccount())
	assert.Nil(t, err)
	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, phase, uint8(state.ShareDistribution))

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, height, suite.ShareDistTasks[0].Start)

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]

		var skipLoop = false

		for _, undistIdx := range undistributedSharesIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		shareDistTask := suite.ShareDistTasks[idx]

		err := shareDistTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		for _, badIdx := range badSharesIdx {
			if idx == badIdx {
				// inject bad shares
				for _, s := range state.Participants[state.Account.Address].EncryptedShares {
					s.Set(big.NewInt(0))
				}
			}
		}

		err = shareDistTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		assert.True(t, shareDistTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			err = suite.DKGStates[j].OnSharesDistributed(
				logger,
				state.Account.Address,
				state.Participants[state.Account.Address].EncryptedShares,
				state.Participants[state.Account.Address].Commitments,
			)
			assert.Nil(t, err)
		}

	}

	disputeShareDistributionTasks := make([]*dkg.DisputeShareDistributionTask, n)
	keyshareSubmissionTasks := make([]*dkg.KeyShareSubmissionTask, n)
	disputeMissingKeySharesTasks := make([]*dkg.DisputeMissingKeySharesTask, n)

	if len(undistributedSharesIdx) == 0 {
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispShareDistStartBlock uint64

		// this means all validators distributed their shares and now the phase is
		// set phase to DisputeShareDistribution
		for i := 0; i < n; i++ {
			disputeShareDistributionTask, keyshareSubmissionTask, disputeMissingKeySharesTask := events.UpdateStateOnShareDistributionComplete(suite.DKGStates[i], height)

			dispShareDistStartBlock = disputeShareDistributionTask.GetStart()

			disputeShareDistributionTasks[i] = disputeShareDistributionTask
			keyshareSubmissionTasks[i] = keyshareSubmissionTask
			disputeMissingKeySharesTasks[i] = disputeMissingKeySharesTask
		}

		suite.DisputeShareDistTasks = disputeShareDistributionTasks
		suite.KeyshareSubmissionTasks = keyshareSubmissionTasks
		suite.DisputeMissingKeyshareTasks = disputeMissingKeySharesTasks

		// skip all the way to DisputeShareDistribution phase
		testutils.AdvanceTo(suite.Eth, dispShareDistStartBlock)
	} else {
		// this means some validators did not distribute shares, and the next phase is DisputeMissingShareDistribution
		testutils.AdvanceTo(suite.Eth, suite.DKGStates[0].PhaseStart+suite.DKGStates[0].PhaseLength)
	}

	return suite
}

func StartFromKeyShareSubmissionPhase(t *testing.T, n int, undistributedShares int, phaseLength uint16, workingDir string) *TestSuite {
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength, workingDir)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	keyshareSubmissionStartBlock := suite.KeyshareSubmissionTasks[0].Start
	testutils.AdvanceTo(suite.Eth, keyshareSubmissionStartBlock)

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]

		if idx >= n-undistributedShares {
			continue
		}

		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		err = keyshareSubmissionTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			suite.DKGStates[j].OnKeyShareSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	mpkSubmissionTasks := make([]*dkg.MPKSubmissionTask, n)

	if undistributedShares == 0 {
		// at this point all the validators submitted their key shares
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		// this means all validators submitted their respective key shares and now the phase is
		// set phase to MPK
		var mpkSubmissionTaskStart uint64
		for i := 0; i < n; i++ {
			mpkSubmissionTask := events.UpdateStateOnKeyShareSubmissionComplete(suite.DKGStates[i], height)
			mpkSubmissionTaskStart = mpkSubmissionTask.GetStart()

			mpkSubmissionTasks[i] = mpkSubmissionTask
		}

		// skip all the way to MPKSubmission phase
		testutils.AdvanceTo(suite.Eth, mpkSubmissionTaskStart)
	} else {
		// this means some validators did not submit key shares, and the next phase is DisputeMissingKeyShares
		testutils.AdvanceTo(suite.Eth, suite.DKGStates[0].PhaseStart+suite.DKGStates[0].PhaseLength)
	}

	suite.MpkSubmissionTasks = mpkSubmissionTasks

	return suite
}

func StartFromMPKSubmissionPhase(t *testing.T, n int, phaseLength uint16, workingDir string) *TestSuite {
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, phaseLength, workingDir)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	dkgStates := suite.DKGStates
	eth := suite.Eth

	// Do MPK Submission task (once is enough)

	for idx := 0; idx < n; idx++ {
		task := suite.MpkSubmissionTasks[idx]
		state := dkgStates[idx]
		err := task.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		if task.AmILeading(ctx, eth, logger, state) {
			err = task.DoWork(ctx, logger, eth)
			assert.Nil(t, err)
		}
	}

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	gpkjSubmissionTasks := make([]*dkg.GPKjSubmissionTask, n)
	disputeMissingGPKjTasks := make([]*dkg.DisputeMissingGPKjTask, n)
	disputeGPKjTasks := make([]*dkg.DisputeGPKjTask, n)

	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask := events.UpdateStateOnMPKSet(state, height, new(adminHandlerMock))

		gpkjSubmissionTasks[idx] = gpkjSubmissionTask
		disputeMissingGPKjTasks[idx] = disputeMissingGPKjTask
		disputeGPKjTasks[idx] = disputeGPKjTask
	}

	suite.GpkjSubmissionTasks = gpkjSubmissionTasks
	suite.DisputeMissingGPKjTasks = disputeMissingGPKjTasks
	suite.DisputeGPKjTasks = disputeGPKjTasks

	return suite
}

func StartFromGPKjPhase(t *testing.T, n int, undistributedGPKjIdx []int, badGPKjIdx []int, phaseLength uint16, workingDir string) *TestSuite {
	suite := StartFromMPKSubmissionPhase(t, n, phaseLength, workingDir)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]

		var skipLoop = false

		for _, undistIdx := range undistributedGPKjIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		gpkjSubTask := suite.GpkjSubmissionTasks[idx]

		err := gpkjSubTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		for _, badIdx := range badGPKjIdx {
			if idx == badIdx {
				// inject bad shares
				// mess up with group private key (gskj)
				gskjBad := new(big.Int).Add(state.GroupPrivateKey, big.NewInt(1))
				// here's the group public key
				gpkj := new(cloudflare.G2).ScalarBaseMult(gskjBad)
				gpkjBad, err := bn256.G2ToBigIntArray(gpkj)
				assert.Nil(t, err)

				state.GroupPrivateKey = gskjBad
				state.Participants[state.Account.Address].GPKj = gpkjBad
			}
		}

		err = gpkjSubTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		assert.True(t, gpkjSubTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			suite.DKGStates[j].OnGPKjSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].GPKj,
			)
		}

	}

	disputeGPKjTasks := make([]*dkg.DisputeGPKjTask, n)
	completionTasks := make([]*dkg.CompletionTask, n)

	if len(undistributedGPKjIdx) == 0 {
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispGPKjStartBlock uint64

		// this means all validators submitted their GPKjs and now the phase is
		// set phase to DisputeGPKjDistribution
		for i := 0; i < n; i++ {
			disputeGPKjTask, completionTask := events.UpdateStateOnGPKJSubmissionComplete(suite.DKGStates[i], height)

			dispGPKjStartBlock = disputeGPKjTask.GetStart()

			disputeGPKjTasks[i] = disputeGPKjTask
			completionTasks[i] = completionTask
		}

		suite.DisputeGPKjTasks = disputeGPKjTasks
		suite.CompletionTasks = completionTasks

		// skip all the way to DisputeGPKj phase
		testutils.AdvanceTo(suite.Eth, dispGPKjStartBlock)
	} else {
		// this means some validators did not submit their GPKjs, and the next phase is DisputeMissingGPKj
		testutils.AdvanceTo(suite.Eth, suite.DKGStates[0].PhaseStart+suite.DKGStates[0].PhaseLength)
	}

	return suite
}

func StartFromCompletion(t *testing.T, n int, phaseLength uint16, workingDir string) *TestSuite {
	suite := StartFromGPKjPhase(t, n, []int{}, []int{}, phaseLength, workingDir)

	// move to Completion phase
	testutils.AdvanceTo(suite.Eth, suite.CompletionTasks[0].Start+suite.DKGStates[0].ConfirmationLength)

	return suite
}
