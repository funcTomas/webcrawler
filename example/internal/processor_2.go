package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"webcrawler/module"
)

func genItemProcessorsV2(dirPath string) []module.ProcessItem {
	saveText := func(item module.Item) (result module.Item, err error) {
		if item == nil {
			return nil, errors.New("invalid item")
		}
		var absDirPath string
		if absDirPath, err = checkDirPath(dirPath); err != nil {
			return
		}
		srcUrl, ok := item["srcUrl"].(string)
		if !ok {
			return nil, fmt.Errorf("incorrect name type: %T", srcUrl)
		}
		title, ok := item["title"].(string)
		if !ok {
			return nil, fmt.Errorf("incorrect title type: %T", title)
		}
		catagory, ok := item["catagory"].(string)
		if !ok {
			return nil, fmt.Errorf("incorrect catagory type: %T", catagory)
		}
		content, ok := item["content"].(string)
		if !ok {
			return nil, fmt.Errorf("incorrect content type: %T", content)
		}
		fileName := catagory
		fPath := filepath.Join(absDirPath, fileName)
		file, err := os.OpenFile(fPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, fmt.Errorf("couldn't create file: %s (path: %s)", err, fPath)
		}
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
		if err != nil {
			return nil, fmt.Errorf("couldn't lock file: %s (path: %s)", err, fPath)
		}
		defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		defer file.Close()
		file.WriteString(title + "\n")
		file.WriteString(content + "\n")
		file.WriteString(srcUrl + "\n\n")
		result = map[string]any{}
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
	fileCheck := func(item module.Item) (result module.Item, err error) {
		v := item["file_path"]
		path, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("incorrect file path type: %T", v)
		}
		v = item["file_size"]
		size, ok := v.(int64)
		if !ok {
			return nil, fmt.Errorf("incorrect file size type: %T", v)
		}
		logger.Infof("Saved file: %s, size: %d byte(s)", path, size)
		return nil, nil
	}
	return []module.ProcessItem{saveText, fileCheck}
}
