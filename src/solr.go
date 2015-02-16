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
  fieldsData(&Index);
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
  fieldData []DocFieldDataInfo
}

type DocFieldDataInfo struct {
  fieldNum uint64
  filedBits [8]byte
  value string
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
      for ; size < 8; {
        var number = new(big.Int).SetBytes(buf[size:size+8]).Uint64() 
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

  var offset int64 = 0
  count := 1

  for {
    
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

    docFieldData.fieldCount = value;
    offset += int64(read)

    if value != 0 { // has field
      docFieldData.fieldData  = make([]DocFieldDataInfo, value)

      offset += int64(read)

      // compute all field data
      num := uint64(0)
      for ; num < value; {

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
        fieldDataInfo.fieldNum = value;        
        offset += int64(read)
        // TODO Bits
        offset += 1 
        // Value        
        bufValueSize := make([]byte, binary.MaxVarintLen64)
        bytesValueSize, err := fi.ReadAt(bufValueSize, offset)
        // string length
        size, readValueSize := binary.Uvarint(bufValueSize[0:bytesValueSize])
        offset += int64(readValueSize)        
        // string value
        bufValue := make([]byte, size)
        bytesValue, err := fi.ReadAt(bufValue, offset)
        fieldDataInfo.value = string(bufValue)
        offset += int64(bytesValue)

        docFieldData.fieldData[num] = fieldDataInfo;
                
        num += 1        
      }           

      count += 1
      
    }    
    
  }


}

