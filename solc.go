package gosolc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"rogchap.com/v8go"
)

type SolcOptimizerConfig struct {
	Enabled bool `json:"enabled"`
	Runs    int  `json:"runs"`
}

type CompilerConfig struct {
	EVMVersion    string               `json:"evmVersion"`
	SolcOptimizer *SolcOptimizerConfig `json:"optimizer"`
}

var defaultConfig = &CompilerConfig{
	EVMVersion:    "cancun",
	SolcOptimizer: &SolcOptimizerConfig{Enabled: false, Runs: 0},
}

func NewCompilerConfig(EVMVersion string, OptimizerEnabled bool, OptimizerRuns uint) *CompilerConfig {
	return &CompilerConfig{
		EVMVersion:    EVMVersion,
		SolcOptimizer: &SolcOptimizerConfig{Enabled: OptimizerEnabled, Runs: int(OptimizerRuns)},
	}
}

type Compiler struct {
	*CompilerConfig `json:"compilerConfig"`

	Sources       map[string]map[string]string `json:"sources"`
	CompilerInput string                       `json:"compilerInput"`
	SolcJs        string                       `json:"solcJs"`
}

func NewCompiler(contractsDir string, config *CompilerConfig, solcJsPath string) (*Compiler, error) {
	c := &Compiler{
		CompilerConfig: config,
	}

	var err error
	c.Sources, err = contractsDirToSourcesMap(contractsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read contracts directory: %v", err)
	}

	c.CompilerInput, err = c.getInputJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to get input JSON: %v", err)
	}

	if solcJsPath != "" {
		solcJSBytes, err := solcJsFromPath(solcJsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read solc-js: %v", err)
		}

		c.SolcJs = string(solcJSBytes)
	} else {
		c.SolcJs = solcJS_0_8_29
	}

	return c, nil

}

func NewCompiler_0_8_29(contractsDir string) (*Compiler, error) {
	return NewCompiler(contractsDir, defaultConfig, "")
}

func NewDefaultCompiler(contractsDir string) (*Compiler, error) {
	return NewCompiler_0_8_29(contractsDir)
}

func (c Compiler) Compile() (CompilerOutput, error) {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	script := fmt.Sprintf(wrapperScript, c.SolcJs)
	_, err := ctx.RunScript(script, "soljson_wrapper.js")
	if err != nil {
		return nil, fmt.Errorf("failed to load solc-js: %w", err)
	}

	compileScript := fmt.Sprintf(`solc.compile('%s', '')`, c.CompilerInput)

	outputVal, err := ctx.RunScript(compileScript, "compile.js")
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}
	outputStr := outputVal.String()

	var output map[string]interface{}
	err = json.Unmarshal([]byte(outputStr), &output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compiler output: %v", err)
	}

	contracts, ok := output["contracts"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid contracts output")
	}

	return contracts, nil
}

func (c Compiler) CompileAndWriteOutput() error {
	contracts, err := c.Compile()
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	if err := c.writeOutput(contracts); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}

func (c Compiler) writeOutput(contracts CompilerOutput) error {
	outputDir := "solc-go-build"
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, fileContracts := range contracts {
		for name, contract := range fileContracts.(map[string]interface{}) {
			// Marshal contract data to JSON
			contractJSON, err := json.MarshalIndent(contract, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal contract data to JSON: %w", err)
			}
			// Write to file
			outputFile := filepath.Join(outputDir, name+".json")
			err = os.WriteFile(outputFile, contractJSON, 0644)
			if err != nil {
				return fmt.Errorf("failed to write contract data to file: %w", err)
			}
		}
	}

	return nil
}
