package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type (
	Controller struct {
		FileSystem
	}
)

func (c *Controller) callback(ctx context.Context, url *url.URL, meta *FileMetadata) error {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(meta)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), body)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Controller) upload(w http.ResponseWriter, r *http.Request) {
	storeMeta := &FileMetadata{}
	if wantMD5 := r.URL.Query().Get("md5"); wantMD5 != "" {
		storeMeta.MD5 = &MD5Hash{}
		err := storeMeta.MD5.FromString(wantMD5)
		if err != nil {
			http.Error(w, "bad md5", http.StatusBadRequest)
			return
		}
	}
	var cbURL *url.URL
	if cb := r.URL.Query().Get("cb"); cb != "" {
		var err error
		cbURL, err = url.Parse(cb)
		if err != nil {
			http.Error(w, "bad callback url", http.StatusBadRequest)
			return
		}
	}
	meta, err := c.FileSystem.Store(r.Context(), r.Body, storeMeta)
	switch {
	case errors.Is(err, fs.ErrExist):
		http.Error(w, "", http.StatusConflict)
		return
	case err == ErrMD5Mismatch:
		http.Error(w, "md5 mismatch", http.StatusBadRequest)
		return
	case err != nil:
		log.Println("store() error:", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if cbURL != nil {
		err := c.callback(r.Context(), cbURL, meta)
		if err != nil {
			log.Println("failed to post callback for", cbURL, meta.MD5, err)
		}
	}

	w.Header().Add("content-type", "application/json")
	err = json.NewEncoder(w).Encode(&meta)
	if err != nil {
		log.Println("failed to encode response:", err)
	}
}
func (c *Controller) download(w http.ResponseWriter, r *http.Request) {
	hashStr, ok := mux.Vars(r)["hash"]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	hash := &MD5Hash{}
	if err := hash.FromString(hashStr); err != nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	f, err := c.FileSystem.OpenHash(hash)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		http.Error(w, "", http.StatusNotFound)
		return
	case err != nil:
		log.Println("open() error:", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if r.URL.Query().Get("validate") != "" {
		sum := md5.New()
		_, err := io.Copy(sum, f)
		if err != nil {
			log.Println("failed to calculate md5 when validating")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		if !hash.Equals((*MD5Hash)(sum.Sum(nil))) {
			log.Println("file", hash, "corrupt")
			http.Error(w, "file corrupt", http.StatusServiceUnavailable)
			return
		}

	}

	stat, err := f.Stat()
	if err != nil {
		log.Println("stat() failed:", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "", stat.ModTime(), f)
}
func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	hashStr, ok := mux.Vars(r)["hash"]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	hash := &MD5Hash{}
	err := hash.FromString(hashStr)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	err = c.FileSystem.Delete(hash)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		http.Error(w, "", http.StatusNotFound)
		return
	case err != nil:
		log.Println("delete() error:", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) Router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/upload", c.upload).Methods(http.MethodPost)
	/*r.PathPrefix("/download").Handler(
		http.FileServer(c.FileSystem),
	).Methods(http.MethodGet)*/
	r.HandleFunc("/download/{hash}", c.download).Methods(http.MethodGet)
	r.HandleFunc("/delete/{hash}", c.delete).Methods(http.MethodDelete)

	return r
}
