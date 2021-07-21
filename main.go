package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Result struct {
	ipv4 string
	ipv6 string
}

// Location contains all the relevant data for an IP
type Location struct {
	As          string  `json:"as"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Isp         string  `json:"isp"`
	Lat         float32 `json:"lat"`
	Lon         float32 `json:"lon"`
	Org         string  `json:"org"`
	Query       string  `json:"query"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	Status      string  `json:"status"`
	Timezone    string  `json:"timezone"`
	Zip         string  `json:"zip"`
}

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}

func Index(w http.ResponseWriter, r *http.Request) {
	var answer string

	ip, err := getIP(r)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("No valid ip"))
		return
	}

	// Remove port
	ip = strings.Split(ip, ":")[0]

	response, err := http.Get("http://ip-api.com/json/" + ip)

	if err != nil {
		fmt.Print(err.Error())
	}

	defer response.Body.Close()

	if len(ip) > 0 {
		answer = ip

		responseData, err := ioutil.ReadAll(response.Body)
		if err == nil {
			var res Location
			json.Unmarshal([]byte(responseData), &res)

			if res.Status == "success" {
				answer = answer + "\n" + res.City + ", " + res.Country
				answer = answer + "\n" + res.Timezone
				answer = answer + "\n" + res.Isp
			}
		}
	} else {
		answer = "Impossible to get IP. Please try later"
	}
	answer = answer + "\n"
	answer = answer + "\n" + "Created by Quentin Laffont. https://qlaffont.com"
	w.WriteHeader(200)
	w.Write([]byte(answer))
}

func main() {
	var port = os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}

	fmt.Println("Listening on :" + port)

	http.HandleFunc("/", Index)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
