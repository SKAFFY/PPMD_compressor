package compressor_decompressor

import (
	"PPMA_compressor/internal/pkg/arithmetic_encoder_decoder"
	"PPMA_compressor/internal/pkg/context_tree"
	"PPMA_compressor/internal/pkg/sliding_window"
	"encoding/binary"
	"io"
)

type Decompressor struct {
	decoder       *arithmetic_encoder_decoder.ArithmeticDecoder
	contextTree   *context_tree.ContextTree
	maxOrder      int
	slidingWindow *sliding_window.SlidingWindow
	remaining     uint64
	originalSize  uint64 // сохраняем оригинальный размер
}

// NewDecompressor читает заголовок из r и создаёт декомпрессор с арифметическим декодером.
func NewDecompressor(r io.Reader) (*Decompressor, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	maxOrder := int(header[0])
	originalSize := binary.LittleEndian.Uint64(header[1:])

	decoder := arithmetic_encoder_decoder.NewArithmeticDecoder(r)

	return &Decompressor{
		decoder:       decoder,
		contextTree:   context_tree.NewContextTree(maxOrder),
		maxOrder:      maxOrder,
		slidingWindow: sliding_window.NewSlidingWindow(maxOrder),
		remaining:     originalSize,
		originalSize:  originalSize,
	}, nil
}

func (d *Decompressor) OriginalSize() uint64 {
	return d.originalSize
}

// Read – без изменений, останавливается при remaining == 0
func (d *Decompressor) Read(p []byte) (n int, err error) {
	for n < len(p) && d.remaining > 0 {
		sym, err := d.decodeNextSymbol()
		if err != nil {
			return n, err
		}
		p[n] = byte(sym)
		n++
		d.remaining--
		// обновление модели
		fullContext := d.slidingWindow.GetContext(d.maxOrder)
		d.contextTree.Update(byte(sym), fullContext)
		d.slidingWindow.Push(byte(sym))
	}
	if d.remaining == 0 && n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// decodeNextSymbol декодирует один символ (0..255) из арифметического потока.
// Возвращает символ или ошибку (включая io.EOF).
func (d *Decompressor) decodeNextSymbol() (int, error) {
	order := d.maxOrder
	context := d.slidingWindow.GetContext(order)

	for order >= 0 {
		node := d.contextTree.GetNode(context)
		if node != nil && node.Total > 0 {
			escapeFreq := uint64(1)
			if node != nil {
				escapeFreq = uint64(len(node.Freq)) // метод C: количество различных символов
			}
			cumFreq := buildCumFreqWithEscape(node.Freq, escapeFreq)
			totalFreq := uint64(node.Total) + escapeFreq
			sym, err := d.decoder.Decode(cumFreq, totalFreq)
			if err != nil {
				return 0, err
			}
			if sym != Escape {
				return sym, nil
			}
			// escape — переходим к меньшему порядку
		} else {
			// узел не существует — кодировался только escape (частота 1)
			cumFreq := make([]uint64, 258)
			cumFreq[257] = 1
			sym, err := d.decoder.Decode(cumFreq, 1)
			if err != nil {
				return 0, err
			}
			if sym != Escape {
				// такого не должно быть, но игнорируем
			}
		}
		order--
		if order >= 0 {
			context = d.slidingWindow.GetContext(order)
		}
	}
	// Порядок -1: равномерное распределение (256 символов)
	uniformCumFreq := make([]uint64, 257)
	for i := 0; i < 256; i++ {
		uniformCumFreq[i+1] = uint64(i + 1)
	}
	return d.decoder.Decode(uniformCumFreq, 256)
}
