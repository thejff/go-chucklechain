package key

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
)

type Key[T any] interface {
	Sign(data T) (signature string, hash string, SignError error)
	GetPrivatePEM() string
	GetPublicPEM() string
	Verify(signature []byte, checksum []byte) error
}

type rsaKey[T any] struct {
	PrivateKey *rsa.PrivateKey
	PublicPEM  string
	PrivatePEM string
}

const (
	bitSize = 2048
)

func NewRSA[T any]() (Key[T], error) {
	key := rsaKey[T]{}

	reader := rand.Reader

	k, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return &key, err
	}

	prv := privateToPemRSA(k)

	pub := publicToPemRSA(k)

	key = rsaKey[T]{
		PrivateKey: k,
		PublicPEM:  pub,
		PrivatePEM: prv,
	}

	return &key, nil
}

func LoadRSA[T any](prvPem string) (Key[T], error) {
	block, _ := pem.Decode([]byte(prvPem))

	if block == nil {
		return &rsaKey[T]{}, errors.New("failed to decode pem to rsa key")
	}

	p, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return &rsaKey[T]{}, err
	}

	pubPem := publicToPemRSA(p)

	k := &rsaKey[T]{
		PrivateKey: p,
		PrivatePEM: prvPem,
		PublicPEM:  pubPem,
	}

	return k, nil
}

func (r *rsaKey[T]) Sign(data T) (string, string, error) {
	bData, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	h := sha256.New()
	h.Write(bData)
	digest := h.Sum(nil)

	sig, err := r.PrivateKey.Sign(rand.Reader, digest, crypto.SHA256)
	if err != nil {
		return "", "", err
	}

	if err = r.Verify(sig, digest); err != nil {
		return "", "", nil
	}

	return string(sig), string(digest), nil
}

func (r *rsaKey[T]) Verify(sig []byte, check []byte) error {
	return rsa.VerifyPKCS1v15(&r.PrivateKey.PublicKey, crypto.SHA256, check, sig)
}

func (r *rsaKey[T]) GetPrivatePEM() string {
	return r.PrivatePEM
}

func (r *rsaKey[T]) GetPublicPEM() string {
	return r.PublicPEM
}

func privateToPemRSA(k *rsa.PrivateKey) string {
	bytes := x509.MarshalPKCS1PrivateKey(k)

	pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: bytes,
		},
	)

	return string(pem)
}

func publicToPemRSA(k *rsa.PrivateKey) string {
	bPub := x509.MarshalPKCS1PublicKey(&k.PublicKey)
	// Above might not work with Activeledger, if so change PKCS1 to PKIX

	pubPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: bPub,
		},
	)

	return string(pubPem)
}
