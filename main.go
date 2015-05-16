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
	"fmt"
	"os"
	"regexp"
	
	"encoding/json"
	"io/ioutil"
	"net/http"
	
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
	}
	
	app.Action = func (c *cli.Context)  {
		cli.ShowAppHelp (c)
	}
	
	app.Run (os.Args)
}

type testResponse struct {
    Status string
}


func testAction (c *cli.Context) {
	re := regexp.MustCompile("(https?://[\\w.]+)/?")
	host_match := re.FindStringSubmatch (c.GlobalString ("host"))
	
	if len (host_match) == 0 || host_match[1] == "" {
		fmt.Println ("Please specify a host.")
		return
	}
	
	host := host_match[1]
	
	fmt.Println ("Testing API of instance: ", host)

	resp, err := http.Get (host + "/api.php?action=test")
	
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
