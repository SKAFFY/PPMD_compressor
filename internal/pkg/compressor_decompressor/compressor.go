package compressor_decompressor

import (
	"PPMA_compressor/internal/pkg/context_tree"
	"io"
)

const (
	MaxOrder int = 4
	Escape   int = 1
)

type Encoder interface {
	Encode(sym byte, cumFreq []uint32, totalFreq uint32)
	Flush() []byte
}

type Compressor struct {
	dst io.Writer

	encoder       Encoder
	contextTree   *context_tree.ContextTree
	slidingWindow *slidingWindow
}

func NewCompressor(dst io.Writer, encoder Encoder) *Compressor {
	return &Compressor{
		dst:           dst,
		encoder:       encoder,
		contextTree:   context_tree.NewContextTree(MaxOrder),
		slidingWindow: NewSlidingWindow(MaxOrder),
	}
}

func (c *Compressor) Write(p []byte) (n int, err error) {

}

func (c *Compressor) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *Compressor) compress(data []byte) []byte {

	for _, sym := range data {
		order := MaxOrder
		context := c.slidingWindow.GetContext(order)

		for order >= 0 {
			probs := tree.GetProbabilities(context)

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
			encoder.EncodeSymbolUniform(sym, 256)
		}

		c.contextTree.Update(sym, c.slidingWindow.GetContext(MaxOrder))
		c.slidingWindow.Push(sym)
	}

	return c.encoder.Flush()
}
