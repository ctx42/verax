# spec

**Serializable type specifications for Go** — describe a value or process as a
portable `Spec`, then encode it to JSON and rebuild it later.

```
import "github.com/ctx42/verax/pkg/spec"
```

---

### What it does

A `Spec` is a named bag of arguments (`map[string]any`) that fully describes how
to reconstruct a value. Types that implement `Specable` return their own `Spec`;
a `Registry[T]` encodes those specs to JSON and decodes them back into concrete
values via registered `Builder` functions. Go types are preserved across the
round trip (through [`jsontype`](https://github.com/ctx42/jsontype)), and values
that cannot be serialized — such as functions — are referenced by name through a
registered `Source`.

This is the mechanism `verax` uses to store validation rules in a database or
exchange them over an API; see `verax.Builders` for a full set of builders.

### Usage

```go
// greeting describes itself as a spec and can be rebuilt from one.
type greeting struct {
	Name string
}

func (g greeting) Spec() (*spec.Spec, error) {
	return spec.NewSpec("greeting").SetArg(spec.ArgValue, g.Name), nil
}

func newGreeting(spc *spec.Spec) (greeting, error) {
	name, _ := spc.Args[spec.ArgValue].(string)
	return greeting{Name: name}, nil
}

func main() {
	reg := spec.NewRegistry[greeting]()
	reg.RegisterBuilder("greeting", newGreeting)

	// Encode a value's spec to JSON.
	spc, _ := greeting{Name: "World"}.Spec()
	data, _ := reg.EncodeSpec(spc)
	fmt.Println(string(data))
	// {"name":"greeting","args":{"value":"World"}}

	// Decode the JSON and rebuild the value.
	have, _ := reg.DecodeAndBuild(data)
	fmt.Printf("Hello, %s!\n", have.Name)
	// Hello, World!
}
```

See `ExampleRegistry` in the package tests for the runnable version.
