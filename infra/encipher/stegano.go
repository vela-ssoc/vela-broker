package encipher

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
)

const payloadSize = 2 + sha1.Size // 2 位长度(uint16) + 20 位校验码

var (
	ErrDataTooLong  = errors.New("加密数据太长")
	ErrDataChecksum = errors.New("数据校验错误")
	ErrNotFoundData = errors.New("没有找到数据")
)

// WriteFile 将数据加密写入到文件尾部
func WriteFile(name string, v any) error {
	data, err := EncryptJSON(v)
	if err != nil {
		return err
	}

	open, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = open.Write(data)
	_ = open.Close()

	return err
}

// ReadFile 从文件中读取加密数据
func ReadFile(name string, v any) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	var filesize int64
	if stat, ex := file.Stat(); ex == nil {
		filesize = stat.Size()
	}

	if filesize < payloadSize {
		return ErrNotFoundData
	}

	if _, err = file.Seek(-payloadSize, io.SeekEnd); err != nil {
		return err
	}
	payload := make([]byte, payloadSize)
	if _, err = file.Read(payload); err != nil {
		return err
	}
	size := binary.LittleEndian.Uint16(payload) // 获取数据长度
	if _, err = file.Seek(-(int64(size) + payloadSize), io.SeekEnd); err != nil {
		return err
	}
	data := make([]byte, size)
	if _, err = io.ReadFull(file, data); err != nil {
		return err
	}

	dec := instance.decrypt(data)
	sum := sha1.Sum(dec)
	if !bytes.Equal(sum[:], payload[2:]) {
		return ErrDataChecksum
	}

	return json.Unmarshal(dec, v)
}

func EncryptJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	sum := sha1.Sum(raw)
	enc := instance.encrypt(raw)
	size := len(enc)
	if size > math.MaxUint16 {
		return nil, ErrDataTooLong
	}

	data := make([]byte, size+payloadSize)
	copy(data, enc)
	binary.LittleEndian.PutUint16(data[size:], uint16(size))
	copy(data[size+2:], sum[:])

	return data, nil
}

func DecryptJSON(enc []byte, v any) error {
	size := len(enc)
	if size < payloadSize {
		return ErrNotFoundData
	}

	payload := enc[size-payloadSize:]
	dsz := int(binary.LittleEndian.Uint16(payload))
	if dsz > size-payloadSize {
		return ErrNotFoundData
	}
	data := enc[size-dsz-payloadSize : size-payloadSize]
	raw := instance.decrypt(data)
	sum := sha1.Sum(raw)
	if !bytes.Equal(sum[:], payload[2:]) {
		return ErrDataChecksum
	}

	return json.Unmarshal(raw, v)
}

func MultiReader(r io.Reader, v any) (io.Reader, int, error) {
	data, err := EncryptJSON(v)
	if err != nil {
		return nil, 0, err
	}

	br := bytes.NewReader(data)
	size := len(data)

	multi := io.MultiReader(r, br)
	as := &multiRead{r: multi}

	return as, size, err
}

type multiRead struct {
	r io.Reader
}

func (m *multiRead) Read(p []byte) (int, error) {
	return m.r.Read(p)
}
