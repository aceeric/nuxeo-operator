package nuxeo

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/pem"
	goerrors "errors"
	"fmt"
	"testing"

	"github.com/pavel-v-chernykh/keystore-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Puts a PEM-encoded cert into a trust store, gets the cert out of the store, PEM encodes it and compares
// the result to the original
func (suite *certSuite) TestTrustStore() {
	originalPemEncodedCert := getCert()
	trustStore, pass, err := trustStoreFromPEM(originalPemEncodedCert)
	require.Nil(suite.T(), err, "trustStoreFromPEM failed")
	require.Equal(suite.T(), 12, len(pass), "Incorrect password generated")
	require.NotNil(suite.T(), trustStore, "trustStoreFromPEM failed")
	store, err := readStore(trustStore, pass)
	require.Nil(suite.T(), err, "Couldn't read trust store")

	var gotCert bool
	for _, e := range store {
		switch k := e.(type) {
		case *keystore.PrivateKeyEntry:
			require.Fail(suite.T(), "found private key in truststore")
		case *keystore.TrustedCertificateEntry:
			pemFromStore, err := trustedCertEntryToPEM(k, 1)
			require.Nil(suite.T(), err, err)
			require.Equal(suite.T(), originalPemEncodedCert, pemFromStore, "cert didn't encode or decode correctly")
			gotCert = true
		default:
			require.Fail(suite.T(),"Unexpected store entry")
		}
	}
	require.True(suite.T(), gotCert, "trust store missing cert")
}

// Puts a PEM-encoded cert and private key into a key store, gets them both out out of the store, PEM encodes them,
// and compares the results to the originals
func (suite *certSuite) TestKeyStore() {
	originalPemEncodedCert := getCert()
	originalPemEncodedPrivateKey := getPrivateKey()
	keyStore, pass, err := keyStoreFromPEM(originalPemEncodedCert, originalPemEncodedPrivateKey)
	require.Nil(suite.T(), err, "keyStoreFromPEM failed")
	require.Equal(suite.T(), 12, len(pass), "Incorrect password generated")
	require.NotNil(suite.T(), keyStore, "keyStoreFromPEM failed")
	store, err := readStore(keyStore, pass)
	require.Nil(suite.T(), err, "Couldn't read key store")

	var gotCert, gotKey bool
	for _, e := range store {
		switch k := e.(type) {
		case *keystore.PrivateKeyEntry:
			pemFromStore, err := privateKeyEntryToPEM(k)
			require.Nil(suite.T(), err, err)
			require.Equal(suite.T(), originalPemEncodedPrivateKey, pemFromStore, "private key didn't encode or decode correctly")
			gotKey = true
			pemFromStore, err = certificatesToPEM(k.CertChain, 1)
			require.Nil(suite.T(), err, err)
			require.Equal(suite.T(), originalPemEncodedCert, pemFromStore, "cert didn't encode or decode correctly")
			gotCert = true
		default:
			require.Fail(suite.T(),"Unexpected store entry")
		}
	}
	require.True(suite.T(), gotCert && gotKey, "keystore missing cert and/or key")
}

// certSuite is the Cert test suite structure
type certSuite struct {
	suite.Suite
	r ReconcileNuxeo
}

