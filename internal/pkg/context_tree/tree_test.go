package context_tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextTree_UpdateAndGetNode(t *testing.T) {
	tests := []struct {
		name     string
		maxOrder int
		updates  []struct {
			sym     byte
			context []byte
		}
		checks []struct {
			context      []byte
			expectedSym  byte
			expectedFreq int
			expectedNode bool
		}
	}{
		{
			name:     "single update zero order",
			maxOrder: 2,
			updates: []struct {
				sym     byte
				context []byte
			}{
				{'A', []byte{}},
			},
			checks: []struct {
				context      []byte
				expectedSym  byte
				expectedFreq int
				expectedNode bool
			}{
				{[]byte{}, 'A', 1, true},
				{[]byte{0}, 'A', 0, false},
			},
		},
		{
			name:     "update with order 1 context",
			maxOrder: 2,
			updates: []struct {
				sym     byte
				context []byte
			}{
				{'B', []byte{'x'}},
				{'B', []byte{'x'}},
				{'C', []byte{'x'}},
			},
			checks: []struct {
				context      []byte
				expectedSym  byte
				expectedFreq int
				expectedNode bool
			}{
				{[]byte{}, 'B', 2, true},
				{[]byte{}, 'C', 1, true},
				{[]byte{'x'}, 'B', 2, true},
				{[]byte{'x'}, 'C', 1, true},
				{[]byte{'y'}, 'B', 0, false},
			},
		},
		{
			name:     "update with order 2 context",
			maxOrder: 3,
			updates: []struct {
				sym     byte
				context []byte
			}{
				{'D', []byte{'a', 'b'}},
				{'D', []byte{'a', 'b'}},
				{'E', []byte{'a', 'b'}},
				{'F', []byte{'a', 'b'}},
			},
			checks: []struct {
				context      []byte
				expectedSym  byte
				expectedFreq int
				expectedNode bool
			}{
				{[]byte{}, 'D', 2, true},
				{[]byte{}, 'E', 1, true},
				{[]byte{}, 'F', 1, true},
				{[]byte{'a'}, 'D', 2, true},
				{[]byte{'a'}, 'E', 1, true},
				{[]byte{'a'}, 'F', 1, true},
				{[]byte{'a', 'b'}, 'D', 2, true},
				{[]byte{'a', 'b'}, 'E', 1, true},
				{[]byte{'a', 'b'}, 'F', 1, true},
				{[]byte{'b', 'c'}, 'D', 0, false},
			},
		},
		{
			name:     "different contexts share same suffix nodes",
			maxOrder: 2,
			updates: []struct {
				sym     byte
				context []byte
			}{
				{'X', []byte{'p'}},
				{'Y', []byte{'p'}},
				{'X', []byte{'q'}},
			},
			checks: []struct {
				context      []byte
				expectedSym  byte
				expectedFreq int
				expectedNode bool
			}{
				{[]byte{}, 'X', 2, true},
				{[]byte{}, 'Y', 1, true},
				{[]byte{'p'}, 'X', 1, true},
				{[]byte{'p'}, 'Y', 1, true},
				{[]byte{'q'}, 'X', 1, true},
				{[]byte{'q'}, 'Y', 0, true}, // узел существует, но Y не встречался после 'q'
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewContextTree(tt.maxOrder)

			for i, upd := range tt.updates {
				tree.Update(upd.sym, upd.context)
				node := tree.GetNode(upd.context)
				require.NotNil(t, node, "update %d: context %v should exist after update", i, upd.context)
				assert.GreaterOrEqual(t, node.Freq[upd.sym], 1, "update %d: Freq[%c] should be >=1", i, upd.sym)
				assert.Equal(t, node.Total, sumFreq(node.Freq), "update %d: Total mismatch", i)
			}

			for _, check := range tt.checks {
				node := tree.GetNode(check.context)
				if !check.expectedNode {
					assert.Nil(t, node, "context %v should not exist", check.context)
					continue
				}
				require.NotNil(t, node, "context %v should exist", check.context)
				Freq := node.Freq[check.expectedSym] // отсутствие ключа даёт 0
				assert.Equal(t, check.expectedFreq, Freq, "context %v Freq[%c]", check.context, check.expectedSym)
			}
		})
	}
}

func sumFreq(Freq map[byte]int) int {
	sum := 0
	for _, v := range Freq {
		sum += v
	}
	return sum
}

func TestContextTree_GetNodeNonExistent(t *testing.T) {
	tree := NewContextTree(3)
	node := tree.GetNode([]byte("hello"))
	assert.Nil(t, node)

	tree.Update('a', []byte{})
	node = tree.GetNode([]byte{})
	require.NotNil(t, node)
	assert.Equal(t, 1, node.Freq['a'])

	node = tree.GetNode([]byte{'a'})
	assert.Nil(t, node)
}

func TestContextTree_MaxOrderNotExceeded(t *testing.T) {
	tree := NewContextTree(2)
	longContext := []byte{'x', 'y', 'z'}
	tree.Update('b', longContext)

	nodeXY := tree.GetNode([]byte{'x', 'y'})
	require.NotNil(t, nodeXY)
	assert.Equal(t, 1, nodeXY.Freq['b'])

	nodeXYZ := tree.GetNode([]byte{'x', 'y', 'z'})
	require.NotNil(t, nodeXYZ)
	assert.Equal(t, 1, nodeXYZ.Freq['b'])
}
