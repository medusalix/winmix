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

package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/medusalix/winmix/internal"

	ole "github.com/go-ole/go-ole"
)

var commands = map[string]func(control *internal.VolumeControl, args []string){
	"get":    getVolume,
	"set":    setVolume,
	"change": changeVolume,
}

func main() {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing COM: %s.", err)

		return
	}
	defer ole.CoUninitialize()

	if len(os.Args) < 3 {
		printUsage()

		return
	}

	commandName := strings.ToLower(os.Args[1])
	processName := os.Args[2]
	args := os.Args[3:]
	command, ok := commands[commandName]

	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s.", commandName)

		return
	}

	process, err := internal.NewProcess(processName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding process: %s.", err)

		return
	}

	control, err := internal.NewVolumeControl(process)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating volume control: %s.", err)

		return
	}

	command(control, args)
	control.Release()
}

func printUsage() {
	fmt.Println("winmix v1.0.0 ©Severin v. W.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  winmix get <process-name>")
	fmt.Println("  winmix set <process-name> <volume>")
	fmt.Println("  winmix change <process-name> <volume>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  get		Get the volume of a process.")
	fmt.Println("  set		Set the volume of a process [0 to 100].")
	fmt.Println("  change	Change the volume of a process [-100 to 100].")
}

func getVolume(control *internal.VolumeControl, args []string) {
	level, err := control.GetVolume()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting volume: %s.", err)

		return
	}

	fmt.Println(level)
}

func setVolume(control *internal.VolumeControl, args []string) {
	if len(args) < 1 {
		printUsage()

		return
	}

	// Volume ranges from 0 to 100 percent
	volume, ok := parseVolume(args[0], 0, 100)

	if !ok {
		return
	}

	if err := control.SetVolume(volume); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting volume: %s.", err)

		return
	}

	fmt.Println(volume)
}

func changeVolume(control *internal.VolumeControl, args []string) {
	if len(args) < 1 {
		printUsage()

		return
	}

	// Volume change ranges from -100 to 100 percent
	deltaVolume, ok := parseVolume(args[0], -100, 100)

	if !ok {
		return
	}

	volume, err := control.GetVolume()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting volume: %s.", err)

		return
	}

	// Clamp new volume to range 0 - 100
	newVolume := float64(volume + deltaVolume)
	clampedVolume := int(math.Max(0, math.Min(newVolume, 100)))

	if err := control.SetVolume(clampedVolume); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting volume: %s.", err)

		return
	}

	fmt.Println(clampedVolume)
}

func parseVolume(volume string, min int, max int) (int, bool) {
	number, err := strconv.Atoi(volume)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Volume must be a number.")

		return 0, false
	}

	if number < min || number > max {
		fmt.Fprintf(os.Stderr, "Volume must be in the range of %d - %d.", min, max)

		return 0, false
	}

	return number, true
}
