package arithmetic_encoder_decoder

import (
	"io"
)

type ArithmeticEncoder struct {
	low, high uint64
	pending   uint64
	out       io.Writer
	buf       byte
	bits      uint8
	err       error
}

func NewArithmeticEncoder(w io.Writer) *ArithmeticEncoder {
	return &ArithmeticEncoder{
		low:  0,
		high: TopValue - 1,
		out:  w,
	}
}

func (e *ArithmeticEncoder) writeBit(bit uint64) {
	if e.err != nil {
		return
	}
	if bit == 1 {
		e.buf |= 1 << (7 - e.bits)
	}
	e.bits++
	if e.bits == 8 {
		_, e.err = e.out.Write([]byte{e.buf})
		e.buf = 0
		e.bits = 0
	}
}

// Encode кодирует символ sym (0..256, где 256 = escape) с использованием накопительных частот.
func (e *ArithmeticEncoder) Encode(sym int, cumFreq []uint64, totalFreq uint64) {
	if e.err != nil || totalFreq == 0 {
		return
	}
	rng := e.high - e.low + 1
	e.high = e.low + (rng*cumFreq[sym+1])/totalFreq - 1
	e.low = e.low + (rng*cumFreq[sym])/totalFreq

	for {
		if e.high < Half {
			e.writeBit(0)
			for e.pending > 0 {
				e.writeBit(1)
				e.pending--
			}
		} else if e.low >= Half {
			e.writeBit(1)
			for e.pending > 0 {
				e.writeBit(0)
				e.pending--
			}
			e.low -= Half
			e.high -= Half
		} else if e.low >= FirstQtr && e.high < ThirdQtr {
			e.pending++
			e.low -= FirstQtr
			e.high -= FirstQtr
		} else {
			break
		}
		e.low <<= 1
		e.high = (e.high << 1) | 1
	}
}

func (e *ArithmeticEncoder) Flush() error {
	if e.err != nil {
		return e.err
	}
	e.pending++
	if e.low < FirstQtr {
		e.writeBit(0)
		for e.pending > 0 {
			e.writeBit(1)
			e.pending--
		}
	} else {
		e.writeBit(1)
		for e.pending > 0 {
			e.writeBit(0)
			e.pending--
		}
	}
	if e.bits > 0 {
		_, e.err = e.out.Write([]byte{e.buf})
	}
	return e.err
}
