package gosolc

import (
	"encoding/json"
	"fmt"
	"log"
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

type Compiler struct {
	*CompilerConfig `json:"compilerConfig"`

	Sources       map[string]map[string]string `json:"sources"`
	CompilerInput []byte                       `json:"compilerInput"`
	SolcJs        []byte                       `json:"solcJs"`
}

func NewCompilerConfig(EVMVersion string, OptimizerEnabled bool, OptimizerRuns uint) *CompilerConfig {
	return &CompilerConfig{
		EVMVersion:    EVMVersion,
		SolcOptimizer: &SolcOptimizerConfig{Enabled: OptimizerEnabled, Runs: int(OptimizerRuns)},
	}
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

	c.SolcJs, err = solcJsFromPath(solcJsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read solc-js: %v", err)
	}

	return c, nil

}

func solcJsFromPath(path string) ([]byte, error) {
	solcJS, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read solc-js: %v", err)
	}
	return solcJS, nil
}

func NewDefaultCompiler(contractsDir string) (*Compiler, error) {
	return NewCompiler_0_8_29(contractsDir)
}

func NewCompiler_0_8_29(contractsDir string) (*Compiler, error) {
	c := &Compiler{
		CompilerConfig: defaultConfig,
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

	c.SolcJs = solcJS_0_8_29

	return c, nil
}

func contractsDirToSourcesMap(contractsDir string) (map[string]map[string]string, error) {
	sources := make(map[string]map[string]string)

	err := filepath.Walk(contractsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".sol" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(contractsDir, path)
		sources[relPath] = map[string]string{"content": string(content)}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return sources, nil
}

func (c Compiler) getInputJSON() ([]byte, error) {
	compilerInput := map[string]any{
		"language": "Solidity",
		"sources":  c.Sources,
		"settings": map[string]any{
			"optimizer": map[string]any{
				"enabled": c.CompilerConfig.SolcOptimizer.Enabled,
				"runs":    c.CompilerConfig.SolcOptimizer.Runs,
			},
			"evmVersion": c.CompilerConfig.EVMVersion,

			"outputSelection": map[string]any{
				"*": map[string][]string{
					"*": {"abi", "evm.bytecode"},
				},
			},
		},
	}

	inputJSON, err := json.Marshal(compilerInput)
	return inputJSON, err
}

func (c Compiler) Compile() (map[string]interface{}, error) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	_, err := ctx.RunScript(string(c.SolcJs), "solc.js")
	if err != nil {
		return nil, fmt.Errorf("failed to load solc-js: %w", err)
	}

	compileJS := `
		var solc = globalThis.Module;
		function compile(input) {
			return JSON.stringify(JSON.parse(solc.compile(input)));
		}
	`
	_, err = ctx.RunScript(compileJS, "compile-wrapper.js")
	if err != nil {
		return nil, fmt.Errorf("failed to define compile wrapper: %w", err)
	}

	val, err := ctx.RunScript(fmt.Sprintf(`compile(%q)`, string(c.CompilerInput)), "run-compile.js")
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(val.String()), &output); err != nil {
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

func (c Compiler) writeOutput(contracts map[string]interface{}) error {
	if err := os.MkdirAll("solc-go-build", 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	for _, fileContracts := range contracts {
		for name, data := range fileContracts.(map[string]interface{}) {
			contract := data.(map[string]interface{})
			abi := contract["abi"]
			bytecode := contract["evm"].(map[string]interface{})["bytecode"].(map[string]interface{})["object"]

			out := map[string]interface{}{
				"abi":      abi,
				"bytecode": bytecode,
			}
			outputPath := filepath.Join("solc-go-build", name+".json")
			fileContent, _ := json.MarshalIndent(out, "", "  ")
			if err := os.WriteFile(outputPath, fileContent, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", name, err)
			}
			fmt.Printf("✅ Wrote: %s\n", outputPath)
		}
	}

	return nil
}
