// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package bpm

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/go-logr/logr"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

type NsType string

const (
	MountNS NsType = "mnt"
	// uts namespace is not supported yet
	// UtsNS   NsType = "uts"
	IpcNS NsType = "ipc"
	NetNS NsType = "net"
	PidNS NsType = "pid"
	// user namespace is not supported yet
	// UserNS  NsType = "user"
)

var nsArgMap = map[NsType]string{
	MountNS: "m",
	// uts namespace is not supported by nsexec yet
	// UtsNS:   "u",
	IpcNS: "i",
	NetNS: "n",
	PidNS: "p",
	// user namespace is not supported by nsexec yet
	// UserNS:  "U",
}

const (
	pausePath  = "/usr/local/bin/pause"
	nsexecPath = "/usr/local/bin/nsexec"

	DefaultProcPrefix = "/proc"
)

// ProcessPair is an identifier for process
type ProcessPair struct {
	Pid        int
	CreateTime int64
}

// Stdio contains stdin, stdout and stderr
type Stdio struct {
	sync.Locker
	Stdin, Stdout, Stderr io.ReadWriteCloser
}

// BackgroundProcessManager manages all background processes
type BackgroundProcessManager struct {
	deathSig    *sync.Map
	identifiers *sync.Map
	stdio       *sync.Map

	rootLogger logr.Logger
}

// NewBackgroundProcessManager creates a background process manager
func NewBackgroundProcessManager(rootLogger logr.Logger) BackgroundProcessManager {
	return BackgroundProcessManager{
		deathSig:    &sync.Map{},
		identifiers: &sync.Map{},
		stdio:       &sync.Map{},
		rootLogger:  rootLogger,
	}
}

// StartProcess manages a process in manager
func (m *BackgroundProcessManager) StartProcess(cmd *ManagedProcess) (*process.Process, error) {
	var identifierLock *sync.Mutex
	if cmd.Identifier != nil {
		lock, _ := m.identifiers.LoadOrStore(*cmd.Identifier, &sync.Mutex{})

		identifierLock = lock.(*sync.Mutex)

		identifierLock.Lock()
	}

	err := cmd.Start()
	if err != nil {
		m.rootLogger.Error(err, "fail to start process")
		return nil, err
	}

	pid := cmd.Process.Pid
	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}
	ct, err := procState.CreateTime()
	if err != nil {
		return nil, err
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: ct,
	}

	channel, _ := m.deathSig.LoadOrStore(pair, make(chan bool, 1))
	deathChannel := channel.(chan bool)

	stdio := &Stdio{Locker: &sync.Mutex{}}
	if cmd.Stdin != nil {
		if stdin, ok := cmd.Stdin.(io.ReadWriteCloser); ok {
			stdio.Stdin = stdin
		}
	}

	if cmd.Stdout != nil {
		if stdout, ok := cmd.Stdout.(io.ReadWriteCloser); ok {
			stdio.Stdout = stdout
		}
	}

	if cmd.Stderr != nil {
		if stderr, ok := cmd.Stderr.(io.ReadWriteCloser); ok {
			stdio.Stderr = stderr
		}
	}

	m.stdio.Store(pair, stdio)

	logger := m.rootLogger.WithValues("pid", pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				status := exitErr.Sys().(syscall.WaitStatus)
				if status.Signaled() && status.Signal() == syscall.SIGTERM {
					logger.Info("process stopped with SIGTERM signal")
				}
			} else {
				logger.Error(err, "process exited accidentally")
			}
		}

		logger.Info("process stopped")

		deathChannel <- true
		m.deathSig.Delete(pair)
		if io, loaded := m.stdio.LoadAndDelete(pair); loaded {
			if stdio, ok := io.(*Stdio); ok {
				stdio.Lock()
				if stdio.Stdin != nil {
					if err = stdio.Stdin.Close(); err != nil {
						logger.Error(err, "stdin fails to be closed")
					}
				}
				if stdio.Stdout != nil {
					if err = stdio.Stdout.Close(); err != nil {
						logger.Error(err, "stdout fails to be closed")
					}
				}
				if stdio.Stderr != nil {
					if err = stdio.Stderr.Close(); err != nil {
						logger.Error(err, "stderr fails to be closed")
					}
				}
				stdio.Unlock()
			}
		}

		if identifierLock != nil {
			identifierLock.Unlock()
			m.identifiers.Delete(*cmd.Identifier)
		}
	}()

	return procState, nil
}

