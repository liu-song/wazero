module github.com/tetratelabs/wazero/internal/integration_test/vs/wasmtime

go 1.18

require (
	github.com/bytecodealliance/wasmtime-go/v8 v8.0.0
	github.com/tetratelabs/wazero v0.0.0
)

replace github.com/tetratelabs/wazero => ../../../..
