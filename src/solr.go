package main

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"os"
)

func main() {
	Index := Index{}
	//fieldsData(&Index)
	//fieldsPosition(&Index)
	termDictionary(&Index)
}

type Index struct {
	fieldsPosition []Position
	terms          []TermInfo
}

type Position struct {
	value uint64
}

type DocFieldData struct {
	fieldCount uint64
	fieldData  []DocFieldDataInfo
}

type TermInfo struct {
	prefixLength uint64
	suffix       string
	fieldNum     uint64
	freqDelta    uint64
	docFreq      uint64
	proxDelta    uint64
	skipDelta    uint64
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
	fiStat, err := fi.Stat()
	var filesize int64 = fiStat.Size()

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
	var count int64 = 1

	/*// see first chars
	first := make([]byte, 500)
	firstBytes, err := fi.ReadAt(first, offset)*/

	for {

		// create new doc
		docFieldData := DocFieldData{}

		// read field count
		value, read := readVarIntBuf(fi, offset)
		docFieldData.fieldCount = value

		offset += int64(read)

		if value != 0 { // has field
			docFieldData.fieldData = make([]DocFieldDataInfo, value)

			// compute all field data
			num := uint64(0)

			for num < value {

				//fmt.Println("FIELD_COUNT %v", num)

				fieldDataInfo := DocFieldDataInfo{}

				// FieldNum
				value, read := readVarIntBuf(fi, offset)
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

				// string length
				size, readValueSize := readVarIntBuf(fi, offset)
				offset += int64(readValueSize)

				// string value
				bufValue, bytesValue := readBytesBuf(fi, offset, size)

				//fmt.Println("VALUE %v ", string(bufValue))

				fieldDataInfo.value = string(bufValue)
				offset += int64(bytesValue)

				docFieldData.fieldData[num] = fieldDataInfo

				num += 1
			}

			//fmt.Println("%v", docFieldData)

			count += 1

			if count > filesize-10 {
				fmt.Println("end")
				break
			}

		}

	}
}

func termDictionary(index *Index) {
	// open input file
	fi, err := os.Open("dump_solr_allcountries_v2/data/index/_9ea.tis")
	fiStat, err := fi.Stat()
	var filesize int64 = fiStat.Size()

	fmt.Println("filesize %v", filesize)

	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	var offset int64 = 0
	var count int64 = 1

	// 1)
	// TIVersion, TermCount, IndexInterval, SkipInterval, MaxSkipLevels
	TIVersion := readUInt32Buf(fi, offset)
	fmt.Println(TIVersion)
	offset += 4
	TermCount := readUInt64Buf(fi, offset)
	fmt.Println(TermCount)
	offset += 8
	IndexInterval := readUInt32Buf(fi, offset)
	fmt.Println(IndexInterval)
	offset += 4
	SkipInterval := readUInt32Buf(fi, offset)
	fmt.Println(SkipInterval)
	offset += 4
	MaxSkipLevels := readUInt32Buf(fi, offset)
	fmt.Println(MaxSkipLevels)
	offset += 4

	terms := make([]TermInfo, TermCount)

	// 2)
	/*
	  TermInfos --> <TermInfo> TermCount
	  TermInfo --> <Term, DocFreq, FreqDelta, ProxDelta, SkipDelta>
	  Term --> <PrefixLength, Suffix, FieldNum>
	  Suffix --> String
	*/
	for {

		// TERM
		PrefixLength, read := readVarIntBuf(fi, offset)
		offset += int64(read)
		// string length
		size, readSuffixSize := readVarIntBuf(fi, offset)
		offset += int64(readSuffixSize)
		// string value
		Suffix, bytesValue := readBytesBuf(fi, offset, size)
		offset += int64(bytesValue)
		// FieldNum
		FieldNum, FieldNumSize := readVarIntBuf(fi, offset)
		offset += int64(FieldNumSize)
		// DocFreq
		DocFreq, DocFreqSize := readVarIntBuf(fi, offset)
		offset += int64(DocFreqSize)
		// FreqDelta
		FreqDelta, FreqDeltaSize := readVarIntBuf(fi, offset)
		offset += int64(FreqDeltaSize)
		// ProxDelta
		ProxDelta, ProxDeltaSize := readVarIntBuf(fi, offset)
		offset += int64(ProxDeltaSize)
		// SkipDelta
		SkipDelta, SkipDeltaSize := readVarIntBuf(fi, offset)
		offset += int64(SkipDeltaSize)

		termInfo := TermInfo{PrefixLength, string(Suffix), FieldNum, DocFreq, FreqDelta, ProxDelta, SkipDelta}
		terms = append(terms, termInfo)

		if offset > filesize {
			fmt.Println("end %v", offset)
			break
		}

		if count < 50 {
			fmt.Println("terminfo %v", termInfo.suffix)
			fmt.Println("offset %v", offset)
		}

		count += 1

	}

	index.terms = terms
	fmt.Println("len length ", len(terms))
}

// UTILS
func readBytesBuf(file *os.File, offset int64, size uint64) ([]byte, int) {
	var buf = make([]byte, size)
	bytes, err := file.ReadAt(buf, offset)

	if err != nil && err != io.EOF {
		panic(err)
	}

	if bytes == 0 {

	}

	return buf, bytes
}

func readVarIntBuf(file *os.File, offset int64) (uint64, int) {
	var buf = make([]byte, binary.MaxVarintLen64)
	bytes, err := file.ReadAt(buf, offset)

	if err != nil && err != io.EOF {
		panic(err)
	}

	return binary.Uvarint(buf[0:bytes])
}

func read_int32(data []byte) (ret int32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

func readUInt32Buf(file *os.File, offset int64) uint32 {
	var buf = make([]byte, 4)
	bytes, err := file.ReadAt(buf, offset)

	if err != nil && err != io.EOF {
		panic(err)
	}

	return binary.BigEndian.Uint32(buf[0:bytes])
}

func readUInt64Buf(file *os.File, offset int64) uint64 {
	var buf = make([]byte, 8)
	bytes, err := file.ReadAt(buf, offset)

	if err != nil && err != io.EOF {
		panic(err)
	}

	return binary.BigEndian.Uint64(buf[0:bytes])
}
