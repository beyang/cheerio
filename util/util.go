package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
)

type CompressionType string

const (
	Zip CompressionType = "zip"
	Tar                 = "tar"
)

func RemoteDecompress(uri string, pattern *regexp.Regexp, compressType CompressionType) ([]byte, error) {
	switch compressType {
	case Zip:
		return remoteUnzip(uri, pattern)
	case Tar:
		return remoteUntar(uri, pattern)
	}
	return nil, fmt.Errorf("Unrecognized compression type: %s", compressType)
}

func remoteUntar(uri string, pattern *regexp.Regexp) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decompressed io.Reader
	if filepath.Ext(uri) == ".bz2" {
		decompressed = bzip2.NewReader(resp.Body)
	} else {
		decompressed, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	tr := tar.NewReader(decompressed)
	var data []byte
	matched := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if hdr == nil {
			return nil, fmt.Errorf("Error untarring %s: nil header (may be malformed)", uri)
		}

		if pattern.MatchString(hdr.Name) {
			buf := bytes.NewBuffer(make([]byte, 0, hdr.Size))
			io.Copy(buf, tr)
			data = append(data, buf.Bytes()...)
			matched = true
		}
	}
	if !matched {
		return nil, fmt.Errorf("No file matched pattern %+v", pattern)
	}

	return data, nil
}

func remoteUnzip(uri string, pattern *regexp.Regexp) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	zipdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(zipdata), resp.ContentLength)
	if err != nil {
		return nil, err
	}

	var data []byte
	matched := false
	for _, file := range zr.File {
		if file == nil {
			return nil, fmt.Errorf("Error unzipping %s: nil file (may be malformed)", uri)
		}

		if pattern.MatchString(file.Name) {
			fr, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer fr.Close()
			filedata, err := ioutil.ReadAll(fr)
			if err != nil {
				return nil, err
			}
			data = append(data, filedata...)
			matched = true
		}
	}
	if !matched {
		return nil, fmt.Errorf("No file matched pattern %+v", pattern)
	}

	return data, nil
}
