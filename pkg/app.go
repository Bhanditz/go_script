package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Script struct {
	sync.Mutex
	Command    string `json:"command"`
	Log        string `json:"log"`
	ArchiveLog string `json:"logArchive"`
}

func ReadFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	return string(data), err
}

func (s *Script) ReadConfig() {
	s.Command = "body() { IFS= read -r header; printf '%s %s\n %s\n' `date \"+%Y-%m %H:%M:%S\"` \"$header\"; \"$@\"; } && ps aux| body sort -n -r -k 4 && free"
	s.Log = "/tmp/s.log"
	s.ArchiveLog = "/tmp/archive/s.log"

}

type Writer interface {
	Write(file string, data []byte) (n int, err error)
}

func ZeroOut(file string) (int64, error) {

	_ = os.Remove(file)
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	if fi, err := f.Stat(); err != nil {
		return -1, err
	} else {
		f.WriteString(fmt.Sprintf("file size:%v\n", fi.Size()))
		return fi.Size(), err
	}
}

func WriteData(file string, data []byte) (int, int64, error) {

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return -1, -1, err
	}
	defer f.Close()

	n, e := f.Write(data)
	if fi, err := f.Stat(); err != nil {
		return n, -1, err
	} else {
		f.WriteString(fmt.Sprintf("file size:%v\n", fi.Size()))
		return n, fi.Size(), e
	}
}

// LogProcess:
//   Run process and return bytes written and total bytes in file
func (s *Script) LogProcess(ctx context.Context) (int, int64) {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	//cmd := exec.Command("sh", "-c", "date '+%Y-%m-%d %H:%M:%S\n' 1>&2;top -b -n1 -c -w 400 -o +%MEM|head -n30 1>&2")
	cmd := exec.CommandContext(ctx, "sh", "-c", s.Command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	n, fileSize, err := WriteData(s.Log, output)
	if err != nil {
		panic(err)
	}
	return n, fileSize

}
