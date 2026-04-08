package arithmetic_encoder_decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uniformCumFreq возвращает накопительные частоты для 256 символов (0..255)
func uniformCumFreq() ([]uint64, uint64) {
	cum := make([]uint64, 257) // индексы 0..256
	for i := 0; i < 256; i++ {
		cum[i+1] = cum[i] + 1
	}
	return cum, 256
}

// uniformCumFreqWithEscape возвращает cumFreq для 256 символов + escape (индекс 256)
func uniformCumFreqWithEscape() ([]uint64, uint64) {
	cum := make([]uint64, 258) // индексы 0..257
	for i := 0; i < 256; i++ {
		cum[i+1] = cum[i] + 1
	}
	cum[257] = cum[256] + 1 // escape имеет частоту 1
	return cum, 257
}

// stringToIntSlice преобразует строку в срез int (символы 0..255)
func stringToIntSlice(s string) []int {
	res := make([]int, len(s))
	for i, b := range []byte(s) {
		res[i] = int(b)
	}
	return res
}

func TestEncoderDecoder(t *testing.T) {
	tests := []struct {
		name    string
		symbols []int
	}{
		{"single symbol", []int{'A'}},
		{"two symbols", []int{0, 1, 0, 1, 0}},
		{"three symbols", []int{0, 1, 2, 0, 2, 1, 0}},
		{"hello world", stringToIntSlice("Hello, world!")},
		{"with escape", []int{1, 2, 256, 3, 256, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Для тестов с escape используем cumFreq с escape, иначе стандартную
			hasEscape := false
			for _, sym := range tt.symbols {
				if sym == 256 {
					hasEscape = true
					break
				}
			}
			var cum []uint64
			var total uint64
			if hasEscape {
				cum, total = uniformCumFreqWithEscape()
			} else {
				cum, total = uniformCumFreq()
			}

			enc := NewArithmeticEncoder(&buf)
			for _, sym := range tt.symbols {
				enc.Encode(sym, cum, total)
			}
			err := enc.Flush()
			require.NoError(t, err)

			t.Logf("Compressed length: %d bytes", buf.Len())
			if buf.Len() == 0 {
				t.Fatal("No data written")
			}

			dec := NewArithmeticDecoder(&buf)

			decoded := make([]int, len(tt.symbols))
			for i := 0; i < len(tt.symbols); i++ {
				sym, err := dec.Decode(cum, total)
				require.NoError(t, err, "i=%d", i)
				decoded[i] = sym
			}
			assert.Equal(t, tt.symbols, decoded)
		})
	}
}
