package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	helper "filesender/utils"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"

	config "filesender/config"
	crypto "filesender/crypto"
)

// ##### Variables ###########################################################

var cmdGenerate = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"g"},
	Short:   "Generates the crypto data",
	Long:    `Generates the crypto data and stores within the config file`,
	Run:     generate,
}

// ##### Functions ###########################################################

// Add the command to the cobra setup
func init() {

	cmdRoot.AddCommand(cmdGenerate)
}

// generate performs the generation of the crypto data
func generate(cmd *cobra.Command, args []string) {

	generateKey()
}

// generateKey creates a new encryption key and encrypts it using the user-provided password
func generateKey() {

	password := getPassword()

	// Derive a 64-byte hash from the passphrase using PBKDF2 with 65536 rounds of SHA256.
	salt := getRandomBytes(32)
	hash := pbkdf2.Key([]byte(password), salt, 65536, 64, sha256.New)
	if len(hash) != 64 {
		helper.OutputAndExit(fmt.Sprintf("Incorrect key size returned by pbkdf2: %d", len(hash)))
	}

	// We'll store the first 32 bytes of the hash to use to confirm the
	// correct passphrase is given on subsequent runs.
	passHash := hash[:32]
	// And we'll use the remaining 32 bytes as a key to encrypt the actual
	// encryption key. (These bytes are *not* stored).
	keyEncryptKey := hash[32:]

	// Generate a random encryption key and encrypt it using the key
	// derived from the passphrase.
	key := getRandomBytes(32)
	iv := getRandomBytes(16)
	encryptedKey := encryptBytes(keyEncryptKey, iv, key)

	// Write the crypto config
	c := new(config.Config)
	c.Initialise()
	c.Salt = hex.EncodeToString(salt)
	c.PasswordHash = hex.EncodeToString(passHash)
	c.EncryptedKey = hex.EncodeToString(encryptedKey)
	c.EncryptedKeyIv = hex.EncodeToString(iv)
	c.Save()

	fmt.Printf("Crypto data generated and written to the configuration\n")
}

// getPassword prompts the user for a password and password confirmation
func getPassword() string {

	fmt.Printf("Enter password: ")
	password1, err := gopass.GetPasswd()
	if err != nil {
		helper.OutputAndExit("Error reading password")
	}
	if len(password1) == 0 {
		helper.OutputAndExit("Password not entered")
	}

	fmt.Printf("Repeat password: ")
	password2, err := gopass.GetPasswd()
	if err != nil {
		helper.OutputAndExit("Error reading password")
	}
	if string(password1) != string(password2) {
		helper.OutputAndExit("Passwords do not match")
	}

	return string(password1)
}

// Encrypt the given plaintext using the given encryption key 'key' and
// initialization vector 'iv'. The initialization vector should be 16 bytes
// (the AES block-size), and should be randomly generated and unique for
// each file that's encrypted.
func encryptBytes(key []byte, iv []byte, plaintext []byte) []byte {

	r, _ := ioutil.ReadAll(crypto.MakeEncrypterReader(key, iv, bytes.NewReader(plaintext)))
	return r
}

// Return the given number of bytes of random values, using a
// cryptographically-strong random number source.
func getRandomBytes(n int) []byte {

	bytes := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Unable to get random bytes %v", err))
	}

	return bytes
}
