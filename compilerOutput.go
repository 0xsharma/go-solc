package gosolc

import "fmt"

type CompilerOutput map[string]interface{}

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
