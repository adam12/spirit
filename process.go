package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Process struct {
	name          string
	command       string
	arguments     []string
	pidFile       string
	daemonPidFile string
	logFile       string
}

func (p *Process) start() error {
	// If Pidfile exists and is process running, return
	if _, err := os.Stat(p.pidFile); os.IsExist(err) && p.isRunning() {
		return nil
	}

	cmdLine := []string{p.command}
	cmdLine = append(cmdLine, p.arguments...)

	args := []string{
		"-t", p.name,
		"-r",
		"-o", p.logFile,
		"-p", p.pidFile,
		"-P", p.daemonPidFile,
		"sh", "-c",
		strings.Join(cmdLine, " "),
	}

	path, err := exec.LookPath("daemon")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, args...)

	return cmd.Run()
}

func (p *Process) stop() error {
	if _, err := os.Stat(p.daemonPidFile); os.IsNotExist(err) {
		return nil
	}

	pid, err := p.getDaemonPid()
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err = proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	for i := 1; i <= 12; i++ {
		if err = proc.Signal(syscall.Signal(0)); err != nil {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("Process with pid %d never exited", pid)
}

func (p *Process) restart() error {
	if err := p.stop(); err != nil {
		return err
	}

	if err := p.start(); err != nil {
		return err
	}

	return nil
}

func (p *Process) status() string {
	// no daemon pid, no process pid = stopped
	if _, err := os.Stat(p.pidFile); os.IsNotExist(err) {
		if _, err := os.Stat(p.daemonPidFile); os.IsNotExist(err) {
			return "stopped"
		}
	}

	// process pid && running = running
	if p.isRunning() {
		return "running"
	}

	// else dead
	return "dead"
}

func (p *Process) isRunning() bool {
	pid, err := p.getPid()
	if err != nil {
		panic(err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		panic(err)
	}

	return proc.Signal(syscall.Signal(0)) == nil
}

func (p *Process) getDaemonPid() (int, error) {
	if _, err := os.Stat(p.daemonPidFile); os.IsNotExist(err) {
		return 0, err
	}

	data, err := ioutil.ReadFile(p.daemonPidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func (p *Process) getPid() (int, error) {
	if _, err := os.Stat(p.pidFile); os.IsNotExist(err) {
		return 0, err
	}

	data, err := ioutil.ReadFile(p.pidFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

func (p *Process) viewLog() error {
	cmd := exec.Command("less", p.logFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (p *Process) tailLog() error {
	cmd := exec.Command("tail", "-f", p.logFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func NewProcess(name, command string, arguments []string) *Process {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &Process{
		name:          name,
		command:       command,
		arguments:     arguments,
		pidFile:       filepath.Join(cwd, "tmp", "pids", name+".pid"),
		daemonPidFile: filepath.Join(cwd, "tmp", "pids", name+".daemon.pid"),
		logFile:       filepath.Join(cwd, "tmp", "logs", name+".log"),
	}
}
