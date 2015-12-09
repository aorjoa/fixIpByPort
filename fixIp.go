package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"net/http/httputil"
	"net/http/cookiejar"
	"strings"
	"strconv"
)
var ipRange = "192.168.5"
var linkPort = []string{"45","46","47","48","49","50","51","52"}
type ResponseAuth struct {
		Redirect	string `json:"redirect"`
		Error		string `json:"error"`
	}

var ipAndMacMapping = map[string]string{}

func saveDhcpConf(){
	var header = `
############ SETTING ############
######### set dhcp range ########

default-lease-time 60000;
max-lease-time 72000;
subnet 192.168.5.0 netmask 255.255.255.0 {
	range 192.168.5.11 192.168.5.200;
	option subnet-mask 255.255.255.0;
	option broadcast-address 192.168.5.255;
	option routers 192.168.5.9;
	option domain-name-servers 8.8.8.8, 8.8.4.4;
	option domain-name "aiyara.lab.sut.ac.th";
} 

######### reserved ip  ########`
var body = ""
for ip,mac := range ipAndMacMapping {
	body = fmt.Sprintf("%s\nport-%s { hardware ethernet %s; fixed-address %s.%s; }", body, ip, mac, ipRange, ip)
}
err := ioutil.WriteFile("./dhcpd.conf", []byte(header+body), 0644)
if err != nil {
	fmt.Printf("error writefile: %s",err)
}
fmt.Println(header + body)
}

func main() {
	//get data from SW1,SW2
	getData(".253","SW1");
	getData(".254","SW2");

	//build configuration file
	saveDhcpConf()
}

func getData(ipAddr string,swNo string) {
	var err error
	credential := map[string]string{
		"username": "admin",
		"password": "",
	}
	// Create cookie.
  	cookieJar, _ := cookiejar.New(nil)

	contentReader := bytes.NewReader([]byte("username="+credential["username"]+"&password="+credential["password"]))
	req, err := http.NewRequest("POST", "http://"+ipRange+ipAddr+"/htdocs/login/login.lua", contentReader)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("==== REQ ====\n%s\n=============\n",dump)
	client := &http.Client{
	    Jar: cookieJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error Request: %s", err)
	}
	defer resp.Body.Close()

	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error ReadAll: %s", err)
	}

	var responseAuth ResponseAuth
	err = json.Unmarshal(body, &responseAuth)
	if err != nil {
		fmt.Printf("error Unmarshal: %s", err)
	}
	fmt.Printf("Redirect: %v\n", responseAuth.Redirect)
	fmt.Printf("Error: %v\n", responseAuth.Error)
	fmt.Printf("Status: %v\n", resp.Status)


	req, err = http.NewRequest("GET", "http://"+ipRange+ipAddr+"/htdocs/pages/base/mac_address_table.lsp", nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Cache-Control", "no-cache")
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("error Request: %s", err)
	}
	body,err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error ReadAll: %s", err)
	}
	bodyByLine := strings.Split(string(body),"\n")
	
	for _,line := range bodyByLine {
		if strings.HasPrefix(line, "['") && strings.HasSuffix(line, "']") {
			lineAttr := strings.Split(string(line),"', '")

			if _,err = strconv.Atoi(lineAttr[2]); err == nil {
				if contains(linkPort,lineAttr[2]) {
					break;
				}
				var ipIntLogical,err = strconv.Atoi(lineAttr[2])
				if err != nil {
					panic("ERR : Cannot convert ip to integer")
				}
				if(swNo == "SW2"){
					ipIntLogical = ipIntLogical + 44
				}

				var ipLogical = strconv.Itoa(ipIntLogical)
				fmt.Println(ipLogical + " " + lineAttr[1])
				ipAndMacMapping[ipLogical] = lineAttr[1]
			}
		}
	}

	// Logout to clear session (because it's limited).
	req, err = http.NewRequest("GET", "http://"+ipRange+ipAddr+"/htdocs/pages/main/logout.lsp", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("error Request: %s", err)
	}
}

func contains(slice []string, element string) bool {
    for _, item := range slice {
        if item == element {
            return true
        }
    }
    return false
}