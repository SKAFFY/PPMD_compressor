package sliding_window

type SlidingWindow struct {
	buf  []byte
	pos  int
	size int // сколько реально символов записано (не больше len(buf))
}

func NewSlidingWindow(maxOrder int) *SlidingWindow {
	return &SlidingWindow{
		buf:  make([]byte, maxOrder),
		pos:  0,
		size: 0,
	}
}

func (w *SlidingWindow) Push(b byte) {
	w.buf[w.pos] = b
	w.pos = (w.pos + 1) % len(w.buf)

	if w.size < len(w.buf) {
		w.size++
	}
}

func (w *SlidingWindow) GetContext(order int) []byte {
	if order > w.size {
		order = w.size
	}

	if order == 0 {
		return []byte{}
	}

	context := make([]byte, order)
	for i := 0; i < order; i++ {
		idx := (w.pos - order + i) % len(w.buf)
		if idx < 0 {
			idx += len(w.buf)
		}
		context[i] = w.buf[idx]
	}

	return context
}
