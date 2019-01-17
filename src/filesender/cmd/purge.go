package cmd

import (
	drive "filesender/drive"
	helper "filesender/utils"

	"github.com/spf13/cobra"
	"github.com/woanware/gdriver"
)

// ##### Variables ###########################################################

var cmdPurge = &cobra.Command{
	Use:     "purge",
	Aliases: []string{"p"},
	Short:   "Purges all files",
	Long:    `Purges all files from google drive e.g. related to filesender!`,
	Run:     purge,
}

// ##### Functions ###########################################################

// Add the command to the cobra setup
func init() {

	cmdRoot.AddCommand(cmdPurge)
}

// send performs the sending of the file
func purge(cmd *cobra.Command, args []string) {

	_, gdrive := drive.InitialiseGoogleDrive()

	err := gdrive.ListDirectory("/filesender", func(fi *gdriver.FileInfo) error {

		// If there is no mnemonicode app property or it is zero length, then leave the file
		if len(fi.DriveFile().AppProperties["mnemonicode"]) == 0 {
			return nil
		}

		// Now delete the file from google drive
		err := gdrive.Delete(fi.Path())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		helper.OutputAndExit(err.Error())
	}
}
