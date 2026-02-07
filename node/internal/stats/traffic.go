package stats

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type TrafficSample struct {
	Interface string    `json:"interface"`
	RxBytes   uint64    `json:"rx_bytes"`
	TxBytes   uint64    `json:"tx_bytes"`
	At        time.Time `json:"at"`
}

// DetectDefaultInterface tries to find the default route interface by parsing /proc/net/route.
// This works in Docker with host networking, and avoids pulling in netlink deps.
func DetectDefaultInterface() (string, error) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	// Skip header.
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return "", err
		}
		return "", errors.New("empty /proc/net/route")
	}
	for s.Scan() {
		// Iface  Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
		fields := strings.Fields(s.Text())
		if len(fields) < 4 {
			continue
		}
		iface := fields[0]
		dest := fields[1]
		flagsHex := fields[3]
		if dest != "00000000" {
			continue
		}
		flags, err := strconv.ParseUint(flagsHex, 16, 64)
		if err != nil {
			continue
		}
		// RTF_UP (0x1)
		if flags&0x1 == 0 {
			continue
		}
		return iface, nil
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", errors.New("default route not found")
}

// ReadNetDev reads /proc/net/dev and returns rx/tx bytes for the given interface.
func ReadNetDev(iface string) (rx uint64, tx uint64, err error) {
	iface = strings.TrimSpace(iface)
	if iface == "" {
		return 0, 0, errors.New("missing interface")
	}
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	// Skip 2 header lines.
	for i := 0; i < 2; i++ {
		if !s.Scan() {
			if err := s.Err(); err != nil {
				return 0, 0, err
			}
			return 0, 0, errors.New("unexpected /proc/net/dev format")
		}
	}
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name != iface {
			continue
		}
		cols := strings.Fields(strings.TrimSpace(parts[1]))
		// https://man7.org/linux/man-pages/man5/proc.5.html
		// Receive: bytes packets errs drop fifo frame compressed multicast
		// Transmit: bytes packets errs drop fifo colls carrier compressed
		if len(cols) < 16 {
			return 0, 0, fmt.Errorf("unexpected columns for %s in /proc/net/dev", iface)
		}
		rx, err = strconv.ParseUint(cols[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid rx bytes for %s: %w", iface, err)
		}
		tx, err = strconv.ParseUint(cols[8], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid tx bytes for %s: %w", iface, err)
		}
		return rx, tx, nil
	}
	if err := s.Err(); err != nil {
		return 0, 0, err
	}
	return 0, 0, fmt.Errorf("interface %s not found in /proc/net/dev", iface)
}

func CurrentSample(iface string) (TrafficSample, error) {
	if strings.TrimSpace(iface) == "" {
		detected, err := DetectDefaultInterface()
		if err != nil {
			return TrafficSample{}, err
		}
		iface = detected
	}
	rx, tx, err := ReadNetDev(iface)
	if err != nil {
		return TrafficSample{}, err
	}
	return TrafficSample{
		Interface: iface,
		RxBytes:   rx,
		TxBytes:   tx,
		At:        time.Now().UTC(),
	}, nil
}
