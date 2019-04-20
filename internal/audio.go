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
	"unsafe"

	ole "github.com/go-ole/go-ole"
	wca "github.com/medusalix/go-wca"
)

func getProcessVolumes(process *Process) ([]*wca.ISimpleAudioVolume, error) {
	device, err := getAudioDevice()

	if err != nil {
		return nil, err
	}
	defer device.Release()

	enumerator, err := getSessionEnumerator(device)

	if err != nil {
		return nil, err
	}
	defer enumerator.Release()

	return getSessionVolumes(enumerator, process)
}

func getAudioDevice() (*wca.IMMDevice, error) {
	var deviceEnumerator *wca.IMMDeviceEnumerator

	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&deviceEnumerator,
	); err != nil {
		return nil, err
	}
	defer deviceEnumerator.Release()

	var device *wca.IMMDevice

	return device, deviceEnumerator.GetDefaultAudioEndpoint(
		wca.ERender,
		wca.EMultimedia,
		&device,
	)
}

func getSessionEnumerator(device *wca.IMMDevice) (*wca.IAudioSessionEnumerator, error) {
	var manager *wca.IAudioSessionManager2

	if err := device.Activate(
		wca.IID_IAudioSessionManager2,
		wca.CLSCTX_ALL,
		nil,
		&manager,
	); err != nil {
		return nil, err
	}
	defer manager.Release()

	var enumerator *wca.IAudioSessionEnumerator

	return enumerator, manager.GetSessionEnumerator(&enumerator)
}

func getSessionVolumes(enumerator *wca.IAudioSessionEnumerator, process *Process) ([]*wca.ISimpleAudioVolume, error) {
	var count int

	if err := enumerator.GetCount(&count); err != nil {
		return nil, err
	}

	volumes := make([]*wca.ISimpleAudioVolume, 0)

	for i := 0; i < count; i++ {
		var session *wca.IAudioSessionControl

		if err := enumerator.GetSession(i, &session); err != nil {
			return nil, err
		}

		dispatch, err := session.QueryInterface(wca.IID_IAudioSessionControl2)
		session.Release()

		if err != nil {
			return nil, err
		}

		session2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))

		var pid uint32

		if err := session2.GetProcessId(&pid); err != nil {
			// Ignore AUDCLNT_S_NO_CURRENT_PROCESS (0x889000D) - no error
			if err.(*ole.OleError).Code() != 0x889000D {
				session2.Release()

				continue
			}
		}

		if process.hasPid(pid) {
			dispatch, err := session2.QueryInterface(wca.IID_ISimpleAudioVolume)

			if err != nil {
				return nil, err
			}

			volume := (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch))
			volumes = append(volumes, volume)
		}

		session2.Release()
	}

	return volumes, nil
}
