package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"flag"
)

func save(httpClient http.Client, url string, dir string) bool{
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return false;
	}

	req.Header.Set("User-Agent", "client")
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Println(getErr)
		return false;
	}
	
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Println(readErr)
		return false;
	}
	
	dt := time.Now()
	ioutil.WriteFile(dir + "/response-" + dt.Format("2006-01-02_15-04-05"), body, 0644)
	return true
}

func main() {
	intervalPtr := flag.Int("t", 10, "Request interval")
	flag.Parse()

	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	
	_ = os.Mkdir("data_gortrans", 0666)
	_ = os.Mkdir("data_permkrai", 0666)
	
	for true {
		if (save(httpClient, "http://map.gortransperm.ru/json/uvb/getVehiclesInDistrict", "data_gortrans")) {
			fmt.Println(time.Now().Format("15:04:05 01.02.2006") + ": данные gortransperm.ru прочитаны")
		}
		
		if (save(httpClient, "https://head.permkrai.ru/api/transport/monitoring/keep_map/current", "data_permkrai")) {
			fmt.Println(time.Now().Format("15:04:05 01.02.2006") + ": данные permkrai.ru прочитаны")
		}

		time.Sleep(time.Duration(*intervalPtr) * time.Second)
	}
}