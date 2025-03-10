package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Конфигурация
const (
	serverURL         = "http://srv.msk01.gigacorp.local/_stats" // Используем правильный URL
	pollInterval      = 60 * time.Second                        // Интервал опроса
	errorThreshold    = 3                                       // Количество ошибок перед выводом сообщения о недоступности данных
	loadAverageThresh = 30                                      // Порог для Load Average
	memoryUsageThresh = 0.8                                     // Порог для использования памяти (80%)
	diskSpaceThresh   = 0.9                                     // Порог для использования диска (90%)
	networkUsageThresh = 0.9                                    // Порог для использования сети (90%)
)

var errorCount int

func fetchServerStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		errorCount++
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorCount++
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	// Чтение тела ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorCount++
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Разделяем данные по запятым
	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 { // Ожидаем 7 значений
		errorCount++
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}

	// Преобразуем строки в числа
	stats := make([]float64, 7)
	for i, part := range parts {
		stats[i], err = strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			errorCount++
			return nil, fmt.Errorf("invalid data format: %v", err)
		}
	}

	errorCount = 0 // Сброс счетчика ошибок при успешном запросе
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
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// Проверка использования памяти
	if totalMem > 0 { // Избегаем деления на ноль
		memoryUsage := usedMem / totalMem
		if memoryUsage > memoryUsageThresh {
			fmt.Printf("Memory usage too high: %.0f%%\n", memoryUsage*100)
		}
	}

	// Проверка свободного места на диске
	if totalDisk > 0 { // Избегаем деления на ноль
		diskUsage := usedDisk / totalDisk
		if diskUsage > diskSpaceThresh {
			freeSpaceMB := (totalDisk - usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeSpaceMB)
		}
	}

	// Проверка загруженности сети
	if totalNet > 0 { // Избегаем деления на ноль
		netUsage := usedNet / totalNet
		if netUsage > networkUsageThresh {
			freeBandwidthMbit := (totalNet - usedNet) / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeBandwidthMbit)
		}
	}
}

func main() {
	for {
		stats, err := fetchServerStats()
		if err != nil {
			fmt.Printf("Error: %v\n", err) // Логируем ошибку для отладки
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistic")
			}
		} else {
			checkThresholds(stats)
		}

		time.Sleep(pollInterval)
	}
}
