package nuxeo

import (
	"bytes"
	"encoding/pem"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pavel-v-chernykh/keystore-go"
)

// Generates a random 12-position password, encodes the PEM data in the cert arg and returns:
// JKS bytes, password, error
func toTrustStoreFromBytes(cert []byte)([]byte, string, error) {
	pass := genPass()
	if store, err := toTrustStore(cert, pass); err != nil {
		return nil, "", err
	} else {
		return store, pass, nil
	}
}

// Decodes PEM-encoded certificates in the passed cert arg and returns the data encoded as a JKS truststore. This
// function was cloned from https://github.com/redhat-cop/cert-utils-operator. A copy of their license is in
// resources/licenses/cert-utils/LICENSE. Note - this function uses the pavel-v-chernykh go library to do the actual
// JKS encoding and that lib doesn't support P12. Hence JKS.
func toTrustStore(cert []byte, password string) ([]byte, error) {
	keyStore := keystore.KeyStore{}
	i := 0
	for block, rest := pem.Decode(cert); block != nil; block, rest = pem.Decode(rest) {
		keyStore["alias"+strconv.Itoa(i)] = &keystore.TrustedCertificateEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			Certificate: keystore.Certificate{
				Type:    "X.509",
				Content: block.Bytes,
			},
		}
		i++
	}
	buffer := bytes.Buffer{}
	err := keystore.Encode(&buffer, keyStore, []byte(password))
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

// 12-position random password generator. Courtesy of https://yourbasic.org/golang/generate-random-string/
func genPass() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < 12; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

