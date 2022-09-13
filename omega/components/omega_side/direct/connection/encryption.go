package connection

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

// EncryptionSession is a session unique to a player, that handles encryption between the server and the
// client.
type EncryptionSession struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Salt       []byte

	sharedSecret   []byte
	secretKeyBytes [32]byte
	cipherBlock    cipher.Block

	encryptIV []byte
	decryptIV []byte
}

// Init initialises the encryption session, computing the shared secret and secret key bytes as required to
// initialise the cipher blocks for encryption and decryption.
func (session *EncryptionSession) Init() error {
	session.computeSharedSecret()
	return session.computeIVs()
}

// computeSharedSecret computes the shared secret required for encryption and decryption.
func (session *EncryptionSession) computeSharedSecret() {
	// We only care about the 'x' part of this.
	x, _ := session.PublicKey.Curve.ScalarMult(session.PublicKey.X, session.PublicKey.Y, session.PrivateKey.D.Bytes())
	//fmt.Println(x)
	session.sharedSecret = x.Bytes()
}

// computeIVs computes the IVs and cipher required to start encrypting and decrypting.
func (session *EncryptionSession) computeIVs() error {
	var err error

	// First compute the secret key bytes, which is a hash of the Salt and the shared secret computed using
	// the method above.
	first12 := append([]byte(nil), session.sharedSecret[:12]...)
	sec := append(first12, 0, 0, 1, 228)
	session.secretKeyBytes = sha256.Sum256(append(sec, session.Salt...))
	session.cipherBlock, err = aes.NewCipher(session.secretKeyBytes[:])
	if err != nil {
		return fmt.Errorf("error creating AES cipher: %v", err)
	}

	session.encryptIV = append([]byte{}, session.secretKeyBytes[:aes.BlockSize]...)
	session.decryptIV = append([]byte{}, session.secretKeyBytes[:aes.BlockSize]...)
	return nil
}

// Encrypt encrypts the data passed in the slice itself.
func (session *EncryptionSession) Encrypt(data []byte) {
	for i := range data {
		cipherFeedback := cipher.NewCFBEncrypter(session.cipherBlock, session.encryptIV)
		cipherFeedback.XORKeyStream(data[i:i+1], data[i:i+1])
		session.encryptIV = append(session.encryptIV[1:], data[i])
	}
}

// Decrypt decrypts the data passed in the slice itself.
func (session *EncryptionSession) Decrypt(data []byte) {
	for i, b := range data {
		cipherFeedback := cipher.NewCFBDecrypter(session.cipherBlock, session.decryptIV)
		cipherFeedback.XORKeyStream(data[i:i+1], data[i:i+1])
		session.decryptIV = append(session.decryptIV[1:], b)
	}
}
