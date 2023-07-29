package generator_test

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	t.Parallel()

	s := new(generator.Simple)

	require.Implements(t, new(generator.Generator), s)
	require.Len(t, s.ShortURLKey(), 6)
}
