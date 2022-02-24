package msync

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

type Rsync interface {
	StartWorker()
	Sync() error
	Wait()
	Done()
}

type rsync struct {
	srcPath   string
	dstPath   string
	threadNum int
	ch        chan os.FileInfo
	waitGroup sync.WaitGroup
	done      chan struct{}
}

func New(src string, dst string, num int) Rsync {
	return &rsync{
		srcPath:   src,
		dstPath:   dst,
		threadNum: num,
		done:      make(chan struct{}),
		waitGroup: sync.WaitGroup{},
		ch:        make(chan os.FileInfo, num),
	}
}

func (r *rsync) GetSourceFileList() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(r.srcPath)
	if err != nil {
		return nil, errors.Wrapf(err, "get directory:%s file list error", r.srcPath)
	}
	return files, nil
}
func (r *rsync) Sync() error {

	fileLists, err := r.GetSourceFileList()
	if err != nil {
		return errors.Wrapf(err, "get file list error")
	}
	log.Printf("prepare sync filelist %d\n", len(fileLists))
	for _, v := range fileLists {
		r.ch <- v
	}
	return nil
}
func (r *rsync) sync(fileInfo os.FileInfo) error {
	name := fileInfo.Name()
	var cmd string
	if fileInfo.IsDir() {
		cmd = fmt.Sprintf("rclone sync -P   %s/%s %s/%s", r.srcPath, name, r.dstPath, name)
	} else {
		cmd = fmt.Sprintf("rsync -avzr -P  %s/%s %s/%s", r.srcPath, name, r.dstPath, name)
	}

	command := exec.Command("sh", "-c", cmd)
	out, err := command.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "exec cmd error")
	}
	log.Printf("cmd[%s] , msg[%s]\n", cmd, string(out))
	return nil
}
func (r *rsync) StartWorker() {
	log.Printf("prepare start worker num %d\n", r.threadNum)
	r.waitGroup.Add(r.threadNum)
	for i := 0; i < r.threadNum; i++ {
		go func(index int) {
			r.Worker(index)
		}(i)
	}
}
func (r *rsync) Worker(index int) error {
	log.Printf("start worker %d\n", index)
	for {
		select {
		case fileInfo := <-r.ch:
			err := r.sync(fileInfo)
			if err != nil {
				log.Printf("worker:%d sync:%s error:%s\n", index, fileInfo.Name(), err.Error())
			}
		case <-r.done:
			log.Printf("worker %d quit because sync done\n", index)
			r.waitGroup.Done()
			return nil
		}
	}
}

func (r *rsync) Wait() {
	r.waitGroup.Wait()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	select {
	case <-sig:
		log.Println("sync done")
	}
}

func (r *rsync) Done() {
	close(r.done)
}
