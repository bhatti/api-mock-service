package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVariable(t *testing.T) {
	assert.Equal(
		t,
		&PostmanVariable{
			Name:  "a-name",
			Value: "a-value",
			Type:  "string",
		},
		CreatePostmanVariable("a-name", "a-value"),
	)
}
