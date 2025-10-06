package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"slices"
	"testing"
)

func TestMagicMd5(t *testing.T) {
	// 读取 test 文件，seek 到 0x1540 的位置，计算后面的 md5 值
	f, err := os.Open("./成语_官方.scel")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// 读取原文件的检验和
	chksum := make([]byte, 16)
	f.ReadAt(chksum, 0xC)

	// 计算 0x1540 之后的 md5 值
	f.Seek(0x1540, 0)
	state := CheckSumStream(f)
	chksum1 := make([]byte, 16)

	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(chksum1[4*i:], state[i])
	}

	result := slices.Compare(chksum, chksum1)
	fmt.Printf("original chksum: %X\n", chksum)
	fmt.Printf("my chksum: %X\n", chksum1)
	fmt.Printf("result: %v\n", result)
}
