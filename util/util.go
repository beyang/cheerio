package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type CompressionType string

const (
	Zip CompressionType = "zip"
	Tar                 = "tar"
)

func RemoteDecompress(uri string, pattern string, compressType CompressionType) ([]byte, error) {
	switch compressType {
	case Zip:
		return remoteUnzip(uri, pattern)
	case Tar:
		return remoteUntar(uri, pattern)
	}
	return nil, fmt.Errorf("Unrecognized compression type: %s", compressType)
}

func remoteUntar(uri string, pattern string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gunzipped, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(gunzipped)
	var data []byte
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		matches, err := filepath.Match(pattern, hdr.Name)
		if err != nil {
			return nil, err
		}
		if matches {
			buf := bytes.NewBuffer(make([]byte, 0, hdr.Size))
			io.Copy(buf, tr)
			data = append(data, buf.Bytes()...)
		}
	}

	return data, nil
}

func remoteUnzip(uri string, pattern string) ([]byte, error) {
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
	for _, file := range zr.File {
		matches, err := filepath.Match(pattern, file.Name)
		if err != nil {
			return nil, err
		}
		if matches {
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
		}
	}

	return data, nil
}
