package cache

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type FileCache struct{}

func (fc *FileCache) WriteCache(file string, v string) (err error) {
	bt, _ := json.Marshal(v)
	month := time.Now().Month().String()
	cacheDir := getCacheDir() + month + "/"
	_ = os.MkdirAll(cacheDir, 0660)
	file = cacheDir + file + ".json"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer logFile.Close()
	if err != nil {
		return err
	}
	// 检查过期文件
	//checkCacheOvertimeFile()
	_, err = io.WriteString(logFile, string(bt))
	return
}

func (fc *FileCache) ReadCache(file string) (value string, err error) {
	month := time.Now().Month().String()
	cacheDir := getCacheDir() + month + "/"
	file = cacheDir + file + ".json"

	if !checkFileIsExist(file) {
		return "", errors.New("验证码已过期")
	}

	bt, err := ioutil.ReadFile(file)
	err = os.Remove(file)
	if err == nil {
		return string(bt), nil
	}
	return "", err

}

func (fc *FileCache) Expire(key string, duration time.Duration) (err error) {
	return
}
func (fc *FileCache) RemoveCache(key string) (err error) {
	return os.Remove(getCacheDir() + key)
}

func (fc *FileCache) KeyExists(key string) (exists bool) {
	return checkFileIsExist(key)
}

/**
 * @Description: Get cache dir path
 * @return string
 */
func getCacheDir() string {
	return getPWD() + "/.cache/"
}

/**
 * @Description: Get pwd dir path
 * @return string
 */
func getPWD() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path
}

/**
 * @Description: Check file exist
 * @param filename
 * @return bool
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

/**
 * @Description: 启动定时任务, 5分钟执行一次
 */
func runTimedTask() {
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			checkCacheOvertimeFile()
		}
	}()
}
func GetFileCreateTime(path string) int64 {
	return time.Now().Unix()
}

/**
 * @Description: 检查缓存超时文件， 30分钟
 */
func checkCacheOvertimeFile() {
	files, files1, _ := listDir(getCacheDir())
	for _, table := range files1 {
		temp, _, _ := listDir(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	for _, file := range files {
		t := GetFileCreateTime(file)
		ex := time.Now().Unix() - t
		if ex > (60 * 30) {
			_ = os.Remove(file)
		}
	}
}

/**
 * @Description: 获取目录文件列表
 * @param dirPth
 * @return files
 * @return files1
 * @return err
 */
func listDir(dirPth string) (files []string, files1 []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			files1 = append(files1, dirPth+PthSep+fi.Name())
			_, _, _ = listDir(dirPth + PthSep + fi.Name())
		} else {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, files1, nil
}
