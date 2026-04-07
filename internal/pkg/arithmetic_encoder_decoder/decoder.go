package arithmetic_encoder_decoder

import (
	"io"
)

type ArithmeticDecoder struct {
	low, high uint32
	value     uint32
	in        io.Reader
	buf       byte
	bits      uint8
	err       error
}

// NewArithmeticDecoder создаёт декодер и читает первые 32 бита в value.
func NewArithmeticDecoder(r io.Reader) (*ArithmeticDecoder, error) {
	d := &ArithmeticDecoder{
		low:  0,
		high: 0xFFFFFFFF,
		in:   r,
	}

	// Читаем 32 бита для начального значения value.
	for i := 0; i < 32; i++ {
		bit, err := d.readBit()
		if err != nil {
			return nil, err
		}
		d.value = (d.value << 1) | uint32(bit)
	}
	return d, nil
}

// readBit читает один бит из входного потока.
func (d *ArithmeticDecoder) readBit() (byte, error) {
	if d.bits == 0 {
		n, err := d.in.Read([]byte{d.buf})
		if err != nil {
			if err == io.EOF {
				// Конец потока — считаем, что дальше идут нулевые биты.
				// Это стандартный приём для инициализации и нормализации.
				d.buf = 0
				d.bits = 8
				return 0, nil
			}
			return 0, err
		}
		if n == 0 {
			return 0, io.EOF
		}
		d.bits = 8
	}
	d.bits--
	return (d.buf >> d.bits) & 1, nil
}

// Decode декодирует следующий символ, используя кумулятивные частоты cumFreq.
func (d *ArithmeticDecoder) Decode(cumFreq []uint64, totalFreq uint64) (byte, error) {
	if d.err != nil {
		return 0, d.err
	}

	rng := uint64(d.high-d.low) + 1

	// Вычисляем масштабированное значение для поиска символа.
	val := ((uint64(d.value-d.low)+1)*totalFreq - 1) / rng

	// Бинарный поиск символа по cumFreq.
	lo, hi := 0, len(cumFreq)-2
	for lo <= hi {
		mid := (lo + hi) / 2
		if cumFreq[mid] <= val && val < cumFreq[mid+1] {
			sym := byte(mid)

			// Обновляем границы так же, как в кодере.
			d.high = d.low + uint32((rng*cumFreq[sym+1])/totalFreq) - 1
			d.low = d.low + uint32((rng*cumFreq[sym])/totalFreq)

			// Нормализация.
			for {
				if d.high < Half {
					// E1: ничего не делаем, старший бит 0.
				} else if d.low >= Half {
					// E2: старший бит 1.
					d.value -= Half
					d.low -= Half
					d.high -= Half
				} else if d.low >= Quarter && d.high < Quarter+Half {
					// E3: убираем второй бит.
					d.value -= Quarter
					d.low -= Quarter
					d.high -= Quarter
				} else {
					break
				}

				// Сдвиг и подгрузка нового бита.
				d.low <<= 1
				d.high = (d.high << 1) | 1
				bit, err := d.readBit()
				if err != nil {
					d.err = err
					return 0, err
				}
				d.value = (d.value << 1) | uint32(bit)
			}
			return sym, nil
		} else if val < cumFreq[mid] {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return 0, io.ErrUnexpectedEOF
}
