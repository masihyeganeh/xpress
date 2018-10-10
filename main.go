package xpress

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type xpressHeader struct {
	DestSize   uint32
	SourceSize uint32
}

func Decompress(inputBuffer []byte) ([]byte, error) {
	processed := 0
	header := xpressHeader{}
	decompressed := []byte("")

	for processed < len(inputBuffer) {
		binary.Read(bytes.NewReader(inputBuffer[processed:]), binary.LittleEndian, &header)
		if len(inputBuffer) < int(header.SourceSize) + 8 {
			return nil, errors.New("corrupted data")
		}
		part := DecompressRaw(inputBuffer[8+processed:8+processed+int(header.SourceSize)], int(header.DestSize))
		decompressed = append(decompressed, part...)
		processed += 8 + int(header.SourceSize)
	}

	return decompressed, nil
}

func DecompressRaw(inputBuffer []byte, outputSize int) []byte {
	outputIndex, inputIndex, indicator, indicatorBit, length, offset, nibbleIndex := 0, 0, 0, 0, 0, 0, 0
	inputSize := len(inputBuffer)
	outputBuffer := make([]byte, outputSize)

	for (outputIndex < outputSize) && (inputIndex < inputSize) {
		if indicatorBit == 0 {
			if inputIndex+3 >= inputSize {
				goto Done
			}
			indicator = int(binary.LittleEndian.Uint32(inputBuffer[inputIndex : inputIndex+4]))
			inputIndex += 4
			indicatorBit = 32
		}
		indicatorBit--

		// check whether the bit specified by IndicatorBit is set or not
		// set in Indicator. For example, if IndicatorBit has value 4
		// check whether the 4th bit of the value in Indicator is set

		if ((indicator >> uint(indicatorBit)) & 1) == 0 {
			if outputIndex >= outputSize {
				goto Done
			}
			outputBuffer[outputIndex] = inputBuffer[inputIndex]
			inputIndex += 1
			outputIndex++
		} else {
			if inputIndex+1 >= inputSize {
				goto Done
			}

			length = int(binary.LittleEndian.Uint16(inputBuffer[inputIndex : inputIndex+2]))

			inputIndex += 2

			offset = length / 8
			length = length % 8

			if length == 7 {
				if nibbleIndex == 0 {
					nibbleIndex = inputIndex
					if inputIndex >= inputSize {
						goto Done
					}
					length = int(inputBuffer[inputIndex] % 16)

					inputIndex += 1
				} else {
					length = int(inputBuffer[nibbleIndex] / 16)
					nibbleIndex = 0
				}

				if length == 15 {
					if inputIndex >= inputSize {
						goto Done
					}

					length = int(inputBuffer[inputIndex])
					inputIndex += 1

					if length == 255 {
						if inputIndex+2 >= inputSize {
							goto Done
						}
						length = int(binary.LittleEndian.Uint16(inputBuffer[inputIndex : inputIndex+2]))
						inputIndex += 2
						length -= 15 + 7
					}
					length += 15
				}
				length += 7
			}

			length += 3

			for ; length != 0; length-- {
				if (outputIndex >= outputSize) || ((offset + 1) > outputIndex) {
					break
				}
				outputBuffer[outputIndex] = outputBuffer[outputIndex-offset-1]
				outputIndex++
			}
		}
	}

Done:

	if outputIndex < len(outputBuffer) {
		outputBuffer = outputBuffer[:outputIndex]
	}

	return outputBuffer
}
