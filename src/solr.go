package main

import (
	//"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"os"
)

func main() {
	Index := Index{}
	fieldsData(&Index)
	//fieldsPosition(&Index);

}

type Index struct {
	fieldsPosition []Position
}

type Position struct {
	value uint64
}

type DocFieldData struct {
	fieldCount uint64
	fieldData  []DocFieldDataInfo
}

type DocFieldDataInfo struct {
	fieldNum  uint64
	filedBits [8]byte
	value     string
}

func fieldsPosition(index *Index) {
	// open input file
	fi, err := os.Open("dump_solr_allcountries_v2/data/index/_9ea.fdx")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	positions := make([]Position, 128)

	// make a buffer to keep chunks that are read
	for {
		// read a chunk
		buf := make([]byte, 8)
		bytes, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if bytes == 0 {
			break
		}

		size := 0
		for size < 8 {
			var number = new(big.Int).SetBytes(buf[size : size+8]).Uint64()
			position := Position{number}
			positions = append(positions, position)
			//fmt.Println("%v", position.value)
			//fmt.Printf("%v", buf[size:size+8])
			size += 8
		}
	}

	index.fieldsPosition = positions
	fmt.Println("%v", len(index.fieldsPosition))
}

func fieldsData(index *Index) {
	// open input file
	fi, err := os.Open("dump_solr_allcountries_v2/data/index/_9ea.fdt")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	var offset int64 = 4
	count := 1

	// see first chars
	first := make([]byte, 500)
	firstBytes, err := fi.ReadAt(first, offset)
	fmt.Println("firstBytes %v", firstBytes)
	fmt.Println("firstBytes %v", first[0:500])

	for {
		fmt.Println("OFFSET %v", offset)

		docFieldData := DocFieldData{}

		// read a chunk
		buf := make([]byte, binary.MaxVarintLen64)
		bytes, err := fi.ReadAt(buf, offset)

		if err != nil && err != io.EOF {
			panic(err)
		}
		if bytes == 0 {
			break
		}

		// fmt.Println(offset)

		// FieldCount
		value, read := binary.Uvarint(buf[0:bytes])
		docFieldData.fieldCount = value
		offset += int64(read)
		fmt.Println("*** VALUE %v", value)
		fmt.Println("OFFSET %v", offset)
		//fmt.Println("READED %v", read)

		if value != 0 { // has field
			docFieldData.fieldData = make([]DocFieldDataInfo, value)

			// compute all field data
			num := uint64(0)
			fmt.Println("FIELD COUNT", value)
			for num < value {
				fmt.Println("*********** TOKEN ***************")
				//fmt.Println("FIELD_COUNT %v", num)

				fieldDataInfo := DocFieldDataInfo{}

				//newOffset, err :=fi.Seek(offset, 0)
				buf := make([]byte, binary.MaxVarintLen64)
				bytes, err := fi.ReadAt(buf, offset)

				if err != nil && err != io.EOF {
					panic(err)
				}
				if bytes == 0 {
					break
				}

				// FieldNum
				value, read := binary.Uvarint(buf[0:bytes])
				fieldDataInfo.fieldNum = value
				offset += int64(read)

				/*bufBits := make([]byte, 1)
				bytesBits, err := fi.ReadAt(bufBits, offset)

				bytesBits += 1*/
				/*if bufBits[1] != INDEXED_FIELDS {
					fmt.Println("binary")
				}*/
				//fmt.Println(bufBits[0]&COMP_FIELDS == COMP_FIELDS)
				//fmt.Println(bytesBits)
				offset += 1

				// Value
				bufValueSize := make([]byte, binary.MaxVarintLen64)
				bytesValueSize, err := fi.ReadAt(bufValueSize, offset)

				if err != nil && err != io.EOF {
					panic(err)
				}
				if bytes == 0 {
					break
				}
				// string length
				size, readValueSize := binary.Uvarint(bufValueSize[0:bytesValueSize])
				//fmt.Println("FIELD_LENGTH_VALUE %v", bufValueSize[0:bytesValueSize])

				//fmt.Println("FIELD_LENGTH_READ_SIZE %v", readValueSize)
				//fmt.Println("FIELD_LENGTH_READ %v", readValueSize)
				offset += int64(readValueSize)
				//fmt.Println("readValueSize", readValueSize)
				// string value
				bufValue := make([]byte, size)
				bytesValue, err := fi.ReadAt(bufValue, offset)

				if err != nil && err != io.EOF {
					panic(err)
				}
				if bytes == 0 {
					break
				}

				fmt.Println("VALUE %v ", string(bufValue))

				fieldDataInfo.value = string(bufValue)
				offset += int64(bytesValue)

				docFieldData.fieldData[num] = fieldDataInfo

				num += 1
			}

			//fmt.Println("%v", docFieldData)

			count += 1
			if count == 4 {
				break
			}

		}

	}

}
