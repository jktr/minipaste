// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/dimfeld/httptreemux/v5"
)

func (s *State) AddPUTStyleUploading(r *httptreemux.TreeMux) {
	f := func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		log.Printf(`upload: ip="%s" ua="%s"`+"\n",
			r.RemoteAddr, r.Header.Get("user-agent"))

		contentType, err := verifyContentType(r.Header.Get("content-type"))
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}

		length, err := verifyLength(r.Header.Get("content-length"), s.limit)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}

		buf, err := readContent(r.Body, length)
		if err != nil {
			log.Println(err)
			http.Error(w, "upload failed", http.StatusInternalServerError)
			return
		}

		paste := Paste{
			Name:          p["name"],
			ContentType:   contentType,
			ContentLength: len(buf),
			Content:       buf,
			Uploaded:      time.Now().UTC(),
		}

		s.SetPaste("put", &paste)
		redirectToPaste(paste, w, r)
	}
	r.PUT("/:name", f)
}

func (s *State) AddNullPointerStyleUploading(r *httptreemux.TreeMux) {
	f := func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		log.Printf(`upload: ip="%s" ua="%s"`+"\n",
			r.RemoteAddr, r.Header.Get("user-agent"))

		length, err := verifyLength(r.Header.Get("content-length"), s.limit)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}

		mr, err := r.MultipartReader()
		if err != nil {
			log.Println(err)
			http.Error(w, "submitted form was corrupt", http.StatusBadRequest)
			return
		}

		for p, err := mr.NextPart(); err != io.EOF; p, err = mr.NextPart() {
			if err != nil {
				log.Println(err)
				http.Error(w, "submitted form was corrupt", http.StatusBadRequest)
				return
			}

			if p.FormName() != "file" {
				log.Println(err)
				http.Error(w, "submitted form contains unsupported fields", http.StatusBadRequest)
				return
			}

			contentType, err := verifyContentType(p.Header.Get("content-type"))
			if err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
				return
			}

			buf, err := readContent(p, length)
			if err != nil {
				log.Println(err)
				http.Error(w, "upload failed", http.StatusInternalServerError)
				return
			}

			paste := Paste{
				Name:          p.FileName(),
				ContentType:   contentType,
				ContentLength: len(buf),
				Content:       buf,
				Uploaded:      time.Now().UTC(),
			}

			s.SetPaste("0x0", &paste)
			redirectToPaste(paste, w, r)
			return
		}
	}

	r.POST("/", f)
}

func redirectToPaste(paste Paste, w http.ResponseWriter, r *http.Request) {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme += "s"
		}
	}

	host := r.URL.Host
	if host == "" {
		host = r.Host
	}

	redirect := fmt.Sprintf("%s://%s/paste\n", scheme, host)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
	fmt.Fprintf(w, redirect)
}

func readContent(r io.ReadCloser, limit int64) ([]byte, error) {
	lbody := http.MaxBytesReader(nil, r, limit)
	defer lbody.Close()

	buf := bytes.Buffer{}
	_, err := io.Copy(&buf, lbody)
	return buf.Bytes(), err
}

func verifyContentType(contentTypeStr string) (string, error) {

	if contentTypeStr == "" {
		contentTypeStr = "application/octet-stream"
	}

	contentType, _, err := mime.ParseMediaType(contentTypeStr)
	if err != nil {
		return "", errors.New("garbage content type")
	}

	return contentType, nil
}

func verifyLength(lenStr string, uploadLimitBytes int64) (int64, error) {

	if lenStr == "" {
		return 0, errors.New("missing content length")
	}

	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return 0, errors.New("garbage content length")
	}

	if int64(length) > uploadLimitBytes {
		return 0, errors.New("upload exceeds size limit")
	}

	return int64(length), nil
}