// SetupSuite initializes the Fake client, a ReconcileNuxeo struct, and various test suite constants
func (suite *certSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *certSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the Cert unit test suite. It is called by 'go test' and will call every
// function in this file with a certSuite receiver that begins with "Test..."
func TestCertUnitTestSuite(t *testing.T) {
	suite.Run(t, new(certSuite))
}

// returns a PEM-encoded private key. The newline at the end is important
func getPrivateKey() []byte {
	return []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAwcdWq8pvgOnpNI+c05qgnzAk4Ez8KIaWO73yEp+/NBK4yv4H
1j6dWl2BBIb0/NTrkfMmnT4mSYjesAUMzs9iWvVZV/1MtuAEKig1i+FQmB2hvByL
MRhZZKiUHyIShpY4Z0a4t+5QYYpqA71AK/6WYGdix3k6piJBzV1v+wB/cgh9/VUH
1VTaX7Lw9L4zmQ7ryl2H9SeTM+xWL7W2G/GsxH1EheWxOpZoGluSssnk+usykCqB
Xis6+kFarSJgKPZ+NUyppu87cZgl0ECz6GKwHGFhvKozCJ7GtnsLOLCc0gVVrD4t
u+Pwmuu8qREfe18W4kPPFUry3mNtU6ROAeaB/wIDAQABAoIBAG2mOmjjF80+jvNr
ADbMnG73cyZo6ZaU8ZXEmaHoOu1gWqiirhSRQcDMgCDrrN0ULmhbylHXxRp/FGNN
uD2eI+2MP44Gis5AXJruPb51NIGe4tHq5OhW+t52dbpYMVtuzWPDJOsPMvS+udZ7
1EAQw06xsbdl5cX0RH/Mi3zgfz0qjA0d65sMs24+fHCI+YOdnixwza6P2Kq5solZ
BT3nJFUKvLUiz0vWjcxURO+O4v9r7aZPWKAOK+nvLxACquVwl/5iSojw6lxubBMY
u1mcBglqWVxf3zzAGukqpzsWQW5OEbZrK29576wYCLlWh34M4twCydqxitbULCZ2
JXgofYECgYEAzNDLoddj30ckO45JU1LcIVzbo/jklDzmL771PlkkY4QitT/Gjd0n
3s+T/7241o8119/V0zMQLrQAeAYsW7SMTOMloKBy/BQQztIpPy/nrYvAKLyizP+F
lGcmX+M4dxoQAq+OsSqyjFdQdXFUCZYB7cUDmj/h24cb27NbXBZpN68CgYEA8jRy
sg8dT2fZRdKpfgXr3kRtI3jXSHP02xBGbZtXHV+B84mccWGIiVPShxW76GjJ8jai
K3PfymquZD7ZHTkM5/XTlRx/JtXxiS62aDIeutgW6utdqgrCfnw/C6nfORlaTgNq
vry6qWSbKX8mpO1eetCdB0MQ22sOqmB0mgMHnrECgYB8dFd4ZVhjoWgL5F78CbqH
b1Rro97JkOPSmXeORj6NVgp9Fl7Bb2Q9yObGnPNHNUhjf7j/l+S6bFholl+37dLf
GZuQqk6UjGDWO/AiXCqsUuIWHuHSLWZvEerIk1qJTMXzy9eqIibSjm/unUmSdZuA
bpnMzgqhCc1MyAS4xUl0MwKBgHHdSlZ/WI4uCi0THm+KpRp3HL/iXYNIUEJ0YkfB
EbFTZypw9UUwTxoQeBbdlttp+BaQrKi07u6gPKAQE83zNigOn4uoO/ar+cM+XK6b
cWrxj8SdJgl8yXbhPlpjX/fd/WBTpulInJBqJa/agPZkSVh/nnL9in081UYv1mFZ
L0nhAoGAIvb5Q+bwtp3tnYus29sIS/fLmgYSoJYNhswq/xpZ0VcKYaN+hoT1Zdze
GlAY9qJVXQe2QxwdbZQH+zGMlfUopAmtxYckVd99ncA0yOk6sMM3YB1dyJ6mJJ1R
u2rBeN52NlPTkZ3gQ8wefO/dNfL2y/Ld/pdu5XAJiRkiaAlkUms=
-----END RSA PRIVATE KEY-----
`)
}

// returns a PEM-encoded cert. The newline at the end is important
func getCert() []byte {
	return []byte(`-----BEGIN CERTIFICATE-----
MIIDszCCApugAwIBAgIUPa/Aj7nUztR8AXeowcff/oDUS5gwDQYJKoZIhvcNAQEL
BQAwaTELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1hcnlsYW5kMRIwEAYDVQQHDAlT
b21ld2hlcmUxCzAJBgNVBAoMAklUMSYwJAYDVQQDDB1udXhlby1zZXJ2ZXIuYXBw
cy1jcmMudGVzdGluZzAeFw0yMDA2MDgxNDQxMjVaFw0yMTA2MDgxNDQxMjVaMGkx
CzAJBgNVBAYTAlVTMREwDwYDVQQIDAhNYXJ5bGFuZDESMBAGA1UEBwwJU29tZXdo
ZXJlMQswCQYDVQQKDAJJVDEmMCQGA1UEAwwdbnV4ZW8tc2VydmVyLmFwcHMtY3Jj
LnRlc3RpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC9ed7SpdT9
LYOv/ri+zWIj1V1Ow9Zq6hNsLqIQ7zG6xIe/e55+DlNiRBkzB30un/cZ1wJS+mfz
zMuRc897/RmrsYY0j6erI8MiVkKCHSttlU3p9f25J8K9VdQNaE69cYIAi1LHNX0j
nmc0JWwifvftqI5pM/0txEZ8+209sAACuLloPt9UTcjTr374LOioJ3RjWvUBMxzT
WxRZAG/uVTCnIh22A/yE3CdkdKAszOTMNiev4MQELPYH+UWBTRTpNDvtIpLqENYH
4vk9egcQjJMywMckUj6oeDUOmJMkocKO5B+ad1+FgKH2Cpfn9OTCRjJGkfkE9d+r
50zKwskRlfodAgMBAAGjUzBRMB0GA1UdDgQWBBSRcZ+9mW1KVydnTEgvevy/cha4
UDAfBgNVHSMEGDAWgBSRcZ+9mW1KVydnTEgvevy/cha4UDAPBgNVHRMBAf8EBTAD
AQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAXtj9wogOSAdhdX86R9CqR6WpDXztzhO4r
C235G4MRol1U+zzoqbbol9SlYwPnNZQM6lOIKR5MGqpptBiQGXAiffEq3x3J92cC
6ujP7FOGcOmErNCoQgS2UvYeclsygU+F57BB7alWonVeZz+wxRIzVZMeJeoQ3+A/
eSasH6G2FH4moYQsuC+IRaFtgx0/0xhbyd3qO/qM9Rcw+gbCvX7WYTXV0csu6UO6
pzmqUK0KJcF2KWIafrIqgV0D5YJRLMm9SM9oGYThzVrtR0r0FmUmA5P3el/sCHn4
blF1aG6n/oXPSREBVFvWvJ/FStZSXLETVLPsUS0frX0ZDyu39xl4
-----END CERTIFICATE-----
`)
}

// Uses the keystore library to read the passed keystore/truststore and decode it into a
// keystore.KeyStore struct
func readStore(storeBytes []byte, password string) (keystore.KeyStore, error) {
	return keystore.Decode(bytes.NewReader(storeBytes), []byte(password))
}

// PEM-encodes the passed certificate(s). Returns an error if the expected cert count doesn't
// match the passed count.
func certificatesToPEM(certs [] keystore.Certificate, cnt int) ([]byte, error) {
	if len(certs) != cnt {
		return nil, goerrors.New(fmt.Sprintf("actual cert count %v did not match expected %v", len(certs), cnt))
	}
	var pemBytes []byte
	for _, cert := range certs {
		if x509Certs, err := x509.ParseCertificates(cert.Content); err != nil {
			return nil, err
		} else if b, err := pemEncodeX509Certs(x509Certs); err != nil {
			return nil, err
		} else {
			pemBytes = append(pemBytes, b...)
		}
	}
	return pemBytes,nil
}

// returns a byte array in PEM encoding of all the certs in the passed array
func pemEncodeX509Certs(certs []*x509.Certificate) ([]byte, error) {
	var pemBytes []byte
	for _, cert := range certs {
		if b, err := pemEncode("CERTIFICATE", cert.Raw); err != nil {
			return nil, err
		} else {
			pemBytes = append(pemBytes, b...)
		}
	}
	return pemBytes, nil
}

// PEM-encodes the passed certificate(s). Returns an error if the expected cert count doesn't
// match the passed count.
func trustedCertEntryToPEM(k *keystore.TrustedCertificateEntry, cnt int) ([]byte, error) {
	if certs, err := x509.ParseCertificates(k.Certificate.Content); err != nil {
		return nil, err
	} else {
		if len(certs) != cnt {
			return nil, goerrors.New(fmt.Sprintf("actual cert count %v did not match expected %v", len(certs), cnt))
		}
		return pemEncodeX509Certs(certs)
	}
}

// PEM-encodes the passed private key.
func privateKeyEntryToPEM(k *keystore.PrivateKeyEntry) ([]byte, error) {
	if pk, err := x509.ParsePKCS1PrivateKey(k.PrivKey); err != nil {
		return nil, err
	} else if keyBytes := x509.MarshalPKCS1PrivateKey(pk); keyBytes == nil {
		return nil, goerrors.New("x509.MarshalPKCS1PrivateKey failed")
	} else {
		return pemEncode("RSA PRIVATE KEY", keyBytes)
	}
}

// Does the PEM encoding
func pemEncode(blockType string, toEncode []byte) ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := pem.Encode(writer, &pem.Block{Type: blockType, Bytes: toEncode}); err != nil {
		return nil, err
	}
	_ = writer.Flush()
	return b.Bytes(), nil
}