// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"encoding/json"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Spec(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("name")

		// --- When ---
		have, err := json.Marshal(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.JSON(t, `{"name":"name"}`, string(have))
	})

	t.Run("with args", func(t *testing.T) {
		// --- Given ---
		spc := &Spec{
			Name: "name",
			Args: map[string]any{"k0": 0, "k1": 1},
		}

		// --- When ---
		have, err := json.Marshal(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.JSON(t, `{"name":"name", "args":{"k0":0, "k1":1}}`, string(have))
	})
}

func Test_NewSpec(t *testing.T) {
	// --- When ---
	have := NewSpec("my-spec")

	// --- Then ---
	assert.Equal(t, &Spec{Name: "my-spec"}, have)
}

func Test_Spec_SetArg(t *testing.T) {
	// --- Given ---
	spc := &Spec{}

	// --- When ---
	have := spc.SetArg("k0", 0)

	// --- Then ---
	assert.Same(t, spc, have)
	assert.Equal(t, map[string]any{"k0": 0}, spc.Args)
}

func Test_Spec_ArgExist(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		spc := &Spec{
			Args: map[string]any{"name": 42},
		}

		// --- When ---
		have := spc.ArgExist("name")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		spc := &Spec{
			Args: map[string]any{"name": 42},
		}

		// --- When ---
		have := spc.ArgExist("other")

		// --- Then ---
		assert.False(t, have)
	})
}
