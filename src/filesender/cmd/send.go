package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	crypto "filesender/crypto"
	drive "filesender/drive"
	helper "filesender/utils"

	uuid "github.com/gofrs/uuid"
	"github.com/howeyc/gopass"
	"github.com/schollz/mnemonicode"
	"github.com/spf13/cobra"
	util "github.com/woanware/goutil"
)

// ##### Variables ###########################################################

var cmdSend = &cobra.Command{
	Use:     "send [file path]",
	Aliases: []string{"s"},
	Short:   "Sends a file",
	Long:    `Sends a file to google drive`,
	Run:     send,
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return errors.New("Requires the path to the sending file")
		}

		if util.DoesFileExist(args[0]) == false {
			return fmt.Errorf("Send file does not exist: %s", args[0])
		}

		return nil
	},
}

// ##### Functions ###########################################################

// Add the command to the cobra setup
func init() {

	cmdSend.Flags().BoolP("encrypt", "e", false, "Encrypt the file using the pre-defined crypto data")
	cmdRoot.AddCommand(cmdSend)
}

// send performs the sending of the file
func send(cmd *cobra.Command, args []string) {

	sendFile := args[0]

	_, gdrive := drive.InitialiseGoogleDrive()

	encrypt, err := cmd.Flags().GetBool("encrypt")
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error reading command line parameters: %v", err))
	}

	var password []byte
	var iv []byte
	var ivHex string
	if encrypt == true {
		fmt.Printf("Enter password: ")
		password, err = gopass.GetPasswd()
		fmt.Println("")
		if err != nil {
			helper.OutputAndExit("Error reading password")
		}
		if len(password) == 0 {
			helper.OutputAndExit("Password not supplied")
		}

		// Compute a unique IV for the file.
		iv = getRandomBytes(aes.BlockSize)
		ivHex = hex.EncodeToString(iv)
	}

	fileReader, length, err := getFileContentsReaderForUpload(string(password), sendFile, encrypt, iv)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error reading file contents: %v", err))
	}
	defer fileReader.Close()

	fmt.Printf("Sending %s file: %s\n", byteCountIEC(length), sendFile)

	// Generate a GUID/UUID, which is used as a file name in google drive,
	// this is designed to overcome file name clashes
	guid, err := uuid.NewV4()
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Failed to generate UUID: %v", err))
	}

	mnemonicode := generateMnemonic()

	// Define the AppProperties meta data
	ap := make(map[string]string, 0)
	ap["mnemonicode"] = mnemonicode
	ap["file_name"] = sendFile
	ap["iv"] = ivHex

	// Also tee reads to the progress bar as they are done so that it
	// stays in sync with how much data has been transmitted.
	cr := &CountingReader{R: fileReader}
	progressBar := getProgressBar(length)
	reader := io.TeeReader(cr, progressBar)

	_, err = gdrive.PutFile("filesender/"+guid.String(), ap, reader)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Failed to upload file: %v", err))
	}

	progressBar.Finish()

	fmt.Printf("\nCode is: %s\n", mnemonicode)
	fmt.Printf("On the other computer run: filesender r %s\n", mnemonicode)
}

//
func generateMnemonic() string {

	result := []string{}
	bs := make([]byte, 4)
	rand.Read(bs)

	return strings.Join(mnemonicode.EncodeWordList(result, bs), "-")
}

// Returns an io.ReadCloser for given file, such that the bytes read are
// ready for upload: specifically, if encryption is enabled, the contents
// are encrypted with the given key and the initialization vector is
// prepended to the returned bytes. Otherwise, the contents of the file are
// returned directly.
func getFileContentsReaderForUpload(password string, path string, encrypt bool, iv []byte) (io.ReadCloser, int64, error) {

	f, err := os.Open(path)
	if err != nil {
		return f, 0, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}
	fileSize := stat.Size()

	if encrypt {
		key := crypto.DecryptEncryptionKey(password)
		r := crypto.MakeEncrypterReader(key, iv, f)

		// Prepend the initialization vector to the returned bytes.
		r = io.MultiReader(bytes.NewReader(iv[:aes.BlockSize]), r)

		readCloser := struct {
			io.Reader
			io.Closer
		}{r, f}
		return readCloser, fileSize + aes.BlockSize, nil
	}
	return f, fileSize, nil
}
