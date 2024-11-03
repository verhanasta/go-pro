package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseResponse(statsStr string) ([]float64, error) {

	resultValues := make([]float64, 0, 7)

	statsStrItems := strings.Split(statsStr, " ")

	if len(statsStrItems) != 7 {
		return nil, fmt.Errorf("unexpected number of stats")
	}
	for i := 0; i < len(statsStrItems); i++ {
		number, err := strconv.ParseFloat(strings.TrimSpace(statsStrItems[i]), 64)
		if err != nil {
			return nil, err
		}
		resultValues[i] = number
	}

	return resultValues, nil
}

func analyzeStats(statsSlice []float64){

	loadAverage := statsSlice[0]
	totalMemory := statsSlice[1]
	usedMemory := statsSlice[2]
	totalDisk := statsSlice[3]
	usedDisk := statsSlice[4]
	totalBandwidth := statsSlice[5]
	usedBandwidth := statsSlice[6]

	messages := []string{}

	
	if loadAverage > 30 {
		messages = append(messages, fmt.Sprintf("Load Average is too high: %.2f", loadAverage))
	}

	memoryUsagePercent := (usedMemory / totalMemory) * 100
	if memoryUsagePercent > 80 {
		messages = append(messages, fmt.Sprintf("Memory usage is critically high: %.2f%%", memoryUsagePercent))
	}

	diskUsagePercent := (usedDisk / totalDisk) * 100
	if diskUsagePercent > 90 {
		messages = append(messages, fmt.Sprintf("Disk usage is critically high: %.2f%%", diskUsagePercent))
	}

	bandwidthUsagePercent := (usedBandwidth / totalBandwidth) * 100
	if bandwidthUsagePercent > 75 {
		messages = append(messages, fmt.Sprintf("Bandwidth usage is critically high: %.2f%%", bandwidthUsagePercent))
	}

	if len(messages) > 0 {
		fmt.Println(strings.Join(messages, "\n"))
	} else {
		fmt.Println("All systems normal.")
	}

}

func getServerStats(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}


func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"

	errorCount := 0

	for {
		responseStr, err := getServerStats(url)
		if err != nil {
			fmt.Println("Error fetching stats:", err)
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistics")
				break
			}
			time.Sleep(2 * time.Second) 
			continue
		}

		stats, err := parseResponse(responseStr)

		analyzeStats(stats)
		errorCount = 0 
		time.Sleep(60 * time.Second) 
	}
}