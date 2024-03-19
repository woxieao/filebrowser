package http

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

var anonymousUser = func(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		path := r.URL.Path
		userId := uint(1)

		user, err := d.store.Users.Get(d.server.Root, userId)

		if err != nil {
			return errToStatus(err), err
		}
		d.user = user

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:         d.user.Fs,
			Path:       path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
		})
		if err != nil {
			return errToStatus(err), err
		}
		// share base path
		basePath := path

		// file relative path
		filePath := ""

		if !file.IsDir {
			// set fs root to the shared file/folder
			d.user.Fs = afero.NewBasePathFs(d.user.Fs, basePath)

			file, err = files.NewFileInfo(files.FileOptions{
				Fs:     d.user.Fs,
				Path:   filePath,
				Modify: d.user.Perm.Modify,
				Expand: true,
			})
			if err != nil {
				return errToStatus(err), err
			}
			d.raw = file
			return fn(w, r, d)
		}

		return 0, nil

	}
}

var fileGetHandler = anonymousUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file := d.raw.(*files.FileInfo)
	return rawFileInlineHandler(w, r, file)
})
