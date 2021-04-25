// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// +build cgo

package ptrace

// RegisterLogger registers a logger on ptrace pkg
func RegisterLogger(logger logr.Logger) {
	// no implement
}

// TracedProgram is a program traced by ptrace
type TracedProgram struct {
}

// Pid return the pid of traced program
func (p *TracedProgram) Pid() int {
	return 0
}

// Trace ptrace all threads of a process
func Trace(pid int) (*TracedProgram, error) {
	return nil, nil
}

// Detach detaches from all threads of the processes
func (p *TracedProgram) Detach() error {
	return nil
}

// Protect will backup regs and rip into fields
func (p *TracedProgram) Protect() error {
	return nil
}

// Restore will restore regs and rip from fields
func (p *TracedProgram) Restore() error {
	return nil
}

// Wait waits until the process stops
func (p *TracedProgram) Wait() error {
	return nil
}

// Step moves one step forward
func (p *TracedProgram) Step() error {
	return nil
}

// Syscall runs a syscall at main thread of process
func (p *TracedProgram) Syscall(number uint64, args ...uint64) (uint64, error) {
	return 0, nil
}

// Mmap runs mmap syscall
func (p *TracedProgram) Mmap(length uint64, fd uint64) (uint64, error) {
	return 0, nil
}

// ReadSlice reads from addr and return a slice
func (p *TracedProgram) ReadSlice(addr uint64, size uint64) (*[]byte, error) {
	return nil, nil
}

// WriteSlice writes a buffer into addr
func (p *TracedProgram) WriteSlice(addr uint64, buffer []byte) error {
	return nil
}

// PtraceWriteSlice uses ptrace rather than process_vm_write to write a buffer into addr
func (p *TracedProgram) PtraceWriteSlice(addr uint64, buffer []byte) error {
	return nil
}

// GetLibBuffer reads an entry
func (p *TracedProgram) GetLibBuffer(entry *mapreader.Entry) (*[]byte, error) {
	return nil, nil
}

// MmapSlice mmaps a slice and return it's addr
func (p *TracedProgram) MmapSlice(slice []byte) (*mapreader.Entry, error) {
	return nil, nil
}

// FindSymbolInEntry finds symbol in entry through parsing elf
func (p *TracedProgram) FindSymbolInEntry(symbolName string, entry *mapreader.Entry) (uint64, error) {
	return 0, nil
}

// WriteUint64ToAddr writes uint64 to addr
func (p *TracedProgram) WriteUint64ToAddr(addr uint64, value uint64) error {
	return nil
}

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64) error {
	return nil
}
