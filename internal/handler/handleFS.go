package handler

import (
	"net/http"
	"os"
)

// Wrapper for default FileSystem
type justFilesFilesystem struct {
	fs http.FileSystem
}

// Wrapper for default File
type neuteredReaddirFile struct {
	http.File
}

// Modifies the behavior of FileSystem.Open to return the neutered version of File
func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	// Check if the requested path is a directory
	// and explicitly return an error to trigger a 404
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, os.ErrNotExist
	}

	return neuteredReaddirFile{f}, nil
}

// Overrides the Readdir method of File to always return nil
func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

// Handles requests for static files, without allowing access to the
// directory viewer and returning 404 if an exact file is not found
func HandleFS(staticFS *http.FileSystem) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nfs := justFilesFilesystem{*staticFS}
			fs := http.FileServer(nfs)
			fs.ServeHTTP(w, r)
		},
	)
}
