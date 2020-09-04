package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

// GenerateRSAKey 生成RSA私钥和公钥，保存到文件中
func GenerateRSAKey(bits int, dir string) error {
	if dir == "" {
		dir = "."
	}

	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	//使用pem格式对x509输出的内容进行编码
	//创建文件保存私钥
	privateFile, err := os.Create(dir + "/private.pem")
	if err != nil {
		return err
	}
	defer privateFile.Close()
	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	//将数据保存到文件
	err = pem.Encode(privateFile, &privateBlock)
	if err != nil {
		return err
	}

	//保存公钥
	//获取公钥的数据
	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}
	//pem格式编码
	//创建用于保存公钥的文件
	publicFile, err := os.Create(dir + "/public.pem")
	if err != nil {
		return err
	}
	defer publicFile.Close()
	//创建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	//保存到文件
	err = pem.Encode(publicFile, &publicBlock)
	return err
}

// RsaEncryptByFile RSA加密
func RsaEncryptByFile(plainText []byte, path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return RsaEncrypt(plainText, buf)
}

// RsaEncrypt RSA加密
func RsaEncrypt(plainText, key []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(key)
	//x509解码

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		return nil, err
	}
	//返回密文
	return cipherText, nil
}

// RsaDecryptByFile 指定密钥文件 解密
func RsaDecryptByFile(cipherText []byte, path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return RsaDecrypt(cipherText, buf)
}

// RsaDecrypt RSA解密
func RsaDecrypt(cipherText, key []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(key)
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	//返回明文
	return plainText, nil
}
