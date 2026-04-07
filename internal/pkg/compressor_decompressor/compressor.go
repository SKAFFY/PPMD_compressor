package compressor_decompressor

import (
	"PPMA_compressor/internal/pkg/context_tree"
	"PPMA_compressor/internal/pkg/sliding_window"
	"io"
)

const (
	Escape int = 1
)

// EncoderWriter should write to destination
type EncoderWriter interface {
	Encode(sym byte, cumFreq []uint64, totalFreq uint64)
	Flush() error
}

type Compressor struct {
	encoder       EncoderWriter
	contextTree   *context_tree.ContextTree
	maxOrder      int
	slidingWindow *sliding_window.SlidingWindow
}

func NewCompressor(encoder EncoderWriter, maxOrder int) *Compressor {
	return &Compressor{
		dst:           dst,
		encoder:       encoder,
		contextTree:   context_tree.NewContextTree(maxOrder),
		maxOrder:      maxOrder,
		slidingWindow: sliding_window.NewSlidingWindow(maxOrder),
	}
}

func (c *Compressor) Write(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Compressor) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *Compressor) compress(data []byte) []byte {
	for _, sym := range data {
		order := c.maxOrder
		context := c.slidingWindow.GetContext(order)

		for order >= 0 {
			probs := c.contextTree.GetProbabilities(context)

			if _, exists := probs[sym]; exists {
				c.encoder.EncodeSymbol(sym, probs)
				break
			} else {
				c.encoder.EncodeSymbol(Escape, probs)
				order--
				context = c.slidingWindow.GetContext(order)
			}
		}

		if order < 0 {
			c.encoder.EncodeSymbolUniform(sym, 256)
		}

		c.contextTree.Update(sym, c.slidingWindow.GetContext(c.maxOrder))
		c.slidingWindow.Push(sym)
	}

	return c.encoder.Flush()
}
