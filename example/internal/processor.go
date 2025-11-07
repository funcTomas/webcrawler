package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"webcrawler/module"
)

func genItemProcessors(dirPath string) []module.ProcessItem {
	savePic := func(item module.Item) (result module.Item, err error) {
		if item == nil {
			return nil, errors.New("invalid item!")
		}
		var absDirPath string
		if absDirPath, err = checkDirPath(dirPath); err != nil {
			return
		}
		v := item["reader"]
		reader, ok := v.(io.Reader)
		if !ok {
			return nil, fmt.Errorf("incorrect reader type: %T", v)
		}
		readcloser, ok := reader.(io.ReadCloser)
		if ok {
			defer readcloser.Close()
		}
		v = item["name"]
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("incorrect name type: %T", v)
		}
		fileName := name
		fPath := filepath.Join(absDirPath, fileName)
		file, err := os.Create(fPath)
		if err != nil {
			return nil, fmt.Errorf("couldn't create file: %s (path: %s)", err, fPath)
		}
		defer file.Close()
		_, err = io.Copy(file, reader)
		if err != nil {
			return nil, err
		}
		result = map[string]interface{}{}
		for k, v := range item {
			result[k] = v
		}
		result["file_path"] = fPath
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		result["file_size"] = fileInfo.Size()
		return result, nil
	}
	recordPicture := func(item module.Item) (result module.Item, err error) {
		v := item["file_path"]
		path, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("incorrect file path type: %T", v)
		}
		v = item["file_size"]
		size, ok := v.(int64)
		if !ok {
			return nil, fmt.Errorf("incorrect file name type: %T", v)
		}
		logger.Infof("Saved file: %s, size: %d byte(s)", path, size)
		return nil, nil
	}
	return []module.ProcessItem{savePic, recordPicture}
}
