package main

import (
	"github.com/garfcat/msync/msync"
	"github.com/spf13/pflag"
)

var (
	threadCount *int
	srcPath     *string
	destPath    *string
)

func init() {
	threadCount = pflag.IntP("thread", "t", 1, "the num of max thread")
	srcPath = pflag.StringP("srcPath", "s", "", "source directory")
	destPath = pflag.StringP("dstPath", "d", "", "destination directory")
	pflag.Parse()
	if len(*srcPath) == 0 || len(*destPath) == 0 {
		pflag.PrintDefaults()
		return
	}
}

func main() {
	r := msync.New(*srcPath, *destPath, *threadCount)
	r.StartWorker()
	err := r.Sync()
	if err != nil {
		r.Done()
		panic(err)
	}
	r.Done()
	r.Wait()
}
