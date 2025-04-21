package gosolc

import "fmt"

// CompilerOutput represents the output of the Solidity compiler.
// It is a map where the keys are file names and the values are maps of contract names to their respective output data.
// The output data includes information such as ABI, bytecode, and source maps.
type CompilerOutput map[string]interface{}

// GetContractByteCodes retrieves the bytecode for each contract in the compiler output.
// It returns a map where the keys are contract names and the values are their respective bytecode strings.
func (contracts CompilerOutput) GetContractByteCodes() (map[string]string, error) {
	contractByteCodes := make(map[string]string)

	for _, fileContracts := range contracts {
		for name, contract := range fileContracts.(map[string]interface{}) {
			evm, ok := contract.(map[string]interface{})["evm"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid evm output for contract %s", name)
			}
			bytecode, ok := evm["bytecode"].(map[string]interface{})["object"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid bytecode output for contract %s", name)
			}
			contractByteCodes[name] = bytecode
		}
	}

	return contractByteCodes, nil
}

// GetDeployedContractByteCodes retrieves the deployed bytecode for each contract in the compiler output.
// It returns a map where the keys are contract names and the values are their respective deployed bytecode strings.
// This is useful for verifying the deployed contract on the blockchain.
func (contracts CompilerOutput) GetDeployedContractByteCodes() (map[string]string, error) {
	contractByteCodes := make(map[string]string)

	for _, fileContracts := range contracts {
		for name, contract := range fileContracts.(map[string]interface{}) {
			evm, ok := contract.(map[string]interface{})["evm"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid evm output for contract %s", name)
			}
			deployedBytecode, ok := evm["deployedBytecode"].(map[string]interface{})["object"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid deployed bytecode output for contract %s", name)
			}
			contractByteCodes[name] = deployedBytecode
		}
	}

	return contractByteCodes, nil
}
