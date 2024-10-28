package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	dir := "/Users/ghost/Downloads" // specify the directory you want to search
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ts") {
			convertToMP4(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %v\n", err)
	}
}

func convertToMP4(tsFile string) {
	mp4File := strings.TrimSuffix(tsFile, ".ts") + ".mp4"
	cmd := exec.Command("ffmpeg", "-i", tsFile, "-c:v", "libx264", "-c:a", "aac", "-strict", "experimental", "-vf", "scale=iw:ih", mp4File)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error converting file %s: %v\n", tsFile, err)
	} else {
		fmt.Printf("Successfully converted %s to %s\n", tsFile, mp4File)
		// Remove the .ts file after successful conversion
		err = os.Remove(tsFile)
		if err != nil {
			fmt.Printf("Error removing file %s: %v\n", tsFile, err)
		} else {
			fmt.Printf("Successfully removed %s\n", tsFile)
		}
	}
}
