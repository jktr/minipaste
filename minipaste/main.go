// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

var (
	server string
	rm     bool
	file   string
)

func init() {
	flag.StringVar(&server, "server", "http://[::1]:8080", "addr of server")
	flag.BoolVar(&rm, "delete", false, "delete paste instead of uploading")
	flag.Parse()
	file = flag.Arg(0)
	if file == "" {
		file = "-"
	}
}

type Source struct {
	Reader        io.Reader
	ContentLength int64
	ContentType   string
}

func main() {

	var req *http.Request
	var err error

	if rm {
		u, err := url.Parse(server)
		if err != nil {
			log.Fatal(err)
		}
		u.Path = path.Join(u.Path, "paste")
		req, err = http.NewRequest(http.MethodDelete, u.String(), nil)
	} else {
		req, err = uploadRequest(file)
	}

	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{
		Timeout: time.Minute,
		CheckRedirect: func(requ *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
}

func uploadRequest(file string) (*http.Request, error) {
	var source *Source
	var err error

	if file == "-" {
		source, err = stdinSource()
	} else {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		source, err = fileSource(f)
		// leaks file
	}

	if err != nil {
		return nil, err
	}

	if source.ContentLength == 0 {
		return nil, errors.New(fmt.Sprintf("file %s has no content", file))
	}

	url := fmt.Sprintf("%s/%s", server, path.Base(file))
	r, err := http.NewRequest(http.MethodPut, url, source.Reader)
	if err != nil {
		return nil, err
	}

	r.Header.Set("content-type", source.ContentType)
	r.ContentLength = source.ContentLength
	return r, nil
}

func stdinSource() (*Source, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)

	return &Source{
		Reader:        r,
		ContentLength: int64(len(b)),
		ContentType:   http.DetectContentType(b),
	}, nil
}

func fileSource(f *os.File) (*Source, error) {
	r := bufio.NewReader(f)
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	buf, _ := r.Peek(512)
	return &Source{
		Reader:        r,
		ContentLength: stat.Size(),
		ContentType:   http.DetectContentType(buf),
	}, nil
}