// KillBackgroundProcess sends SIGTERM to process
func (m *BackgroundProcessManager) KillBackgroundProcess(ctx context.Context, pid int, startTime int64) error {
	logger := m.rootLogger.WithValues("pid", pid)

	p, err := os.FindProcess(int(pid))
	if err != nil {
		logger.Error(err, "unreachable path. `os.FindProcess` will never return an error on unix")
		return err
	}

	procState, err := process.NewProcess(int32(pid))
	if err != nil {
		// return successfully as the process has exited
		return nil
	}
	ct, err := procState.CreateTime()
	if err != nil {
		logger.Error(err, "fail to read create time")
		// return successfully as the process has exited
		return nil
	}

	// There is a bug in calculating CreateTime in the new version of
	// gopsutils. This is a temporary solution before the upstream fixes it.
	if startTime-ct > 1000 || ct-startTime > 1000 {
		logger.Info("process has already been killed", "startTime", ct, "expectedStartTime", startTime)
		// return successfully as the process has exited
		return nil
	}

	ppid, err := procState.Ppid()
	if err != nil {
		logger.Error(err, "fail to read parent id")
		// return successfully as the process has exited
		return nil
	}
	if ppid != int32(os.Getpid()) {
		logger.Info("process has already been killed", "ppid", ppid)
		// return successfully as the process has exited
		return nil
	}

	err = p.Signal(syscall.SIGTERM)

	if err != nil && err.Error() != "os: process already finished" {
		logger.Error(err, "error while killing process")
		return err
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: startTime,
	}
	channel, ok := m.deathSig.Load(pair)
	if ok {
		deathChannel := channel.(chan bool)
		select {
		case <-deathChannel:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	logger.Info("Successfully killed process")
	return nil
}

func (m *BackgroundProcessManager) Stdio(pid int, startTime int64) *Stdio {
	logger := m.rootLogger.WithValues("pid", pid)

	procState, err := process.NewProcess(int32(pid))
	if err != nil {
		logger.Info("fail to get process information", "pid", pid)
		// return successfully as the process has exited
		return nil
	}
	ct, err := procState.CreateTime()
	if err != nil {
		logger.Error(err, "fail to read create time")
		// return successfully as the process has exited
		return nil
	}

	// There is a bug in calculating CreateTime in the new version of
	// gopsutils. This is a temporary solution before the upstream fixes it.
	if startTime-ct > 1000 || ct-startTime > 1000 {
		logger.Info("process has exited", "startTime", ct, "expectedStartTime", startTime)
		// return successfully as the process has exited
		return nil
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: startTime,
	}

	io, ok := m.stdio.Load(pair)
	if !ok {
		logger.Info("fail to load with pair", "pair", pair)
		// stdio is not stored
		return nil
	}

	return io.(*Stdio)
}

// DefaultProcessBuilder returns the default process builder
func DefaultProcessBuilder(cmd string, args ...string) *ProcessBuilder {
	return &ProcessBuilder{
		cmd:        cmd,
		args:       args,
		nsOptions:  []nsOption{},
		pause:      false,
		identifier: nil,
		ctx:        context.Background(),
	}
}

// ProcessBuilder builds a exec.Cmd for daemon
type ProcessBuilder struct {
	cmd  string
	args []string
	env  []string

	nsOptions []nsOption

	pause    bool
	localMnt bool

	identifier *string
	stdin      io.ReadWriteCloser
	stdout     io.ReadWriteCloser
	stderr     io.ReadWriteCloser

	ctx    context.Context
	logger logr.Logger
}

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ NsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", DefaultProcPrefix, pid, string(typ))
}

// SetEnv sets the environment variables of the process
func (b *ProcessBuilder) SetEnv(key, value string) *ProcessBuilder {
	b.env = append(b.env, fmt.Sprintf("%s=%s", key, value))
	return b
}

// SetNS sets the namespace of the process
func (b *ProcessBuilder) SetNS(pid uint32, typ NsType) *ProcessBuilder {
	return b.SetNSOpt([]nsOption{{
		Typ:  typ,
		Path: GetNsPath(pid, typ),
	}})
}

// SetNSOpt sets the namespace of the process
func (b *ProcessBuilder) SetNSOpt(options []nsOption) *ProcessBuilder {
	b.nsOptions = append(b.nsOptions, options...)

	return b
}

// SetIdentifier sets the identifier of the process
//
// The identifier is used to identify the process in BPM, to confirm only one identified process is running.
// If one identified process is already running, new processes with the same identifier will be blocked by lock.
func (b *ProcessBuilder) SetIdentifier(id string) *ProcessBuilder {
	b.identifier = &id

	return b
}

// EnablePause enables pause for process
func (b *ProcessBuilder) EnablePause() *ProcessBuilder {
	b.pause = true

	return b
}

func (b *ProcessBuilder) EnableLocalMnt() *ProcessBuilder {
	b.localMnt = true

	return b
}

// SetContext sets context for process
func (b *ProcessBuilder) SetContext(ctx context.Context) *ProcessBuilder {
	b.ctx = ctx

	return b
}

// SetStdin sets stdin for process
func (b *ProcessBuilder) SetStdin(stdin io.ReadWriteCloser) *ProcessBuilder {
	b.stdin = stdin

	return b
}

// SetStdout sets stdout for process
func (b *ProcessBuilder) SetStdout(stdout io.ReadWriteCloser) *ProcessBuilder {
	b.stdout = stdout

	return b
}

// SetStderr sets stderr for process
func (b *ProcessBuilder) SetStderr(stderr io.ReadWriteCloser) *ProcessBuilder {
	b.stderr = stderr

	return b
}

func (b *ProcessBuilder) getLoggerFromContext(ctx context.Context) logr.Logger {
	// this logger is inherited from the global one
	// TODO: replace it with a specific logger by passing in one or creating a new one
	logger := log.L().WithName("background-process-manager.process-builder")
	return log.EnrichLoggerWithContext(ctx, logger)
}

type nsOption struct {
	Typ  NsType
	Path string
}

// ManagedProcess is a process which can be managed by backgroundProcessManager
type ManagedProcess struct {
	*exec.Cmd

	// If the identifier is not nil, process manager should make sure no other
	// process with this identifier is running when executing this command
	Identifier *string
}
