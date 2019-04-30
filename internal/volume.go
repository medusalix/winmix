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
	"errors"

	wca "github.com/moutend/go-wca"
)

// VolumeControl is used to control the volume of an individual process
type VolumeControl struct {
	volumes []*wca.ISimpleAudioVolume
}

// NewVolumeControl constructs a new control instance from a process
func NewVolumeControl(process *Process) (*VolumeControl, error) {
	volumes, err := getProcessVolumes(process)

	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, errors.New("no audio sessions found")
	}

	return &VolumeControl{
		volumes: volumes,
	}, nil
}

// GetVolume gets the volume level of the process
// Volume is in the range of 0 to 100
func (c *VolumeControl) GetVolume() (int, error) {
	var level float32

	if err := c.volumes[0].GetMasterVolume(&level); err != nil {
		return 0, err
	}

	return int(level * 100), nil
}

// SetVolume sets the volume level of the process
// Volume is in the range of 0 to 100
func (c *VolumeControl) SetVolume(level int) error {
	actualLevel := float32(level) / 100

	for _, volume := range c.volumes {
		if err := volume.SetMasterVolume(actualLevel, nil); err != nil {
			return err
		}
	}

	return nil
}

// Release frees the allocated resources
func (c *VolumeControl) Release() {
	for _, volume := range c.volumes {
		volume.Release()
	}
}
