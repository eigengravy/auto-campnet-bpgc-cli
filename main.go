package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {

	fmt.Println("Auto CampNet BPGC")
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}
	configFileDir := filepath.Join(configDir, "auto_campnet_bpgc")
	configFile := filepath.Join(configFileDir, "credentials.csv")

	if _, err := os.Stat(configFileDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(configFileDir, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	username, password, err := GetCredentialsFromFile(configFile)
	if err != nil {
		username, password, err = GetCredentialsFromUser(configFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	ticker := time.NewTicker(time.Minute)

	go func() {
		Connect(username, password, configFile)
		for {
			select {
			case <-ticker.C:
				Connect(username, password, configFile)
			}
		}
	}()

	select {}
}

func Connect(username string, password string, configFile string) {
	if _, err := http.Get("https://campnet.bits-goa.ac.in:8090/"); err != nil {
		fmt.Println(formattedTime(), "CampNet unavailable")
	} else if _, err := http.Get("https://google.com/"); err == nil {
		fmt.Println(formattedTime(), "Connected to the internet")
	} else {
		resp, err := http.PostForm("https://campnet.bits-goa.ac.in:8090/login.xml",
			url.Values{"mode": {"191"}, "username": {username}, "password": {password}, "producttype": {"1"}, "a": {strconv.FormatInt(time.Now().UnixMilli(), 10)}})
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		text := string(data)

		if strings.Contains(text, "LIVE") {
			fmt.Println(formattedTime(), "Logged in as", username)
		} else if strings.Contains(text, "failed") {
			err := os.Remove(configFile)
			if err != nil {
				log.Fatalln(err)
			}
			log.Fatalln(formattedTime(), "Incorrect username/password")
		} else if strings.Contains(text, "exceeded") {
			log.Fatalln(formattedTime(), "Data limit exceeded")
		}
	}
}

func GetCredentialsFromFile(configFile string) (string, string, error) {

	f, err := os.Open(configFile)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	creds, err := csv.NewReader(f).Read()
	if err != io.EOF {
		return creds[0], creds[1], nil
	}
	return "", "", err

}

func GetCredentialsFromUser(configFile string) (string, string, error) {

	var username, password string
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password ")
	fmt.Scanln(&password)

	f, err := os.Create(configFile)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{username, password}); err != nil {
		return "", "", err
	}
	return username, password, nil
}

func formattedTime() string {
	return "[ " + time.Now().Local().Format(time.RFC822) + " ]"
}
