package main

import (
	"errors"
	"fmt"
	"github.com/syupei/glog"
	"github.com/syupei/goconfig"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	retry   int64
	sleep   int64
	step    float64
	command string
	outfile string
	isLog   bool
)

var err error
var log *glog.Log
var fd *(os.File)

func main() {

	err := parseConf()
	if err != nil {
		return
	}

	log, _ = glog.New(glog.Log{Level: 31, Split: "500m"}, glog.LevelConf{IsPrint: true, IsWrite: isLog, FileName: "supervisor.log"})

	var cmd *(exec.Cmd)

	if len(os.Args) < 2 {
		log.Error("no cmd need run!")
		return
	}

	outfd, err := os.OpenFile(outfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0655)

	if err != nil {
		log.Warning("output file create error!")
	}

	command := os.Args[1]
	var i int = 0
	var s float64 = 0

	//ignore SIGHUP
	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		_ = <-sigc
		//nothing
	}()

	for i <= int(retry) {
		if i > 0 {
			s = float64(sleep) * math.Pow(step, float64(i-1))
		}

		log.Info("wait sleep %d Millisecond...", int(s))
		time.Sleep(time.Duration(s) * time.Millisecond)

		cmd = exec.Command(command)
		cmd.Args = os.Args[1:]

		//bind stdout
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Warning("bind STDOUT error:%s", err.Error())
		}

		//bind stderr
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Warning("bind STDERR error:%s", err.Error())
		}

		//bind stdin
		//stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Warning("bind STDIN error:%s", err.Error())
		}

		if err := cmd.Start(); err != nil {
			log.Warning("cmd [%s] run error:%s, retry %d.", strings.Join(cmd.Args, " "), err.Error(), i)
			i++
			continue
		}

		log.Notice("cmd [%s] run ok! pid:%d", strings.Join(cmd.Args, " "), cmd.Process.Pid)

		go io.Copy(outfd, stdout)
		go io.Copy(outfd, stderr)

		cmd.Wait()

		i++
		log.Warning("process %d die!", cmd.Process.Pid)
	}

	log.Error("Retry to limit !!!")
	log.Close()
}

//parse config
func parseConf() (err error) {
	c, err := goconfig.ReadConfigFile("supervisor.conf")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	retry, err = c.GetInt64("default", "retry")

	sleep, err = c.GetInt64("default", "sleep")

	outfile, err = c.GetString("default", "out")

	step, err = c.GetFloat("default", "step")

	isLog, err = c.GetBool("default", "log")

	if err = valid(); err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}

// valid config
func valid() (err error) {

	if err = validRetry(); err != nil {
		return
	}

	if err = validStep(); err != nil {
		return
	}

	if err = validSleep(); err != nil {
		return
	}

	outfile = strings.Trim(outfile, " ")
	if outfile == "" {
		outfile = "nohup.out"
	}

	return
}

func validRetry() (err error) {
	if retry > 99999 || retry < 0 {
		err = errors.New("retry limit 0-99999")
	}
	return
}

func validSleep() (err error) {
	if sleep < 0 {
		err = errors.New("sleep must > 1")
	}
	return
}

func validStep() (err error) {
	if step < 0 || step > 1000 {
		err = errors.New("step limit 0-1000")
	}
	if step == 0 {
		step = 1
	}

	return
}
