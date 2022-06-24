package cmd

import (
	"strings"
)

func RunRegister(factoryAddress string, validators []string) error {

	bridgeDir := GetBridgePath()

	// iterate on validators and apass it as string array

	//validatorAddresses := make([]string, 0)
	//for _, validatorAddress := range validators {
	//	validatorAddresses = append(validatorAddresses, validatorAddress.String())
	//}
	//validatorAddresses = append(validatorAddresses, "0x61Ae54Fb4DB3d5b0f43Bd24553f69262c5Bc174d")
	//validatorAddresses = append(validatorAddresses, "0x671496a9eb9cd271c05A07Bb71d52656e3c57817")
	//validatorAddresses = append(validatorAddresses, "0x913cFad222B2152D5781Aae072113160eA3891Ab")
	//validatorAddresses = append(validatorAddresses, "0xE3179f6517f1e5752af5C67Ba0259fA883A315E3")

	// Register validator
	_, _, err := executeCommand(bridgeDir, "npx", "hardhat --network dev --show-stack-traces registerValidators --factory-address", factoryAddress, strings.Join(validators, " "))
	if err != nil {
		return err
	}

	return nil
}
