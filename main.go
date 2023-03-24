package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var videoExtensions = []string{".mp4", ".avi", ".rmvb", ".wmv", ".m4v", ".mkv", ".webm", ".mov", ".mpg", ".flv"}

// 判断文件是否是视频
func IsVideo(filename string) bool {
	for _, ext := range videoExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// 返回文件大小，单位byte
func FileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	} else {
		return fi.Size()
	}
}

// 获取视频时长
func VideoLength(path string) (time.Duration, error) {
	cmd := exec.Command("ffmpeg", "-i", path)
	var out bytes.Buffer
	backup := cmd.Stderr
	cmd.Stderr = &out
	_ = cmd.Run()
	cmd.Stderr = backup

	r, _ := regexp.Compile(`Duration: (\d+):(\d+):([0-9.]+)`)
	duration := r.FindSubmatch(out.Bytes())
	if duration == nil {
		return 0, fmt.Errorf("ffmpeg can't read video info of %s", path)
	}

	hour := string(duration[1])
	minute := string(duration[2])
	second := string(duration[3])
	return time.ParseDuration(fmt.Sprintf("%sh%sm%ss", hour, minute, second))
}

// 格式化文件大小
func sizeFormat(size int64) string {
	var K, M, G int64 = 1024, 1024*1024, 1024*1024*1024
	if size > G {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(G))
	} else if size > M {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(M))
	} else if size > K {
		return fmt.Sprintf("%.2f KB", float64(size)/float64(K))
	} else {
		return fmt.Sprintf("%.2f B", float64(size))
	}
}

type VideoTotal struct {
	root      string
	count     int
	duration  time.Duration
	filesSize int64 //bytes
}

func NewVideoTotal(path string) VideoTotal {
	v := VideoTotal{path, 0, 0, 0}
	return v
}

func (v *VideoTotal) Scan(){
	file := make([]string, 10)
	err := filepath.Walk(v.root, func(path string, info os.FileInfo, err error) error {
		//skip error file/path
		if err == nil {
			file = append(file, path)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	count := 0
	var seconds float64 = 0
	var size int64 = 0
	for _, f := range file {
		if IsVideo(f) {
			duration, err := VideoLength(f)
			if err != nil {
				log.Println(err)
			} else {
				count++
				seconds += duration.Seconds()
				size += FileSize(f)
			}
		}
	}

	v.count = count
	v.duration, _ = time.ParseDuration(fmt.Sprintf("%fs", seconds))
	v.filesSize = size
}

func main() {
	root, _ := os.Getwd()
	flag.Parse()
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}
	fmt.Printf("videos in %s:\n", root)
	v := NewVideoTotal(root)
	v.Scan()
	fmt.Printf("%d videos, %s, %s", v.count, v.duration, sizeFormat(v.filesSize))
}
