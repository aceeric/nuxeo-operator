package nuxeo

import (
	"bytes"
	"encoding/pem"
	goerrors "errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pavel-v-chernykh/keystore-go"
)

const (
	aliasStr = "alias"
	x509str  = "X.509"
)

// Keystore/truststore functions in this module were modeled after https://github.com/redhat-cop/cert-utils-operator.
// A copy of their license is in resources/licenses/cert-utils/LICENSE. Note - this code uses the pavel-v-chernykh
// go library to do the actual JKS encoding and that lib doesn't support P12. Hence JKS.

// Generates a random 12-position password, converts PEM-encoded cert arg into a trust store and returns
// JKS bytes, password, error
func trustStoreFromPEM(cert []byte) ([]byte, string, error) {
	pass := genPass()
	if store, err := toTrustStore(cert, pass); err != nil {
		return nil, "", err
	} else {
		return store, pass, nil
	}
}

// Generates a random 12-position password, converts PEM-encoded cert & private key args into a key store and returns
// JKS bytes, password, error
func keyStoreFromPEM(cert []byte, privateKey []byte) ([]byte, string, error) {
	pass := genPass()
	if store, err := toKeyStore(cert, privateKey, pass); err != nil {
		return nil, "", err
	} else {
		return store, pass, nil
	}
}

// Decodes PEM-encoded cert and private key args and returns the data encoded as a JKS key store.
func toKeyStore(cert []byte, privateKey []byte, password string) ([]byte, error) {
	store := keystore.KeyStore{}
	var certs []keystore.Certificate
	for block, rest := pem.Decode(cert); block != nil; block, rest = pem.Decode(rest) {
		certs = append(certs, keystore.Certificate{
			Type:    x509str,
			Content: block.Bytes,
		})
	}
	if block, _ := pem.Decode(privateKey); block == nil {
		return nil, goerrors.New("failed to decode passed key")
	} else {
		if !strings.Contains(block.Type, "PRIVATE KEY") {
			return nil, goerrors.New("passed key does not appear to be a PRIVATE KEY")
		}
		store[aliasStr] = &keystore.PrivateKeyEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			PrivKey:   block.Bytes,
			CertChain: certs,
		}
	}
	buffer := bytes.Buffer{}
	if err := keystore.Encode(&buffer, store, []byte(password)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decodes PEM-encoded cert arg and returns the data encoded as a JKS trust store.
func toTrustStore(cert []byte, password string) ([]byte, error) {
	store := keystore.KeyStore{}
	i := 0
	for block, rest := pem.Decode(cert); block != nil; block, rest = pem.Decode(rest) {
		store[aliasStr+strconv.Itoa(i)] = &keystore.TrustedCertificateEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			Certificate: keystore.Certificate{
				Type:    x509str,
				Content: block.Bytes,
			},
		}
		i++
	}
	buffer := bytes.Buffer{}
	if err := keystore.Encode(&buffer, store, []byte(password)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// 12-position alphanumeric random password generator. Courtesy of https://yourbasic.org/golang/generate-random-string/
func genPass() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < 12; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
