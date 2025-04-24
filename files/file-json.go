package files

import (
	"encoding/json" // Import the correct package for JSON marshaling
	"time"          // Import the time package
)

type FileInfoJSON struct {
	ModTime  string `json:"modified"` // custom serialization
	FileInfo        // all fields from FileInfo
}

// custom JSON marshal function for FileInfo
func (f *FileInfo) MarshalJSON() ([]byte, error) {

	fiJson := FileInfoJSON{
		FileInfo: *f, // copy all from  FileInfo
	}

	// Check if ModTime is the Unix epoch time (zero time)
	if f.ModTime.UnixNano() == 0 {
		fiJson.ModTime = "" // Set ModTime to an empty string
	} else {
		// Otherwise, format the ModTime as usual (ISO 8601 format)
		fiJson.ModTime = f.ModTime.Format(time.RFC3339)
	}

	return json.Marshal(fiJson)
}
