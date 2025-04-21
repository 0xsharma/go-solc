package gosolc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func solcJsFromPath(path string) ([]byte, error) {
	solcJS, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read solc-js: %v", err)
	}
	return solcJS, nil
}

func contractsDirToSourcesMap(contractsDir string) (map[string]map[string]string, error) {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.sol", contractsDir))
	if err != nil {
		return nil, fmt.Errorf("failed to read contracts directory: %v", err)
	}

	sources := make(map[string]map[string]string)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file, err)
		}
		// Escape the content by marshaling it to JSON and removing quotes
		escapedContent, err := json.Marshal(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to marshal file content: %v", err)
		}
		// Remove surrounding quotes from marshaled string
		escapedContentStr := string(escapedContent)[1 : len(escapedContent)-1]
		sources[filepath.Base(file)] = map[string]string{"content": escapedContentStr}
	}

	return sources, nil
}

func (c Compiler) getInputJSON() (string, error) {
	compilerInput := map[string]interface{}{
		"language": "Solidity",
		"sources":  c.Sources,
		"settings": map[string]interface{}{
			"optimizer": map[string]any{
				"enabled": c.CompilerConfig.SolcOptimizer.Enabled,
				"runs":    c.CompilerConfig.SolcOptimizer.Runs,
			},
			"evmVersion": c.CompilerConfig.EVMVersion,
			"outputSelection": map[string]map[string][]string{
				"*": {
					"*": []string{
						"abi",
						"evm.bytecode.object",
						"evm.bytecode.sourceMap",
						"evm.deployedBytecode.object",
						"evm.deployedBytecode.sourceMap",
						"evm.methodIdentifiers",
					},
					"": []string{
						"ast",
					},
				},
			},
		},
	}

	inputJSON, err := json.Marshal(compilerInput)
	if err != nil {
		return "", fmt.Errorf("failed to marshal compiler input JSON: %v", err)
	}

	inputJSONStr := strings.ReplaceAll(string(inputJSON), `'`, `\'`)

	return inputJSONStr, err
}
