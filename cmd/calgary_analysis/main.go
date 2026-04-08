package main

import (
	"PPMC_compressor/internal/pkg/arithmetic_encoder_decoder"
	"PPMC_compressor/internal/pkg/compressor_decompressor"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

// entropyZeroOrder вычисляет энтропию 0-го порядка (бит на символ)
func entropyZeroOrder(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}
	freq := make([]int, 256)
	for _, b := range data {
		freq[b]++
	}
	total := float64(len(data))
	var h float64
	for _, f := range freq {
		if f == 0 {
			continue
		}
		p := float64(f) / total
		h -= p * math.Log2(p)
	}
	return h
}

// entropyFirstOrder вычисляет условную энтропию H(X_{i+1} | X_i)
func entropyFirstOrder(data []byte) float64 {
	if len(data) < 2 {
		return 0.0
	}
	freqSym := make([]int, 256)
	pairFreq := make([]int, 256*256)
	for i := 0; i < len(data)-1; i++ {
		prev := data[i]
		next := data[i+1]
		freqSym[prev]++
		pairFreq[int(prev)<<8|int(next)]++
	}
	totalPairs := len(data) - 1
	var h float64
	for prev := 0; prev < 256; prev++ {
		countPrev := freqSym[prev]
		if countPrev == 0 {
			continue
		}
		for next := 0; next < 256; next++ {
			cnt := pairFreq[prev<<8|next]
			if cnt == 0 {
				continue
			}
			pPair := float64(cnt) / float64(totalPairs)
			pCond := float64(cnt) / float64(countPrev)
			h -= pPair * math.Log2(pCond)
		}
	}
	return h
}

// entropySecondOrder вычисляет условную энтропию H(X_{i+2} | X_i, X_{i+1})
func entropySecondOrder(data []byte) float64 {
	if len(data) < 3 {
		return 0.0
	}
	contextFreq := make([]int, 256*256)
	tripleFreq := make([]int, 256*256*256)
	for i := 0; i < len(data)-2; i++ {
		ctx := int(data[i])<<8 | int(data[i+1])
		next := data[i+2]
		contextFreq[ctx]++
		tripleFreq[ctx<<8|int(next)]++
	}
	totalTriples := len(data) - 2
	var h float64
	for ctx := 0; ctx < 256*256; ctx++ {
		countCtx := contextFreq[ctx]
		if countCtx == 0 {
			continue
		}
		for next := 0; next < 256; next++ {
			cnt := tripleFreq[ctx<<8|next]
			if cnt == 0 {
				continue
			}
			pTriple := float64(cnt) / float64(totalTriples)
			pCond := float64(cnt) / float64(countCtx)
			h -= pTriple * math.Log2(pCond)
		}
	}
	return h
}

// compressedSizeAndBitsPerSymbol сжимает данные и возвращает размер в байтах и биты на символ
func compressedSizeAndBitsPerSymbol(data []byte, maxOrder int) (compSize int, bitsPerSym float64) {
	var buf bytes.Buffer
	enc := arithmetic_encoder_decoder.NewArithmeticEncoder(&buf)
	comp, err := compressor_decompressor.NewCompressor(&buf, enc, maxOrder, uint64(len(data)))
	if err != nil {
		return 0, 0
	}
	_, err = comp.Write(data)
	if err != nil {
		return 0, 0
	}
	err = comp.Close()
	if err != nil {
		return 0, 0
	}
	compSize = buf.Len()
	if len(data) > 0 {
		bitsPerSym = float64(compSize*8) / float64(len(data))
	}
	return compSize, bitsPerSym
}

func main() {
	dir := flag.String("dir", "../test/test_dataset", "Directory containing Calgary corpus files")
	maxOrder := flag.Int("max-order", 4, "Max context order for PPM compressor")
	outputFile := flag.String("output", "", "Save table to file (CSV format, append .csv or .md)")
	flag.Parse()
	if *dir == "" {
		fmt.Println("Please provide -dir with path to Calgary corpus files")
		flag.Usage()
		os.Exit(1)
	}

	fileNames := []string{
		"bib", "book1", "book2", "geo", "news", "obj1", "obj2",
		"paper1", "paper2", "pic", "progc", "progl", "progp", "trans",
	}

	var out io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}

	fmt.Fprintf(out, "%-10s %8s %8s %8s %10s %12s %12s\n",
		"File", "H0 (bits)", "H1 (bits)", "H2 (bits)", "Size (B)", "CompSize (B)", "Bits/sym")
	fmt.Fprintf(out, "---------- -------- -------- -------- ---------- ------------ ------------\n")

	var totalCompSize int64 = 0

	for _, name := range fileNames {
		path := filepath.Join(*dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(out, "%-10s %s\n", name, fmt.Sprintf("ERROR: %v", err))
			continue
		}
		h0 := entropyZeroOrder(data)
		h1 := entropyFirstOrder(data)
		h2 := entropySecondOrder(data)
		compSize, bitsPerSym := compressedSizeAndBitsPerSymbol(data, *maxOrder)
		totalCompSize += int64(compSize)

		fmt.Fprintf(out, "%-10s %8.4f %8.4f %8.4f %10d %12d %12.4f\n",
			name, h0, h1, h2, len(data), compSize, bitsPerSym)
	}
	fmt.Fprintf(out, "\nTotal compressed size (all files): %d bytes\n", totalCompSize)
}
