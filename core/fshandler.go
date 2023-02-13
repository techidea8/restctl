package core

import (
	"embed"
	"io/fs"
	"net/http"

	log "github.com/techidea8/restctl/pkg/log"
)

type FSHandler struct {
	Fs     embed.FS
	Root   string
	Patern string
}

func (h FSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subfs, err := fs.Sub(h.Fs, h.Root)
	if err != nil {
		log.Error(err.Error())
	}
	http.FileServer(http.FS(subfs)).ServeHTTP(w, r)
}
