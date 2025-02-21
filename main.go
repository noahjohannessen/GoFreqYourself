package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate    = 44100 // CD-quality sample rate
	channels      = 1     // Mono recording
	bitsPerSample = 16    // 16-bit audio
	seconds       = 5     // Duration to record
)

func main() {
	// Initialize PortAudio
	err := portaudio.Initialize()
	if err != nil {
		log.Fatalf("PortAudio initialization failed: %v", err)
	}
	defer portaudio.Terminate()

	// List available audio devices
	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatalf("Failed to get audio devices: %v", err)
	}

	fmt.Println("\nAvailable Audio Devices:")
	for i, device := range devices {
		if device.MaxInputChannels > 0 { // List only input devices
			fmt.Printf("[%d] %s - Max Input Channels: %d\n", i, device.Name, device.MaxInputChannels)
		}
	}

	// Ask user to select a microphone
	var deviceID int
	fmt.Print("\nEnter the ID of the microphone to use: ")
	fmt.Scanln(&deviceID)

	// Validate selection
	if deviceID < 0 || deviceID >= len(devices) || devices[deviceID].MaxInputChannels == 0 {
		log.Fatalf("Invalid device selection.")
	}

	// Get the selected device
	selectedDevice := devices[deviceID]

	fmt.Printf("Using device: %s\n", selectedDevice.Name)

	// Buffer to store audio data
	buffer := make([]int16, 0, sampleRate*seconds) // Use a dynamic slice

	// Open input stream with selected microphone
	stream, err := portaudio.OpenStream(portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   selectedDevice,
			Channels: channels,
			Latency:  selectedDevice.DefaultLowInputLatency,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: 1024, // Small buffer for real-time recording
	}, func(in []int16) {
		buffer = append(buffer, in...) // Append new samples dynamically
	})

	if err != nil {
		log.Fatalf("Failed to open audio stream: %v", err)
	}
	defer stream.Close()

	// Start recording
	fmt.Println("\nRecording... Speak into the microphone.")
	err = stream.Start()
	if err != nil {
		log.Fatalf("Failed to start audio stream: %v", err)
	}

	// Wait for the recording duration
	time.Sleep(time.Second * time.Duration(seconds))

	// Stop the stream
	err = stream.Stop()
	if err != nil {
		log.Fatalf("Failed to stop audio stream: %v", err)
	}

	// Save the recorded data as a WAV file
	err = saveWAV("output.wav", buffer)
	if err != nil {
		log.Fatalf("Failed to save WAV file: %v", err)
	}

	fmt.Println("\nRecording saved as output.wav!")
}

// saveWAV writes recorded audio to a WAV file
func saveWAV(filename string, buffer []int16) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write WAV header
	file.WriteString("RIFF")                                          // Chunk ID
	binary.Write(file, binary.LittleEndian, uint32(36+len(buffer)*2)) // Chunk size
	file.WriteString("WAVE")                                          // Format
	file.WriteString("fmt ")                                          // Subchunk1 ID
	binary.Write(file, binary.LittleEndian, uint32(16))               // Subchunk1 size
	binary.Write(file, binary.LittleEndian, uint16(1))                // Audio format (PCM)
	binary.Write(file, binary.LittleEndian, uint16(channels))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*channels*bitsPerSample/8))
	binary.Write(file, binary.LittleEndian, uint16(channels*bitsPerSample/8))
	binary.Write(file, binary.LittleEndian, uint16(bitsPerSample))

	// Data chunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(len(buffer)*2)) // Data size

	// Write PCM audio data
	for _, sample := range buffer {
		binary.Write(file, binary.LittleEndian, sample)
	}

	return nil
}
