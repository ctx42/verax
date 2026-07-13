// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec_test

import (
	"fmt"

	"github.com/ctx42/verax/pkg/spec"
)

// greeting is a small type that describes itself as a [spec.Spec] and can be
// rebuilt from one, making it serializable through a [spec.Registry].
type greeting struct {
	Name string
}

// Spec implements [spec.Specable].
func (g greeting) Spec() (*spec.Spec, error) {
	return spec.NewSpec("greeting").SetArg(spec.ArgValue, g.Name), nil
}

// newGreeting is a [spec.Builder] that rebuilds a greeting from its spec.
func newGreeting(spc *spec.Spec) (greeting, error) {
	name, _ := spc.Args[spec.ArgValue].(string)
	return greeting{Name: name}, nil
}

func ExampleRegistry() {
	reg := spec.NewRegistry[greeting]()
	reg.RegisterBuilder("greeting", newGreeting)

	// Describe a value as a spec and encode the spec to JSON.
	spc, _ := greeting{Name: "World"}.Spec()
	data, _ := reg.EncodeSpec(spc)
	fmt.Println(string(data))

	// Decode the JSON back into a spec and rebuild the value from it.
	have, _ := reg.DecodeAndBuild(data)
	fmt.Printf("Hello, %s!\n", have.Name)

	// Output:
	// {"name":"greeting","args":{"value":"World"}}
	// Hello, World!
}
