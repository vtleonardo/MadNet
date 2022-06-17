package cmd

import (
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func RunRegister(factoryAddress string, validators []common.Address) error {

	bridgeDir := GetBridgePath()

	// iterate on validators and apass it as string array

	validatorAddresses := make([]string, 0)
	for _, validatorAddress := range validators {
		validatorAddresses = append(validatorAddresses, validatorAddress.String())
	}

	// Register validator
	_, _, err := executeCommand(bridgeDir, "npx", "hardhat --network dev --show-stack-traces registerValidators --factory-address", factoryAddress, strings.Join(validatorAddresses, " "))
	if err != nil {
		return err
	}

	return nil
}
