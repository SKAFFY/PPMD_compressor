package compressor_decompressor

import (
	"PPMA_compressor/internal/pkg/context_tree"
	"PPMA_compressor/internal/pkg/sliding_window"
	"encoding/binary"
	"io"
)

// EncoderWriter определяет интерфейс арифметического кодера
type EncoderWriter interface {
	Encode(sym int, cumFreq []uint64, totalFreq uint64)
	Flush() error
}

type Compressor struct {
	encoder       EncoderWriter
	contextTree   *context_tree.ContextTree
	maxOrder      int
	slidingWindow *sliding_window.SlidingWindow
}

// NewCompressor принимает writer (для заголовка) и энкодер (который пишет в тот же writer)
func NewCompressor(w io.Writer, encoder EncoderWriter, maxOrder int, originalSize uint64) (*Compressor, error) {
	// Заголовок: 1 байт maxOrder + 8 байт originalSize
	header := make([]byte, 9)
	header[0] = byte(maxOrder)
	binary.LittleEndian.PutUint64(header[1:], originalSize)
	if _, err := w.Write(header); err != nil {
		return nil, err
	}
	return &Compressor{
		encoder:       encoder,
		contextTree:   context_tree.NewContextTree(maxOrder),
		maxOrder:      maxOrder,
		slidingWindow: sliding_window.NewSlidingWindow(maxOrder),
	}, nil
}

// Write реализует io.Writer – сжимает поступающие данные и передаёт в арифметический кодер
func (c *Compressor) Write(p []byte) (n int, err error) {
	for _, sym := range p {
		order := c.maxOrder
		context := c.slidingWindow.GetContext(order)

		for order >= 0 {
			node := c.contextTree.GetNode(context)
			if node != nil && node.Freq[sym] > 0 {
				// Символ найден в текущем контексте – кодируем его
				cumFreq := buildCumFreq(node.Freq)
				c.encoder.Encode(int(sym), cumFreq, uint64(node.Total))
				break
			} else {
				// Символ не найден – кодируем escape и переходим к меньшему порядку
				escapeFreq := uint64(1)
				if node != nil {
					escapeFreq = uint64(len(node.Freq)) // метод C
				}
				var cumFreq []uint64
				var totalFreq uint64
				if node != nil {
					cumFreq = buildCumFreqWithEscape(node.Freq, escapeFreq)
					totalFreq = uint64(node.Total) + escapeFreq
				} else {
					// Контекст вообще не существует – только escape (нужен массив длиной 258)
					cumFreq = make([]uint64, 258) // индексы 0..257
					cumFreq[257] = escapeFreq
					totalFreq = escapeFreq
				}
				c.encoder.Encode(Escape, cumFreq, totalFreq)

				order--
				if order >= 0 {
					context = c.slidingWindow.GetContext(order)
				}
				continue
			}
		}

		if order < 0 {
			// Порядок -1: равномерное распределение по 256 символам
			uniformCumFreq := make([]uint64, 257)
			for i := 0; i < 256; i++ {
				uniformCumFreq[i+1] = uint64(i + 1)
			}
			c.encoder.Encode(int(sym), uniformCumFreq, 256)
		}

		// Обновляем статистику для всех суффиксов контекста максимальной длины
		fullContext := c.slidingWindow.GetContext(c.maxOrder)
		c.contextTree.Update(sym, fullContext)
		c.slidingWindow.Push(sym)
	}
	return len(p), nil
}

// Close завершает сжатие – вызывает Flush у арифметического кодера
func (c *Compressor) Close() error {
	err := c.encoder.Flush()
	if err != nil {
		return err
	}

	return c.encoder.Flush()
}
