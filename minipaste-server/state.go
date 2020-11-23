// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Paste struct {
	Name          string
	ContentType   string
	ContentLength int
	Content       []byte
	Uploaded      time.Time
}

type State struct {
	l          sync.Mutex
	retention  time.Duration
	retentionT *time.Timer
	paste      *Paste
	limit      int64
}

func NewState(retention time.Duration, limit int64) *State {
	return &State{
		retention: retention,
		limit:     limit,
	}
}

func (s *State) ExpirePaste() {
	s.l.Lock()
	s.paste = nil
	s.retentionT = nil
	s.l.Unlock()
}

func (s *State) GetPaste() *Paste {
	return s.paste
}

func (s *State) SetPaste(method string, paste *Paste) {

	if paste.ContentType == "" || paste.ContentType == "application/octet-stream" {
		paste.ContentType = http.DetectContentType(paste.Content)
		// TODO content type filter
	}

	if paste.Name == "-" {
		// XXX RFC3339 contains ":", which is stripped by browsers
		paste.Name = strconv.FormatInt(time.Now().UTC().Unix(), 10)
	}

	s.l.Lock()
	defer s.l.Unlock()

	if s.retentionT != nil {
		s.retentionT.Stop()
	}

	log.Printf(`upload: method="%s" name="%s" type="%s" length=%d`+"\n",
		method, paste.Name, paste.ContentType, paste.ContentLength)
	s.paste = paste
	if s.retention > 0 {
		s.retentionT = time.AfterFunc(s.retention, func() {
			log.Println("deleted paste: cause=retention")
			s.ExpirePaste()
		})
	}
}
