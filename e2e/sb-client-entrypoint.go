package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	defaultConfigPath = "/etc/sing-box/config.json"
	pollInterval      = 500 * time.Millisecond
)

func main() {
	configPath := os.Getenv("SB_CLIENT_CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	log.Printf("waiting for sing-box config: %s", configPath)
	for {
		ready, err := configReady(configPath)
		if err != nil {
			log.Fatalf("check config failed: %v", err)
		}
		if ready {
			break
		}
		time.Sleep(pollInterval)
	}

	bin, err := exec.LookPath("sing-box")
	if err != nil {
		log.Fatalf("find sing-box binary failed: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"run", "-D", "/etc/sing-box", "-C", "/etc/sing-box"}
	}

	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	log.Printf("starting sing-box: %s %v", bin, args)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("run sing-box failed: %v", err)
	}
}

func configReady(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
	return info.Size() > 0, nil
}
