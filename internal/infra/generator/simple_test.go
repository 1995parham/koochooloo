package generator_test

import (
	"testing"

	domgen "github.com/1995parham/koochooloo/internal/domain/generator"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	t.Parallel()

	s := new(generator.Simple)

	require.Implements(t, new(domgen.Generator), s)
	require.Len(t, s.ShortURLKey(), 6)
}
