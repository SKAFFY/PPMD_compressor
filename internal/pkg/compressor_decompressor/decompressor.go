package compressor_decompressor

type Decoder interface {
	Decode(cumFreq []uint32, totalFreq uint32) byte
}
