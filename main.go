/*
	TemPIC-cli - Copyright (c) PotcFdk, 2015

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at
	
	http://www.apache.org/licenses/LICENSE-2.0
	
	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	
	"github.com/codegangsta/cli"
)

func main () {
	app := cli.NewApp ()
	app.Name    = "tempic-cli"
	app.Usage   = "TemPIC command line toolkit"
	app.Version = "0.0.1a"
	app.Author  = "PotcFdk"
	app.Email   = "PotcFdk@gmx.net"
	
	app.Flags = [] cli.Flag {
		cli.StringFlag {
			Name:   "host, url, u",
			Usage:  "Base URL of the TemPIC instance to use. " +
			        "(e.g. http://tempic.example.com:1234)",
			EnvVar: "TEMPIC_HOST",
		},
	}
	
	app.Commands = [] cli.Command {
		{
			Name:      "test",
			ShortName: "t",
			Usage:     "Checks if the host is reachable and working",
			Action:    testAction,
		},
		{
			Name:      "info",
			ShortName: "i",
			Usage:     "Displays album information",
			Action:    infoAction,
			Flags: []cli.Flag {
				cli.StringFlag{
					Name:  "album-id",
					Usage: "Album ID",
				},
			},
		},
		{
			Name:      "upload",
			ShortName: "up",
			Usage:     "Uploads one or multiple file(s) and creates an album",
			Action:    uploadAction,
			Flags: []cli.Flag {
				cli.StringFlag{
					Name:  "lifetime",
					Value: "default",
					Usage: "Album lifetime",
				},
				cli.StringFlag{
					Name:  "title",
					Value: "",
					Usage: "Album title",
				},
				cli.StringFlag{
					Name:  "description, desc",
					Value: "",
					Usage: "Album description",
				},
			},
		},
	}
	
	app.Action = func (c *cli.Context)  {
		cli.ShowAppHelp (c)
	}
	
	app.Run (os.Args)
}

func cleanUrl (url string) (bool, string) {
	re := regexp.MustCompile ("(https?://[\\w.]+)/?")
	url_match := re.FindStringSubmatch (url)
	
	if len (url_match) == 0 || url_match[1] == "" {
		return false, "empty url"
	}
	url = url_match[1]
	
	return true, url
}

// Test

type testResponse struct {
	Status string
}

func testAction (c *cli.Context) {
	host_ok, host := cleanUrl (c.GlobalString ("host"))
	if !host_ok {
		fmt.Println ("Please specify a host.")
		return
	}
	
	fmt.Println ("Testing API of instance: ", host)

	resp, err := http.Get (host + "/api.php?v1/system/test")
	
	if err != nil {
		fmt.Println ("Error in HTTP request: ", err)
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll (resp.Body)
	
	fmt.Println ("Raw response: ", string (body))
	
	respStruct := &testResponse{}
	
	json.Unmarshal (body, &respStruct)
	
	switch respStruct.Status {
		case "success":
			fmt.Println ("OK")
		default:
			fmt.Println ("ERROR")
	}
}

// Info

type fileInfo struct {
	Url string
	Extension string
	IsImage bool
	Thumbnail string
	Checksums map[string]string
}

type albumInfo struct {
	Name string
	Description string
	Files map[string]fileInfo
}

type infoResponseData struct {
	Albums map[string]albumInfo
}

type infoResponse struct {
	Status string
	Version string
	Data infoResponseData
}

func infoAction (c *cli.Context) {
	host_ok, host := cleanUrl (c.GlobalString ("host"))
	if !host_ok {
		fmt.Println ("Please specify a host.")
		return
	}
	
	album_id := c.String ("album-id")
	
	fmt.Println ("album_id = " + c.String ("album-id"))
	resp, err := http.Get (host + "/api.php?v1/albums/" + album_id + "/info")
	
	if err != nil {
		fmt.Println ("Error in HTTP request: ", err)
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll (resp.Body)
	
	//fmt.Println ("Raw response: ", string (body))
	
	respStruct := &infoResponse{}
	
	json.Unmarshal (body, &respStruct)
	
	switch respStruct.Status {
		case "success":
			fmt.Println ("ANSWER OK")
		default:
			fmt.Println ("ERROR")
			return
	}
	
	var album_info albumInfo = respStruct.Data.Albums[album_id]
	
	fmt.Println ("* Album name: " + album_info.Name)
	fmt.Println ("* Album description: " + album_info.Description)
	fmt.Println ("* Files in album:")
	
	for name, file_info := range album_info.Files {
		fmt.Println ("- " + name + ": ")
		fmt.Println ("  URL: " + file_info.Url)
		fmt.Println ("  SHA-1: " + file_info.Checksums["sha1"])
	}
}

// Upload

func uploadAction (c *cli.Context) {
	// Get the host
	
	host_ok, host := cleanUrl (c.GlobalString ("host"))
	if !host_ok {
		fmt.Println ("Please specify a host.")
		return
	}
	
	// Print information.
	
	fmt.Println ("album_name = " + c.String ("title"))
	fmt.Println ("album_description = " + c.String ("description"))
	fmt.Println ("lifetime = " + c.String ("lifetime"))
	
	fmt.Printf ("uploading files: ")
	for _, arg := range c.Args() {
		fmt.Printf (" " + arg)
	}
	fmt.Println ("\n--")

	// Set the URI.
	
	uri := host + "/upload.php"
	
	// And go!
	
	params := map[string]string {
		"ajax": "true",
		"lifetime": c.String ("lifetime"),
		"album_name": c.String ("title"),
		"album_description": c.String ("description"),
	}
	
	path := c.Args()[0]
	
	file, err := os.Open (path)
	if err != nil {
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter (body)
	part, err := writer.CreateFormFile ("file[]", filepath.Base (path))
	if err != nil {
		return
	}
	_, err = io.Copy (part, file)
 
	for key, val := range params {
		_ = writer.WriteField (key, val)
	}
	err = writer.Close()
	if err != nil {
		return
	}
	
	request, err := http.NewRequest ("POST", uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	
	fmt.Println (request)
	
	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
 
		fmt.Println(body)
	}
}
