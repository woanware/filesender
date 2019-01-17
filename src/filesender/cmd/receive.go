package cmd

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	crypto "filesender/crypto"
	drive "filesender/drive"
	helper "filesender/utils"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/woanware/gdriver"
	util "github.com/woanware/goutil"
)

// ##### Variables ###########################################################

var cmdReceive = &cobra.Command{
	Use:     "receive [mnemonicode]",
	Aliases: []string{"r"},
	Short:   "Receives a file",
	Long:    `Receives a file from google drive`,
	Run:     receive,
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return errors.New("Requires the mnemonicode")
		}

		return nil
	},
}

// ##### Functions ###########################################################

// Add the command to the cobra setup
func init() {

	cmdReceive.Flags().BoolP("leave", "l", false, "Leave the file on google drive e.g. no delete")
	cmdRoot.AddCommand(cmdReceive)
}

// send performs the sending of the file
func receive(cmd *cobra.Command, args []string) {

	leave, err := cmd.Flags().GetBool("leave")
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error reading command line parameters: %v", err))
	}

	mnemonicode := args[0]

	_, gdrive := drive.InitialiseGoogleDrive()

	foundFile := false

	err = gdrive.ListDirectory("/filesender", func(fi *gdriver.FileInfo) error {

		if fi.DriveFile().AppProperties["mnemonicode"] != mnemonicode {
			return nil
		}

		fileName := fi.DriveFile().AppProperties["file_name"]
		if len(fileName) == 0 {
			return fmt.Errorf("File does not contain original file name meta data")
		}

		encrypted, ivp, err := checkIfEncrypted(fi)

		var key []byte
		if encrypted == true {
			key = getDecryptionKey()
		}

		// Get the file contents from google drive
		_, reader, err := gdrive.GetFile(fi.Path())
		if err != nil {
			return err
		}

		var r io.Reader
		r = reader

		if encrypted == true {
			r = validateIv(r, key, ivp)
		}

		foundFile = true

		checkLocalFile(fileName)
		writeLocalFile(fileName, fi.Size(), r)

		if leave == false {
			// Now delete the file from google drive
			err = gdrive.Delete(fi.Path())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		helper.OutputAndExit(err.Error())
	}

	if foundFile == false {
		fmt.Printf("Unable to locate file\n")
	}
}

//
func validateIv(r io.Reader, key []byte, ivp []byte) io.Reader {

	// Read the initialization vector from the start of the file.
	iv := make([]byte, 16)
	n, err := r.Read(iv)
	if err != nil && err != io.EOF {
		helper.OutputAndExit(fmt.Sprintf("Error reading file contents: %v", err))
	}

	if n < aes.BlockSize {
		helper.OutputAndExit(fmt.Sprintf("Contents too short to hold IV: %d bytes", n))
	}

	// Double check that the IV matches the one in the Drive metadata.
	if bytes.Compare(iv, ivp) != 0 {
		helper.OutputAndExit(fmt.Sprintf("File header IV [%s] doesn't match meta data IV [%s]", hex.EncodeToString(iv), hex.EncodeToString(ivp)))
	}

	return crypto.MakeDecryptionReader(key, iv, r)
}

//
func checkIfEncrypted(fi *gdriver.FileInfo) (bool, []byte, error) {

	encrypted := false
	ivHex := fi.DriveFile().AppProperties["iv"]
	if len(ivHex) > 0 {
		encrypted = true
	} else {
		return false, []byte{}, nil
	}

	ivp, err := hex.DecodeString(ivHex)
	if err != nil {
		return encrypted, ivp, err
	}

	if len(ivp) != aes.BlockSize {
		return encrypted, ivp, fmt.Errorf("unexpected length of IV %d", len(ivp))
	}

	return encrypted, ivp, nil
}

//
func getDecryptionKey() []byte {

	fmt.Printf("Enter password: ")
	password, err := gopass.GetPasswd()
	if err != nil {
		helper.OutputAndExit("Error reading password")
	}
	if len(password) == 0 {
		helper.OutputAndExit("Password not supplied")
	}

	return crypto.DecryptEncryptionKey(string(password))
}

// checkLocalFile determines if the file exists in the CWD and
// prompts the user to check if they want to overwrite the file
func checkLocalFile(fileName string) {

	// Check if file already exists locally
	cwd, err := os.Getwd()
	if err != nil {
		helper.OutputAndExit("Unable to determine CWD")
	}

	if util.DoesFileExist(path.Join(cwd, fileName)) == true {
		//Prompt user for overwrite
		fmt.Printf("File exists locally. Do you want to overwrite?:")
		ret, err := util.GetYesNoPrompt(false)
		if err != nil {
			helper.OutputAndExit(fmt.Sprintf("Unable to read user input: %v", err))
		}

		if ret == false {
			helper.OutputAndExit("Receive cancelled")
		}
	}
}

// writeLocalFile writes the google drive file to the
// local disk, and updates progress using a progress bar
func writeLocalFile(fileName string, fileSize int64, r io.Reader) {

	fmt.Println("")

	progressBar := getProgressBar(fileSize)

	// Create the local file
	writer, err := os.Create(fileName)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error creating file: %v", err))
	}
	defer writer.Close()

	// Tee writes to the progress bar, which provides the Writer interface
	// and updates itself according to the number of bytes that it sees.
	mW := io.MultiWriter(writer, progressBar)

	// And here's where the magic happens
	cr := &CountingReader{R: r}
	_, err = io.Copy(mW, cr)
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error copying file contents: %v", err))
	}

	progressBar.Finish()

	fmt.Printf("Received %s file: %s\n", fileName, byteCountIEC(cr.bytesRead))
}
