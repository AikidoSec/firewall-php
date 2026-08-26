package zen_internals

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetZenInternalsLibPath(t *testing.T) {
	for _, libcVariant := range []string{"gnu", "musl"} {
		t.Run(libcVariant, func(t *testing.T) {
			assert.Equal(
				t,
				fmt.Sprintf("/opt/aikido-1.2.3/libzen_internals_x86_64-unknown-linux-%s.so", libcVariant),
				getZenInternalsLibPath("1.2.3", "x86_64", libcVariant),
			)
		})
	}
}
