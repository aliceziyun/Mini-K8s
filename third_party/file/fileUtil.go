package file

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
)

func Bytes2File(raw []byte, name string, dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	file, err := os.OpenFile(path.Join(dirPath, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = io.Copy(file, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	return nil
}

func Unzip(zipPathName string, destDir string) error {
	zipReader, err := zip.OpenReader(zipPathName)
	if err != nil {
		return err
	}

	defer func(zipReader *zip.ReadCloser) {
		err := zipReader.Close()
		if err != nil {

		}
	}(zipReader)

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
		} else {
			var outFile *os.File
			var inFile io.ReadCloser

			err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
			if err != nil {
				return err
			}

			inFile, err = f.Open()
			if err != nil {
				return err
			}

			outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				err := outFile.Close()
				if err != nil {
					return err
				}
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				err := inFile.Close()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
