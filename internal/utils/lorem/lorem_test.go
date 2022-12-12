// Copyright 2012 Derek A. Rhodes.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lorem

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLorem(t *testing.T) {
	for i := 1; i < 14; i++ {
		for j := i + 1; j < 14; j++ {
			require.True(t, len(Word(i, j)) >= 1)
			require.True(t, len(Sentence(i, j)) >= 1)
			require.True(t, len(Sentence(i, j)) >= 1)
		}
	}
}
