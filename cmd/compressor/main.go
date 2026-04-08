package main

import (
	"PPMA_compressor/internal/pkg/arithmetic_encoder_decoder"
	"PPMA_compressor/internal/pkg/compressor_decompressor"
	"bufio"
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
)

func main() {
	sourceFilePath := flag.String("f", "", "source file path")
	targetFilePath := flag.String("t", "", "target file path")
	maxContextOrder := flag.Int("max-context-order", 4, "max context order (default 4)")
	writerBufferSize := flag.Int("buffer-size", 4*1024*1024, "buffer size (default 4MB)")
	showProgress := flag.Bool("progress", true, "show progress bar")
	flag.Parse()

	if *sourceFilePath == "" || *targetFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	sourceFile, err := os.Open(*sourceFilePath)
	if err != nil {
		fmt.Printf("Error opening source file: %v\n", err)
		return
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			fmt.Printf("Error closing source file: %v\n", err)
		}
	}()

	var sourceSize int64
	if *showProgress {
		stat, err := sourceFile.Stat()
		if err == nil {
			sourceSize = stat.Size()
		}
	}

	targetFile, err := os.Create(*targetFilePath)
	if err != nil {
		fmt.Printf("Error creating target file: %v\n", err)
		return
	}
	defer func() {
		if err := targetFile.Close(); err != nil {
			fmt.Printf("Error closing target file: %v\n", err)
		}
	}()

	bufTargetFile := bufio.NewWriterSize(targetFile, *writerBufferSize)
	defer func() {
		if err := bufTargetFile.Flush(); err != nil {
			fmt.Printf("Error flushing buffer: %v\n", err)
		}
	}()

	originalSize := uint64(sourceSize)
	arithmeticEncoder := arithmetic_encoder_decoder.NewArithmeticEncoder(bufTargetFile)

	compressor, err := compressor_decompressor.NewCompressor(bufTargetFile, arithmeticEncoder, *maxContextOrder, originalSize)
	if err != nil {
		fmt.Printf("Error creating compressor: %v\n", err)
		return
	}
	defer func() {
		if err := compressor.Close(); err != nil {
			fmt.Printf("Error closing compressor: %v\n", err)
		}
	}()

	var reader io.Reader = sourceFile
	if *showProgress && sourceSize > 0 {
		bar := progressbar.DefaultBytes(sourceSize, "compressing")
		reader = io.TeeReader(sourceFile, bar)
	}

	_, err = io.Copy(compressor, reader)
	if err != nil {
		fmt.Printf("Error compressing: %v\n", err)
		return
	}

	fmt.Println("\nCompression finished successfully")
}
