package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnTarGz(srcFile string, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	fr, err := os.Open(srcFile)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = fr.Close()
	}()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = gr.Close()
	}()

	tr := tar.NewReader(gr)

	targetFile := ""
	destClean := filepath.Clean(destDir)
	destWithSep := destClean + string(os.PathSeparator)
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		// sanitize path: no absolute paths or parent traversal
		cleaned := filepath.Clean(header.Name)
		if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
			return "", errors.New("tar: unsafe path detected")
		}

		// determine top-level target dir once
		if targetFile == "" {
			parts := strings.Split(cleaned, string(os.PathSeparator))
			if len(parts) > 0 {
				targetFile = filepath.Join(destClean, parts[0])
			}
		}

		filename := filepath.Join(destClean, cleaned)
		nameClean := filepath.Clean(filename)
		if !(nameClean == destClean || strings.HasPrefix(nameClean, destWithSep)) {
			return "", errors.New("tar: path escapes destination")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(nameClean, os.FileMode(header.Mode)); err != nil {
				return "", err
			}

		case tar.TypeReg:
			file, err := os.OpenFile(nameClean, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(file, tr); err != nil {
				_ = file.Close()
				return "", err
			}
			_ = file.Close()

		case tar.TypeSymlink, tar.TypeLink:
			// do not allow links in update archives
			return "", errors.New("tar: links are not allowed in archive")
		}
	}

	return targetFile, nil
}
