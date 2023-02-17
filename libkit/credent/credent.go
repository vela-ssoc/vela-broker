package credent

import "crypto/tls"

// Certifier TLS 证书管理模块
type Certifier interface { // Certifier
	// Match 根据 tls.ClientHelloInfo 获取证书
	// 该接口为了兼容 tls.Config 下的 GetCertificate
	// https://github.com/golang/go/blob/76d39ae3499238ac7efb731f4f4cd47b1b3288ab/src/crypto/tls/common.go#L554-L563
	Match(*tls.ClientHelloInfo) (*tls.Certificate, error)

	// Replace 替换证书，能够在运行中替换升级证书
	Replace(certPEMBlock, keyPEMBlock []byte) error
}

// Single 单证书
func Single(certPEMBlock, keyPEMBlock []byte) (Certifier, error) {
	sl := new(single)
	if err := sl.Replace(certPEMBlock, keyPEMBlock); err != nil {
		return nil, err
	}

	return sl, nil
}

// single 单证书
type single struct {
	certPEMBlock []byte
	keyPEMBlock  []byte
	cert         *tls.Certificate
}

// Match 匹配证书
func (sl *single) Match(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return sl.cert, nil
}

// Replace 替换证书
func (sl *single) Replace(certPEMBlock, keyPEMBlock []byte) error {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err == nil {
		sl.cert = &cert
		sl.certPEMBlock = certPEMBlock
		sl.keyPEMBlock = keyPEMBlock
	}

	return err
}
