/*
 * Copyright (C) 2019 Medusalix
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package internal

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Process represents all pids of a process
type Process struct {
	pids []uint32
}

// NewProcess finds a process from its name
func NewProcess(name string) (*Process, error) {
	pids := make([]uint32, 0)

	// Create a snapshot of all processes - TH32CS_SNAPPROCESS (0x00000002)
	handle, err := syscall.CreateToolhelp32Snapshot(0x00000002, 0)

	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(handle)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := syscall.Process32First(handle, &entry); err != nil {
		return nil, err
	}

	if getProcessName(&entry) == name {
		pids = append(pids, entry.ProcessID)
	}

	for {
		if err := syscall.Process32Next(handle, &entry); err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				break
			}

			return nil, err
		}

		if getProcessName(&entry) == name {
			pids = append(pids, entry.ProcessID)
		}
	}

	if len(pids) == 0 {
		return nil, fmt.Errorf("no process with name '%s' found", name)
	}

	return &Process{
		pids: pids,
	}, nil
}

func (p *Process) hasPid(pid uint32) bool {
	for _, p := range p.pids {
		if p == pid {
			return true
		}
	}

	return false
}

func getProcessName(entry *syscall.ProcessEntry32) string {
	size := len(entry.ExeFile)

	for i := 0; i < size; i++ {
		if entry.ExeFile[i] == 0 {
			return syscall.UTF16ToString(entry.ExeFile[:i])
		}
	}

	return ""
}
