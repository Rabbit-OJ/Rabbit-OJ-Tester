package tester

import (
	"context"
	"log"
	"math"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func TestOne(
	testResult *TestResult,
	i, timeLimit, spaceLimit int64,
	execCommand string,
	execArgs []string) {

	cmd := exec.Command(execCommand, execArgs...)
	peakMemory := float64(0)

	in, err := os.OpenFile(DockerCasePath(i), os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		testResult.Status = StatusRE
		return
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.OpenFile(DockerOutputPath(i), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		_ = out.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	errChan, memoryMonitorChan := make(chan error), make(chan bool)
	defer func() {
		cancel()

		close(errChan)
		close(memoryMonitorChan)
	}()

	cmd.Stdin, cmd.Stdout = in, out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		log.Println(err)
		testResult.Status = StatusRE
		return
	}
	startTime := time.Now()

	go func() {
		waitChan := make(chan error)

		go func() {
			err := cmd.Wait()
			if err != nil && ctx.Err() == nil {
				waitChan <- err
			}
			close(waitChan)
		}()

		select {
		case <-ctx.Done():
			return
		case ans := <-waitChan:
			if ctx.Err() == nil {
				errChan <- ans
			}
		}
	}()

	go func(pid int) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				stat, err := GetStat(pid)
				if err == nil {
					peakMemory = math.Max(peakMemory,
						stat.Memory/1024.0,
					)

					if peakMemory >= float64(spaceLimit)*1024.0 {
						memoryMonitorChan <- true
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}(cmd.Process.Pid)

	select {
	case <-memoryMonitorChan:
		testResult.Status = StatusMLE
		testResult.TimeUsed = uint32(timeLimit)
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Println(err)
		}
	case <-time.After(time.Duration(timeLimit) * time.Millisecond):
		testResult.Status = StatusTLE
		testResult.TimeUsed = uint32(timeLimit)
		testResult.SpaceUsed = uint32(peakMemory)
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Println(err)
		}
	case err := <-errChan:
		usedTime := time.Since(startTime)

		if err != nil {
			log.Println(err)
			testResult.Status = StatusRE
		} else {
			testResult.Status = StatusOK
		}

		testResult.TimeUsed = uint32(usedTime.Milliseconds())
		testResult.SpaceUsed = uint32(peakMemory)
	}
}
