package arithmetic_encoder_decoder

type ArithmeticEncoder struct{}

func NewArithmeticEncoder() *ArithmeticEncoder {
	return &ArithmeticEncoder{}
}

func (a ArithmeticEncoder) Encode(sym byte, cumFreq []uint32, totalFreq uint32) {
	//TODO implement me
	panic("implement me")
}

func (a ArithmeticEncoder) Flush() []byte {
	//TODO implement me
	panic("implement me")
}
