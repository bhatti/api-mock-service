package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCreateEvent(t *testing.T) {
	assert.Equal(
		t,
		&PostmanEvent{
			Listen: Test,
			Script: &PostmanScript{
				Type: "text/javascript",
				Exec: []string{"console.log(\"foo\")"},
			},
		},
		CreatePostmanEvent(Test, []string{"console.log(\"foo\")"}),
	)
}
