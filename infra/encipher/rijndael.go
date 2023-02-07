package encipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// instance 加/解密机（单例）
var instance = newInstance()

// newInstance 创建加解密对象
func newInstance() *machine {
	key := [32]byte{
		147, 70, 122, 151, 130, 4, 51, 110, 132, 118, 247, 103, 147, 117, 79, 233,
		120, 156, 18, 133, 111, 227, 185, 17, 128, 174, 201, 123, 64, 84, 98, 139,
	}
	block, err := aes.NewCipher(key[:]) // 请确保 len(key) 长度是 16/24/32
	if err != nil {
		panic(err)
	}

	return &machine{block: block}
}

type machine struct {
	block cipher.Block
}

// encrypt 加密方法
func (mh *machine) encrypt(plain []byte) []byte {
	// 生成一个随机向量，写在密文头部，同样的明文每次加密结果都不同
	bsz := aes.BlockSize
	dst := make([]byte, len(plain)+bsz)
	_, _ = rand.Reader.Read(dst[:bsz])

	cipher.NewCFBEncrypter(mh.block, dst[:bsz]).XORKeyStream(dst[bsz:], plain)

	return dst
}

// decrypt 解密方法
func (mh *machine) decrypt(ctxt []byte) []byte {
	bsz := aes.BlockSize
	size := len(ctxt)
	if size < bsz {
		return nil
	}
	iv := ctxt[:bsz]

	dst := make([]byte, size-bsz)
	cipher.NewCFBDecrypter(mh.block, iv).XORKeyStream(dst, ctxt[bsz:])

	return dst
}
