package compressor_decompressor

// Вспомогательные функции для построения накопительных частот

// buildCumFreq строит cumFreq для обычных символов 0..255 (без escape)
func buildCumFreq(Freqs map[byte]int) []uint64 {
	cum := make([]uint64, 257) // индексы 0..256
	for i := 0; i < 256; i++ {
		f := uint64(Freqs[byte(i)])
		cum[i+1] = cum[i] + f
	}
	return cum
}

// buildCumFreqWithEscape строит cumFreq для символов 0..255 + escape (индекс 256)
func buildCumFreqWithEscape(Freqs map[byte]int, escapeFreq uint64) []uint64 {
	cum := make([]uint64, 258) // индексы 0..257
	for i := 0; i < 256; i++ {
		f := uint64(Freqs[byte(i)])
		cum[i+1] = cum[i] + f
	}
	cum[257] = cum[256] + escapeFreq
	return cum
}
