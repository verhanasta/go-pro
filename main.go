package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	//"time"
)

func parseReq(statsStr string) ([]float64, error) {

	resultValues := make([]float64, 0, 7)

	statsStrItems := strings.Split(statsStr, " ")

	if len(statsStrItems) != 7 {
		return nil, fmt.Errorf("unexpected number of stats")
	}
	for i := 0; i < len(statsStrItems); i++ {
		resultValues[i], err = strconv.ParseFloat(statsStrItems[i], 64)
		if err != nil {
			return nil, err
		}
	}
	return resultValues, nil
}

func analizeStats(statsSlice []float64){

	loadAverage := statsSlice[0]
	totalMemory := statsSlice[1]
	usedMemory := statsSlice[2]
	totalDisk := statsSlice[3]
	usedDisk := statsSlice[4]
	totalBandwidth := statsSlice[5]
	usedBandwidth := statsSlice[6]

	if loadAverage > 30 {
		fmt.Println("Load Average is too high: ", 30)
	}
	if usedMemory > (totalMemory * 0.8) {
		fmt.Println("Memory usage too high: ", usedMemory  / totalMemory * 100, "%")
	}
	if usedDisk  > (totalDisk * 0.9) {
		fmt.Println("Memory usage too high: ", usedDisk / totalDisk * 100, "%")
	}

	if usedBandwidth  > (totalBandwidth * 0.9) {
		fmt.Println("Memory usage too high: ", usedBandwidth / totalBandwidth * 100, "%")
	}
}

func fetchServerStats(url string, headers map[string]string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
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


func main(){



}