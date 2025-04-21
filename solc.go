package gosolc

import (
	"encoding/json"
	"fmt"

	"rogchap.com/v8go"
)

// CompilerOutput is a map of contract names to their compiled output
type SolcOptimizerConfig struct {
	Enabled bool `json:"enabled"` // Indicates if the optimizer is enabled
	Runs    int  `json:"runs"`    // Number of optimization runs
}

// CompilerOutput is a map of contract names to their compiled output
type CompilerConfig struct {
	EVMVersion    string               `json:"evmVersion"` // EVM version to use for compilation
	SolcOptimizer *SolcOptimizerConfig `json:"optimizer"`  // Optimizer configuration
}

// defaultConfig is the default compiler configuration
var defaultConfig = &CompilerConfig{
	EVMVersion:    "cancun",                                      // Default EVM version
	SolcOptimizer: &SolcOptimizerConfig{Enabled: false, Runs: 0}, // Default optimizer configuration
}

// NewCompilerConfig creates a new CompilerConfig with the specified EVM version and optimizer settings.
// Supported EVMVersions: "osaka", "cancun", "shanghai", "paris", "london", "berlin", "istanbul", "petersburg", "constantinople", "byzantium", "homestead", "spuriousDragon", "tangerineWhistle", "frontier"
func NewCompilerConfig(EVMVersion string, OptimizerEnabled bool, OptimizerRuns uint) *CompilerConfig {
	return &CompilerConfig{
		EVMVersion:    EVMVersion,
		SolcOptimizer: &SolcOptimizerConfig{Enabled: OptimizerEnabled, Runs: int(OptimizerRuns)},
	}
}

// Compiler is a struct that holds the configuration and sources for the Solidity compiler
// It includes methods to compile the sources and write the output to files
// It uses the v8go library to run the solc-js compiler in a JavaScript environment
type Compiler struct {
	*CompilerConfig `json:"compilerConfig"`

	Sources       map[string]map[string]string `json:"sources"`
	CompilerInput string                       `json:"compilerInput"`
	SolcJs        string                       `json:"solcJs"`
}

// NewCompiler creates a new Compiler instance with the specified contracts directory and configuration
// contractsDir: The directory containing the Solidity contracts
// config: The compiler configuration ( generated from gosolc.NewCompilerConfig() )
// solcJsPath: The path to the solc-js file (optional). If not provided, a default version will be used
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

// NewCompiler_0_8_29 creates a new Compiler instance with the specified contracts directory, default configuration
// and the solc-js version (0.8.29).
func NewCompiler_0_8_29(contractsDir string) (*Compiler, error) {
	return NewCompiler(contractsDir, defaultConfig, "")
}

// NewDefaultCompiler creates a new Compiler instance with the specified contracts directory and default configuration
// and the solc-js version (0.8.29) as of now
func NewDefaultCompiler(contractsDir string) (*Compiler, error) {
	return NewCompiler_0_8_29(contractsDir)
}

// Compile() compiles the Solidity contracts using the solc-js compiler
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

// CompileAndWriteOutput compiles the Solidity contracts and writes the output to files in ./solc-go-build
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
