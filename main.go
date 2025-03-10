package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Конфигурация
const (
	serverURL          = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval       = 60 * time.Second // Интервал опроса
	errorThreshold     = 3
	loadAverageThresh  = 30
	memoryUsageThresh  = 0.8
	diskSpaceThresh    = 0.9
	networkUsageThresh = 0.9
)

var errorCount int

func fetchServerStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		errorCount++
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorCount++
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorCount++
		return nil, err
	}

	// Разделяем данные по запятым
	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 6 {
		errorCount++
		return nil, fmt.Errorf("invalid data format: expected 6 values, got %d", len(parts))
	}

	// Преобразуем строки в числа
	stats := make([]float64, 6)
	for i, part := range parts {
		stats[i], err = strconv.ParseFloat(part, 64)
		if err != nil {
			errorCount++
			return nil, fmt.Errorf("invalid data format: %v", err)
		}
	}

	errorCount = 0
	return stats, nil
}

func checkThresholds(stats []float64) {
	loadAvg := stats[0]
	totalMem := stats[1]
	usedMem := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	totalNet := stats[5]
	usedNet := stats[6]

	// Проверка Load Average
	if loadAvg > loadAverageThresh {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}

	// Проверка использования памяти
	memoryUsage := usedMem / totalMem
	if memoryUsage > memoryUsageThresh {
		fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsage*100)
	}

	// Проверка свободного места на диске
	diskUsage := usedDisk / totalDisk
	if diskUsage > diskSpaceThresh {
		freeSpaceMB := (totalDisk - usedDisk) / (1024 * 1024)
		fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeSpaceMB)
	}

	// Проверка загруженности сети
	netUsage := usedNet / totalNet
	if netUsage > networkUsageThresh {
		freeBandwidthMbit := (totalNet - usedNet) / (1024 * 1024)
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeBandwidthMbit)
	}
}

func main() {
	for {
		stats, err := fetchServerStats()
		if err != nil {
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistic")
			}
		} else {
			checkThresholds(stats)
		}

		time.Sleep(pollInterval)
	}
}
