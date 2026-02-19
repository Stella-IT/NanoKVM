package utils

import (
    "archive/zip"
    "errors"
    "io"
    "os"
    "path/filepath"
    "strings"
)

func Unzip(filename string, dest string) error {
    r, err := zip.OpenReader(filename)
    if err != nil {
        return err
    }
    defer func() {
        _ = r.Close()
    }()

    destClean := filepath.Clean(dest)
    destWithSep := destClean + string(os.PathSeparator)
    for _, f := range r.File {
        cleaned := filepath.Clean(f.Name)
        if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
            return errors.New("zip: unsafe path detected")
        }
        dstPath := filepath.Join(destClean, cleaned)
        pathClean := filepath.Clean(dstPath)
        if !(pathClean == destClean || strings.HasPrefix(pathClean, destWithSep)) {
            return errors.New("zip: path escapes destination")
        }
        if f.FileInfo().IsDir() {
            err = os.MkdirAll(pathClean, 0o755)
            if err != nil {
                return err
            }
        } else {
            err = unzipFile(pathClean, f)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func unzipFile(dstPath string, f *zip.File) error {
	err := os.MkdirAll(filepath.Dir(dstPath), 0o755)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	archivedFile, err := f.Open()
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, archivedFile); err != nil {
		return err
	}
	if err = os.Chmod(dstPath, f.Mode()); err != nil {
		return err
	}
	return nil
}
