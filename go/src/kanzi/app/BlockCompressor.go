/*
Copyright 2011, 2012 Frederic Langlet
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
you may obtain a copy of the License at

                http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"kanzi/function"
	"kanzi/io"
	"os"
	"strings"
	"time"
)

const (
	DEFAULT_BUFFER_SIZE = 32768
	WARN_EMPTY_INPUT    = -128
)

type BlockCompressor struct {
	debug        bool
	silent       bool
	overwrite    bool
	inputName    string
	outputName   string
	entropyCodec string
	blockSize    uint
}

func NewBlockCompressor() (*BlockCompressor, error) {
	this := new(BlockCompressor)

	// Define flags
	var help = flag.Bool("help", false, "display the help message")
	var debug = flag.Bool("debug", false, "display the size of the encoded block pre-entropy coding")
	var silent = flag.Bool("silent", false, "silent mode: no output (except warnings and errors)")
	var overwrite = flag.Bool("overwrite", false, "overwrite the output file if it already exists")
	var inputName = flag.String("input", "", "mandatory name of the input file to encode")
	var outputName = flag.String("output", "", "optional name of the output file (defaults to <input.knz>)")
	var blockSize = flag.Int("block", 100000, "size of the block in bytes (max 16 MB / default 100 KB)")
	var entropy = flag.String("entropy", "Huffman", "Entropy codec to use [None|Huffman|Range|PAQ|FPAQ]")

	// Parse
	flag.Parse()

	if *help == true {
		printOut("-help              : display this message", true)
		printOut("-debug             : display the size of the encoded block pre-entropy coding", true)
		printOut("-silent            : silent mode: no output (except warnings and errors)", true)
		printOut("-overwrite         : overwrite the output file if it already exists", true)
		printOut("-input=<filename>  : mandatory name of the input file to encode", true)
		printOut("-output=<filename> : optional name of the output file (defaults to <input.knz>)", true)
		printOut("-block=<size>      : size of the block (max 16 MB / default 100 KB)", true)
		printOut("-entropy=          : Entropy codec to use [None|Huffman|Range|PAQ|FPAQ]", true)
		os.Exit(0)
	}

	if *silent == true && *debug == true {
		printOut("Warning: both 'silent' and 'debug' options were selected, ignoring 'debug'", true)
		*debug = false
	}

	if len(*inputName) == 0 {
		fmt.Printf("Missing input file name, exiting ...\n")
		os.Exit(io.ERR_MISSING_FILENAME)
	}

	if len(*outputName) == 0 {
		*outputName = *inputName + ".knz"
	}

	this.debug = *debug
	this.silent = *silent
	this.overwrite = *overwrite
	this.inputName = *inputName
	this.outputName = *outputName
	this.blockSize = uint(*blockSize)
	this.entropyCodec = strings.ToUpper(*entropy)
	return this, nil
}

func main() {
	bc, err := NewBlockCompressor()

	if err != nil {
		fmt.Printf("Failed to create block compressor: %v\n", err)
		os.Exit(io.ERR_CREATE_COMPRESSOR)
	}

	code, _ := bc.call()
	os.Exit(code)
}

// Return exit code, number of bits written
func (this *BlockCompressor) call() (int, uint64) {
	var msg string
	printOut("Input file name set to '"+this.inputName+"'", this.debug)
	printOut("Output file name set to '"+this.outputName+"'", this.debug)
	msg = fmt.Sprintf("Block size set to %d", this.blockSize)
	printOut(msg, this.debug)
	msg = fmt.Sprintf("Debug set to %t", this.debug)
	printOut(msg, this.debug)
	msg = fmt.Sprintf("Overwrite set to %t", this.overwrite)
	printOut(msg, this.debug)
	ecodec := "no"

	if this.entropyCodec != "NONE" {
		ecodec = this.entropyCodec
	}

	msg = fmt.Sprintf("Using %s entropy codec", ecodec)
	printOut(msg, this.debug)
	written := uint64(0)
	output, err := os.OpenFile(this.outputName, os.O_RDWR, 666)

	if err == nil {
		// File exists
		if this.overwrite == false {
			fmt.Print("The output file exists and the 'overwrite' command ")
			fmt.Println("line option has not been provided")
			output.Close()
			return io.ERR_OVERWRITE_FILE, written
		}
	} else {
		// File does not exist, create
		output, err = os.Create(this.outputName)

		if err != nil {
			fmt.Printf("Cannot open output file '%v' for writing: %v\n", this.outputName, err)
			return io.ERR_CREATE_FILE, written
		}
	}

	defer output.Close()
	debugWriter := os.Stdout

	if this.debug == false {
		debugWriter = nil
	}

	cos, err := io.NewCompressedOutputStream(this.entropyCodec, output, this.blockSize, debugWriter)

	if err != nil {
		if err.(*io.IOError) != nil {
			fmt.Printf("%s\n", err.(*io.IOError).Message())
			return err.(*io.IOError).ErrorCode(), written
		} else {
			fmt.Printf("Cannot create compressed stream: %v\n", err)
			return io.ERR_CREATE_COMPRESSOR, written
		}
	}

	defer cos.Close()
	input, err := os.Open(this.inputName)

	if err != nil {
		fmt.Printf("Cannot open input file '%v': %v\n", this.inputName, err)
		return io.ERR_OPEN_FILE, written
	}

	defer input.Close()

	// Encode
	delta := int64(0)
	len := 0
	read := int64(0)
	printOut("Encoding ...", !this.silent)
	written = cos.GetWritten()
	buffer := make([]byte, DEFAULT_BUFFER_SIZE)
	len, err = input.Read(buffer)

	for len > 0 {
		if err != nil {
			fmt.Printf("Failed to read the next chunk of input file '%v': %v\n", this.inputName, err)
			return io.ERR_READ_FILE, written
		}

		read += int64(len)
		before := time.Now()

		if _, err = cos.Write(buffer[0:len]); err != nil {
			if ioerr, isIOErr := err.(*io.IOError); isIOErr == true {
				fmt.Printf("%s\n", ioerr.Message())
				return ioerr.ErrorCode(), written
			} else {
				fmt.Printf("Error in block codec forward(): %v\n", err)
				return io.ERR_PROCESS_BLOCK, written
			}
		}

		after := time.Now()
		delta += after.Sub(before).Nanoseconds()
		len, err = input.Read(buffer)
	}

	if read == 0 {
		fmt.Println("Empty input file ... nothing to do")
		return WARN_EMPTY_INPUT, written
	}

	// Close streams to ensure all data are flushed
	// Deferred close is fallback for error paths
	cos.Close()

	delta /= 1000000 // convert to ms

	printOut("", !this.silent)
	msg = fmt.Sprintf("Encoding:         %d ms", delta)
	printOut(msg, !this.silent)
	msg = fmt.Sprintf("Input size:       %d", read)
	printOut(msg, !this.silent)
	msg = fmt.Sprintf("Output size:      %d", cos.GetWritten())
	printOut(msg, !this.silent)
	msg = fmt.Sprintf("Ratio:            %f", float64(cos.GetWritten())/float64(read))
	printOut(msg, !this.silent)

	if delta > 0 {
  		msg = fmt.Sprintf("Throughput (KB/s) %d", ((read*int64(1000))>>10)/delta)
		printOut(msg, !this.silent)
	}
	
	printOut("", !this.silent)
	return 0, cos.GetWritten()
}

func printOut(msg string, print bool) {
	if print == true {
		fmt.Println(msg)
	}
}