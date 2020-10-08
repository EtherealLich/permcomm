package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"flag"
	"encoding/json"
)

type Vehicle struct {
	VehicleId int
	LicencePlate string
	VehicleTypeId int
	VehicleTypeName string
	Course int
	T string
	Location string
}

type VehicleType struct {
	VehicleTypeId int
	NumVehicles int
}

type dataResponse struct {
    VehicleTypes []VehicleType `json:"vehicleTypes"`
    Vehicles []Vehicle `json:"vehicles"`
}

func getCsvLine(vehicle Vehicle) string {
	lng := .0
	lat := .0
	fmt.Sscanf(vehicle.Location, "POINT (%f %f)", &lng, &lat)
	return time.Now().Format("2006-01-02T15:04:05") + 
		";" + "\"" + vehicle.LicencePlate + "\"" +
		";" + fmt.Sprintf("%v", vehicle.VehicleTypeId) +
		";" + "\"" + vehicle.VehicleTypeName + "\"" +
		";" + "\"" + vehicle.T + "\"" +
		";" + "\"" + fmt.Sprintf("%v", lng) + "\"" +
		";" + "\"" + fmt.Sprintf("%v", lat) + "\"" +
		"\n"
		
}

func save(httpClient http.Client, url string, f *os.File) bool {
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
	
	jsonRes := dataResponse{}
	json.Unmarshal(body, &jsonRes)

	for _, vehicle := range jsonRes.Vehicles {
		f.WriteString(getCsvLine(vehicle))
	}
	log.Println("Данные прочитаны. Транспортных средств получено :" + fmt.Sprintf("%v", len(jsonRes.Vehicles)))
	
	return true
}

func main() {
	intervalPtr := flag.Int("t", 10, "Request interval")
	flag.Parse()

	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	
	path := "data_gortrans";
	filename := "response-data-" + time.Now().Format("2006-01-02") + ".csv"
	
	_ = os.Mkdir(path, 0666)
	f, _ := os.OpenFile(path + "/" + filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()
	
	// Периодически закрываем и открываем файл
	closeFileTimer := time.NewTicker(time.Second * 300)
	go func() {
        for range closeFileTimer.C {
			f.Close()
			f, _ = os.OpenFile(path + "/" + filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
            // log.Println("Reopened file at", t)
        }
    }()
	
	for true {
		if (filename != "response-data-" + time.Now().Format("2006-01-02") + ".csv") {
			f.Close()
			filename = "response-data-" + time.Now().Format("2006-01-02") + ".csv"
			f, _ = os.OpenFile(path + "/" + filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		}
		save(httpClient, "http://map.gortransperm.ru/json/uvb/getVehiclesInDistrict", f)
		time.Sleep(time.Duration(*intervalPtr) * time.Second)
	}
}