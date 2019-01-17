package cmd

import (
	"fmt"
	"io"
	"os"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// ##### Structs ###########################################################

// CountingReader keeps tracks of how many bytes are actually read via Read() calls.
type CountingReader struct {
	R         io.Reader
	bytesRead int64
}

// Read performs a read operation, and keeps tally the number of bytes read
func (cr *CountingReader) Read(dst []byte) (int, error) {

	read, err := cr.R.Read(dst)
	cr.bytesRead += int64(read)
	return read, err
}

// ##### Methods ###########################################################

// getProgressBar creates and initialises a progress bar
func getProgressBar(nBytes int64) *pb.ProgressBar {

	progressBar := pb.New64(nBytes).SetUnits(pb.U_BYTES)
	progressBar.ShowBar = true
	progressBar.Output = os.Stderr
	progressBar.Start()
	return progressBar
}

//  byteCountIEC converts a size in bytes to a human-readable string in IEC (binary) format.
func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
