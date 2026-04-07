package arithmetic_encoder_decoder

import (
	"io"
)

type ArithmeticEncoder struct {
	low, high uint32
	pending   uint32 // счётчик отложенных битов (сколько раз встретилась ситуация E3)
	out       io.Writer
	buf       byte
	bits      uint8
	err       error
}

func NewArithmeticEncoder(w io.Writer) *ArithmeticEncoder {
	return &ArithmeticEncoder{
		low:  0,
		high: 0xFFFFFFFF,
		out:  w,
	}
}

// writeBit записывает один бит в выходной поток.
func (e *ArithmeticEncoder) writeBit(bit uint32) {
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

// Encode кодирует один символ, используя кумулятивные частоты cumFreq (длина 257)
// и общую сумму частот totalFreq.
func (e *ArithmeticEncoder) Encode(sym byte, cumFreq []uint64, totalFreq uint64) {
	if e.err != nil {
		return
	}

	// Вычисляем текущий диапазон как 64‑битное число, чтобы избежать переполнения.
	rng := uint64(e.high-e.low) + 1

	// Новые границы интервала.
	e.high = e.low + uint32((rng*cumFreq[sym+1])/totalFreq) - 1
	e.low = e.low + uint32((rng*cumFreq[sym])/totalFreq)

	// Нормализация (масштабирование).
	for {
		// Ситуация E1: старший бит low и high равен 0.
		if e.high < Half {
			e.writeBit(0)
			for e.pending > 0 {
				e.writeBit(1)
				e.pending--
			}
		} else if e.low >= Half {
			// Ситуация E2: старший бит равен 1.
			e.writeBit(1)
			for e.pending > 0 {
				e.writeBit(0)
				e.pending--
			}
			e.low -= Half
			e.high -= Half
		} else if e.low >= Quarter && e.high < Quarter+Half {
			// Ситуация E3: второй по старшинству бит различается.
			e.pending++
			e.low -= Quarter
			e.high -= Quarter
		} else {
			break
		}

		// Сдвиг влево на 1 бит.
		e.low <<= 1
		e.high = (e.high << 1) | 1
	}
}

// Flush завершает кодирование и дописывает оставшиеся биты.
func (e *ArithmeticEncoder) Flush() error {
	if e.err != nil {
		return e.err
	}

	// Гарантируем, что декодер сможет однозначно определить последние символы.
	e.pending++
	if e.low < Quarter {
		e.writeBit(0)
	} else {
		e.writeBit(1)
	}
	// Сбрасываем отложенные биты с противоположным значением.
	for e.pending > 0 {
		e.writeBit(1 - (e.low>>31)&1)
		e.pending--
	}

	// Дописываем неполный байт.
	if e.bits > 0 {
		_, err := e.out.Write([]byte{e.buf})
		if err != nil {
			return err
		}
	}
	return nil
}
