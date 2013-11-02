package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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
	curl := exec.Command("curl", uri)
	tar := exec.Command("tar", "-xvO", "--include", pattern)

	curlOut, err := curl.StdoutPipe()
	if err != nil {
		return nil, err
	}
	tarIn, err := tar.StdinPipe()
	if err != nil {
		return nil, err
	}
	tarOut, err := tar.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		io.Copy(tarIn, curlOut)
		tarIn.Close()
	}()

	curl.Start()
	tar.Start()

	tarOutput, err := ioutil.ReadAll(tarOut)
	if err != nil {
		return nil, err
	}

	curl.Wait()
	tar.Wait()

	return tarOutput, nil
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
