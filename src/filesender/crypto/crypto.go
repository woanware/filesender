package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	config "filesender/config"
	helper "filesender/utils"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/pbkdf2"
)

// Returns an io.Reader that encrypts the byte stream from the given io.Reader
// using the given key and initialization vector.
func MakeEncrypterReader(key []byte, iv []byte, reader io.Reader) io.Reader {

	if key == nil {
		helper.OutputAndExit(fmt.Sprintf("Uninitialized key in makeEncrypterReader()"))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Unable to create AES cypher: %v", err))
	}

	if len(iv) != aes.BlockSize {
		helper.OutputAndExit(fmt.Sprintf("IV length %d != aes.BlockSize %d", len(iv), aes.BlockSize))
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	return &cipher.StreamReader{S: stream, R: reader}
}

//
func MakeDecryptionReader(key []byte, iv []byte, reader io.Reader) io.Reader {

	if key == nil {
		helper.OutputAndExit("Uninitialized key in MakeDecryptionReader()")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Unable to create AES cypher"))
	}

	if len(iv) != aes.BlockSize {
		helper.OutputAndExit(fmt.Sprintf("IV length %d != aes.BlockSize %d", len(iv), aes.BlockSize))
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	return &cipher.StreamReader{S: stream, R: reader}
}

// Decrypts the encrypted encryption key using values from the config file
// and the user's passphrase.
func DecryptEncryptionKey(password string) []byte {

	// Read the crypto config
	c := new(config.Config)
	c.Initialise()
	c.Load()

	salt := decodeHexString(c.Salt)
	passphraseHash := decodeHexString(c.PasswordHash)
	encryptedKey := decodeHexString(c.EncryptedKey)
	encryptedKeyIv := decodeHexString(c.EncryptedKeyIv)

	derivedKey := pbkdf2.Key([]byte(password), salt, 65536, 64, sha256.New)
	// Make sure the first 32 bytes of the derived key match the bytes stored
	// when we first generated the key; if they don't, the user gave us
	// the wrong passphrase.
	if !bytes.Equal(derivedKey[:32], passphraseHash) {
		helper.OutputAndExit(fmt.Sprintf("Incorrect password"))
	}

	// Use the last 32 bytes of the derived key to decrypt the actual
	// encryption key.
	keyEncryptKey := derivedKey[32:]
	return decryptBytes(keyEncryptKey, encryptedKeyIv, encryptedKey)
}

// Utility function to decode hex-encoded bytes; treats any encoding errors
// as fatal errors (we assume that checkConfigValidity has already made
// sure the strings in the config file are reasonable.)
func decodeHexString(s string) []byte {

	r, err := hex.DecodeString(s)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Unable to decode hex string: %v", err))
	}
	return r
}

// Decrypt the given cyphertext using the given encryption key and
// initialization vector 'iv'.
func decryptBytes(key []byte, iv []byte, ciphertext []byte) []byte {

	r, _ := ioutil.ReadAll(MakeDecryptionReader(key, iv, bytes.NewReader(ciphertext)))
	return r
}
