# gosolc

## Table of Contents

- [About](#about)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installing](#installing)
- [Usage](#usage)
  - [New Compiler](#new-compiler)
  - [New compiler with config (evm version, optimization, and optimization runs)](#new-compiler-with-config-evm-version-optimization-and-optimization-runs)
  - [Compile contracts](#compile-contracts)
  - [Compile and write ABI and Bytecode to JSON files](#compile-and-write-abi-and-bytecode-to-json-files)
  - [Get Bytecodes from compiler output](#get-bytecodes-from-compiler-output)
- [Contributing](#contributing)


## About <a name = "about"></a>

A Go package that compiles Solidity smart contracts using soljson.js via the embedded V8 engine (v8go). It reads .sol files from a contracts directory, handles imports and dependencies, and outputs ABI and bytecode as JSON files for Ethereum development.

## Features <a name = "features"></a>
- Compiles multiple Solidity files from the contracts directory.
- Supports standard JSON input/output format for solc.
- Extracts and saves ABI, bytecode, and deployed bytecode for each contract.
- Handles imports (e.g., import "./dummy_ERC20.sol").

## Prerequisites <a name = "prerequisites"></a>

What things you need to install the software and how to install them.

```
Go: Version 1.18 or higher.
```

## Installing <a name = "installing"></a>
```
go get -u "github.com/0xsharma/gosolc"
```

## Usage <a name = "usage"></a>

### New Compiler
```go
import (
	"github.com/0xsharma/gosolc"
)

c, err := gosolc.NewDefaultCompiler("./contracts")
if err != nil {
    //handle error
}
```
or 

### New compiler with config (evm version, optimization, optimization runs and custom soljson)
```go
import (
	"github.com/0xsharma/gosolc"
)

cfg := gosolc.NewCompilerConfig("cancun", false, 0)

c, err := gosolc.NewCompiler("./contracts", cfg,"solc-bin/soljson-v0.8.29.js")
if err != nil {
    //handle error
}
```

### Compile contracts
```go
compiled, err := c.Compile()
if err != nil {
    //handle error
}
```

### Compile and write ABI and Bytecode to JSON files
```go
err = c.CompileAndWriteOutput()
if err != nil {
    panic(err)
}
```

### Get Bytecodes from compiler output
```go
bytecodes, err := compiled.GetContractByteCodes()
if err != nil {
    //handle error
}

deployedBytecodes, err := compiled.GetDeployedContractByteCodes()
if err != nil {
    //handle error
}

// Print the bytecodes
for contractName, bytecode := range bytecodes {
    fmt.Printf("ContractName: %s\n\nBytecode: %s\n\n", contractName, bytecode)
}

for contractName, bytecode := range deployedBytecodes {
    fmt.Printf("ContractName: %s\n\nDeployed Bytecode: %s\n\n", contractName, bytecode)
}
```

## Contributing <a name = "contributing"></a>
Contributions are welcome! Currently the project is using `solc version 0.8.29` by default. If you want to add support for a new version, please create a new branch and submit a pull request. Please make sure to update the README.md file with any new features or changes you make.

