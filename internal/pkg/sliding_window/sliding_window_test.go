package sliding_window

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindow(t *testing.T) {
	type getContextReq struct {
		order int
		want  []byte
	}

	tests := []struct {
		name        string
		maxOrder    int
		pushes      []byte
		getContexts []getContextReq
	}{
		{
			name:     "empty_window",
			maxOrder: 3,
			pushes:   []byte{},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{}},
				{2, []byte{}},
				{3, []byte{}},
			},
		},
		{
			name:     "less_than_maxOrder",
			maxOrder: 4,
			pushes:   []byte{'A', 'B'},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{'B'}},
				{2, []byte{'A', 'B'}},
				{3, []byte{'A', 'B'}},
				{4, []byte{'A', 'B'}},
			},
		},
		{
			name:     "exactly_maxOrder",
			maxOrder: 3,
			pushes:   []byte{'X', 'Y', 'Z'},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{'Z'}},
				{2, []byte{'Y', 'Z'}},
				{3, []byte{'X', 'Y', 'Z'}},
				{4, []byte{'X', 'Y', 'Z'}},
			},
		},
		{
			name:     "wrap_around_(overwrite)",
			maxOrder: 3,
			pushes:   []byte{'A', 'B', 'C', 'D', 'E'},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{'E'}},
				{2, []byte{'D', 'E'}},
				{3, []byte{'C', 'D', 'E'}},
			},
		},
		{
			name:     "single_character_repeated_pushes",
			maxOrder: 2,
			pushes:   []byte{'Z', 'Z', 'Z'},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{'Z'}},
				{2, []byte{'Z', 'Z'}},
			},
		},
		{
			name:     "maxOrder_=_1",
			maxOrder: 1,
			pushes:   []byte{'P', 'Q'},
			getContexts: []getContextReq{
				{0, []byte{}},
				{1, []byte{'Q'}},
				{2, []byte{'Q'}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewSlidingWindow(tt.maxOrder)

			for _, b := range tt.pushes {
				w.Push(b)
			}

			for _, req := range tt.getContexts {
				got := w.GetContext(req.order)
				assert.Equal(t, req.want, got,
					"GetContext(%d) mismatch", req.order)
			}
		})
	}
}
