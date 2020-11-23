// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

func (s *State) AddPasteViewing(r *httptreemux.TreeMux) {
	f := func(w http.ResponseWriter, r *http.Request, p map[string]string) {

		log.Printf(`download: ip="%s" path="%s" ua="%s"`+"\n",
			r.RemoteAddr, r.URL.Path, r.Header.Get("user-agent"))

		paste := s.GetPaste()
		if paste == nil {
			// we assume that a previous paste expired
			http.Error(w, "410 Gone", http.StatusGone)
			return
		}

		h := w.Header()

		// security headers
		h.Set("content-security-policy", "default-src 'none'")
		h.Set("referrer-policy", "no-referrer")
		h.Set("x-frame-options", "DENY")
		h.Set("x-content-type-options", "nosniff")
		h.Set("x-xss-protection", "1; mode=lock")

		h.Set("content-type", paste.ContentType)

		if paste.Name != "-" {
			h.Set("content-disposition", fmt.Sprintf(`filename="%s"`, paste.Name))
		}

		http.ServeContent(w, r, "", paste.Uploaded, bytes.NewReader(paste.Content))
	}
	r.GET("/paste", f)
}

func (s *State) AddPasteDeletion(r *httptreemux.TreeMux) {
	f := func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		log.Println("deleted paste: cause=request")
		s.ExpirePaste()
	}
	r.DELETE("/paste", f)
}
