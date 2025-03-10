package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL       = "http://srv.msk01.gigacorp.local/_stats"
	maxRetryCount   = 3                      // Максимальное количество повторов при ошибках
	httpTimeout     = 5 * time.Second        // Таймаут для HTTP-запроса
	requestInterval = 500 * time.Millisecond // Интервал между запросами

	expectedMetricsLength = 7 // Ожидаемое количество метрик в ответе от сервера

	cpuLoadThreshold      = 30 // Порог для нагрузки CPU
	memoryUsageThreshold  = 80 // Порог для использования памяти
	diskUsageThreshold    = 90 // Порог для использования дискового пространства
	networkUsageThreshold = 90 // Порог для использования пропускной способности сети

	bytesInMegabyte = 1024 * 1024 // Количество байтов в одном мегабайте
	bytesInMegabit  = 1000 * 1000 // Количество байтов в одном мегабите
	fullPercent     = 100         // Используется для расчета процентного использования ресурсов
)

type Metric struct {
	capacity   int
	usage      int
	threshold  int
	message    string
	unit       string
	checkUsage func(capacity, usage int) (int, int)
}

func main() {
	resultStream := initiatePolling(serverURL, maxRetryCount)

	for response := range resultStream() {
		metrics, err := parseMetrics(response)
		if err != nil {
			continue
		}

		metricList := []Metric{
			{
				capacity:   metrics.CPULoad,
				usage:      metrics.CPULoad,
				threshold:  cpuLoadThreshold,
				message:    "Load Average is too high: %d\n",
				unit:       "",
				checkUsage: calculateDirectUsage,
			},
			{
				capacity:   metrics.MemoryCapacity,
				usage:      metrics.MemoryUsage,
				threshold:  memoryUsageThreshold,
				message:    "Memory usage too high: %d%%\n",
				unit:       "%",
				checkUsage: calculatePercentageUsage,
			},
			{
				capacity:   metrics.DiskCapacity,
				usage:      metrics.DiskUsage,
				threshold:  diskUsageThreshold,
				message:    "Free disk space is too low: %d Mb left\n",
				unit:       "Mb",
				checkUsage: calculateFreeResource,
			},
			{
				capacity:   metrics.NetworkCapacity,
				usage:      metrics.NetworkActivity,
				threshold:  networkUsageThreshold,
				message:    "Network bandwidth usage high: %d Mbit/s available\n",
				unit:       "Mbit/s",
				checkUsage: calculateFreeNetworkResource,
			},
		}

		for _, metric := range metricList {
			checkResourceUsage(metric)
		}
	}
}

func checkResourceUsage(m Metric) {
	usagePercent, freeResource := m.checkUsage(m.capacity, m.usage)

	if usagePercent > m.threshold {
		if m.unit == "%" || m.unit == "" {
			fmt.Printf(m.message, usagePercent)
		} else {
			fmt.Printf(m.message, freeResource)
		}
	}
}

func calculateDirectUsage(capacity, _ int) (int, int) {
	usagePercent := capacity
	return usagePercent, usagePercent
}

func calculatePercentageUsage(capacity, usage int) (int, int) {
	usagePercent := usage * fullPercent / capacity
	return usagePercent, usagePercent
}

func calculateFreeResource(capacity, usage int) (int, int) {
	usagePercent := usage * fullPercent / capacity
	freeResource := (capacity - usage) / bytesInMegabyte
	return usagePercent, freeResource
}

func calculateFreeNetworkResource(capacity, usage int) (int, int) {
	usagePercent := usage * fullPercent / capacity
	freeResource := (capacity - usage) / bytesInMegabit
	return usagePercent, freeResource
}

func initiatePolling(url string, retries int) func() chan string {
	return func() chan string {
		dataChannel := make(chan string, 3)
		client := http.Client{Timeout: httpTimeout}
		errorCounter := 0

		go func() {
			defer close(dataChannel)

			for {
				time.Sleep(requestInterval)

				if errorCounter >= retries {
					fmt.Println("Unable to fetch server statistics")
					break
				}

				response, err := client.Get(url)
				errorCounter = handleResponseError(response, err, errorCounter)
				if errorCounter > 0 {
					continue
				}

				body, err := io.ReadAll(response.Body)
				if err != nil {
					errorCounter = handlePollingError(err, errorCounter, "failed to parse response")
					continue
				}

				response.Body.Close()
				dataChannel <- string(body)

				errorCounter = 0
			}
		}()

		return dataChannel
	}
}

func handleResponseError(response *http.Response, err error, errorCounter int) int {
	if err != nil {
		return handlePollingError(err, errorCounter, "failed to send request")
	}
	if response.StatusCode != http.StatusOK {
		return handlePollingError(fmt.Errorf("invalid status code: %d", response.StatusCode), errorCounter, "")
	}
	return errorCounter
}

func handlePollingError(err error, errorCounter int, message string) int {
	if message != "" {
		fmt.Printf("%s: %s\n", message, err)
	}
	return errorCounter + 1
}

type ServerMetrics struct {
	CPULoad         int
	MemoryCapacity  int
	MemoryUsage     int
	DiskCapacity    int
	DiskUsage       int
	NetworkCapacity int
	NetworkActivity int
}

func parseMetrics(data string) (ServerMetrics, error) {
	parts := strings.Split(data, ",")
	if len(parts) != expectedMetricsLength {
		return ServerMetrics{}, fmt.Errorf("invalid data format")
	}

	values := make([]int, expectedMetricsLength)
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return ServerMetrics{}, fmt.Errorf("invalid number: %s", part)
		}
		values[i] = value
	}

	return ServerMetrics{
		CPULoad:         values[0],
		MemoryCapacity:  values[1],
		MemoryUsage:     values[2],
		DiskCapacity:    values[3],
		DiskUsage:       values[4],
		NetworkCapacity: values[5],
		NetworkActivity: values[6],
	}, nil
}