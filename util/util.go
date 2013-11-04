package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	// Since zip cannot decompress from stdin, we need to save a temporary file
	f, err := ioutil.TempFile("", "cheerio_unzip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if err := f.Close(); err != nil {
		return nil, err
	}

	wget := exec.Command("wget", uri, "-O", f.Name())
	unzip := exec.Command("unzip", "-cq", f.Name(), pattern)

	err = wget.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running wget: %s", err)
	}

	unzipOutput, err := unzip.Output()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 11") {
			return nil, nil // TODO: should return the error to let client handle; need to make tar consistent with this, too
		} else {
			return nil, fmt.Errorf("Error running unzip on file %s: %s", f.Name(), err)
		}
	}

	return unzipOutput, nil
}
