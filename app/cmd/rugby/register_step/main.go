//go:build js && wasm

// register_step is the rugby widget compiled as a WebAssembly module.
// It registers `stepSimulation` on the JS global and blocks forever so
// the Go runtime stays alive to service per-step calls from
// dexetera's runtime/worker.js.
//
// Build with the codegen-emitted app/rugby/build.sh or directly:
//
//	GOOS=js GOARCH=wasm go build -o app/rugby/src/main.wasm ./app/cmd/rugby/register_step
package main

import (
	"github.com/umbralcalc/dexetera/pkg/simio"
	"github.com/umbralcalc/trywizard/app/pkg/rugby"
)

func main() {
	simio.RegisterStep(rugby.NewConfig())
}
