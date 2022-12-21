package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ProxmoxAPI struct {
	User   string
	Pass   string
	Host   string
	Port   string
	Ticket string
	mux    sync.Mutex
}

type ProxmoxAPITicketResp struct {
	Data struct {
		CSRFPreventionToken string `json:"CSRFPreventionToken"`
		Ticket              string `json:"ticket"`
		Username            string `json:"username"`
	} `json:"data"`
}

type ProxmoxAPINodesListResp struct {
	Data []struct {
		Node   string `json:"node"`
		Status string `json:"status"`
	} `json:"data"`
}

type ProxmoxAPIZpoolListResp struct {
	Data []struct {
		Size   float64 `json:"size"`
		Health string  `json:"health"`
		Alloc  float64 `json:"alloc"`
		Free   float64 `json:"free"`
		Name   string  `json:"name"`
		Frag   int     `json:"frag"`
		Dedup  int     `json:"dedup"`
	} `json:"data"`
}

type ProxmoxAPIZpoolResp struct {
	Data struct {
		Action string `json:"action"`
		Scan   string `json:"scan"`
		Leaf   int    `json:"leaf"`
		Errors string `json:"errors"`
		Name   string `json:"name"`
		State  string `json:"state"`
		//Children
	} `json:"data"`
}

// type ZpoolChildren struct {
// 	Write
// 	Read
// 	Cksum
// 	Msg
// 	Leaf  string `json:"content"`
// 	Name  bool   `json:"proxied"`
// 	State bool   `json:"proxied"`
// }

func (api *ProxmoxAPI) getTicket() string {
	api.mux.Lock()
	defer api.mux.Unlock()
	return api.Ticket
}

func (api *ProxmoxAPI) setTicket(ticket string) {
	api.mux.Lock()
	api.Ticket = ticket
	api.mux.Unlock()
}

func (api *ProxmoxAPI) GetAPITicket() (string, error) {
	//Copy the api vars so we can free the lock up
	api.mux.Lock()
	user := api.User
	pass := api.Pass
	host := api.Host
	port := api.Port
	api.mux.Unlock()

	url := "https://" + host + ":" + port + "/api2/json/access/ticket?username=" + user + "&password=" + pass
	c := &tls.Config{
		// TODO: env variable to disable
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: c}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	//req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //Close the resp body when finished

	if resp.StatusCode != 200 {
		return "", errors.New("Expected status 200, got " + strconv.Itoa(resp.StatusCode) + "; " + resp.Status + "; Check your credentials")
	}

	respBody := ProxmoxAPITicketResp{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println(err)
		log.Println(respBody.Data.Ticket)
		log.Println(respBody.Data.Username)
		return respBody.Data.Ticket, err
	}

	return respBody.Data.Ticket, nil
}

func (api *ProxmoxAPI) GetNodes() (ProxmoxAPINodesListResp, error) {
	//Copy the api vars so we can free up the lock
	api.mux.Lock()
	host := api.Host
	port := api.Port
	ticket := api.Ticket
	api.mux.Unlock()

	url := "https://" + host + ":" + port + "/api2/json/nodes"
	c := &tls.Config{
		// TODO: env variable to disable
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: c}
	client := &http.Client{Transport: tr}
	//client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProxmoxAPINodesListResp{}, err
	}
	req.Header.Add("Content-type", "application/json")
	authCookie := http.Cookie{
		Name:  "PVEAuthCookie",
		Value: ticket,
	}
	req.AddCookie(&authCookie)

	resp, err := client.Do(req)
	if err != nil {
		return ProxmoxAPINodesListResp{}, err
	}
	defer resp.Body.Close() //Close the resp body when finished

	respBody := ProxmoxAPINodesListResp{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return ProxmoxAPINodesListResp{}, err
	}

	return respBody, nil
}

func (api *ProxmoxAPI) GetZpoolList(node string) (ProxmoxAPIZpoolListResp, error) {
	//Copy the api vars so we can free up the lock
	api.mux.Lock()
	host := api.Host
	port := api.Port
	ticket := api.Ticket
	api.mux.Unlock()

	url := "https://" + host + ":" + port + "/api2/json/nodes/" + node + "/disks/zfs"
	c := &tls.Config{
		// TODO: env variable to disable
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: c}
	client := &http.Client{Transport: tr}
	//client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProxmoxAPIZpoolListResp{}, err
	}
	req.Header.Add("Content-type", "application/json")
	authCookie := http.Cookie{
		Name:  "PVEAuthCookie",
		Value: ticket,
	}
	req.AddCookie(&authCookie)

	resp, err := client.Do(req)
	if err != nil {
		return ProxmoxAPIZpoolListResp{}, err
	}
	defer resp.Body.Close() //Close the resp body when finished

	respBody := ProxmoxAPIZpoolListResp{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return ProxmoxAPIZpoolListResp{}, err
	}

	return respBody, nil
}

func (api *ProxmoxAPI) GetZpool(node string, name string) (ProxmoxAPIZpoolResp, error) {
	//Copy the api vars so we can free up the lock
	api.mux.Lock()
	host := api.Host
	port := api.Port
	ticket := api.Ticket
	api.mux.Unlock()

	url := "https://" + host + ":" + port + "/api2/json/nodes/" + node + "/disks/zfs/" + name
	c := &tls.Config{
		// TODO: env variable to disable
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: c}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProxmoxAPIZpoolResp{}, err
	}
	req.Header.Add("Content-type", "application/json")
	authCookie := http.Cookie{
		Name:  "PVEAuthCookie",
		Value: ticket,
	}
	req.AddCookie(&authCookie)

	resp, err := client.Do(req)
	if err != nil {
		return ProxmoxAPIZpoolResp{}, err
	}
	defer resp.Body.Close() //Close the resp body when finished

	respBody := ProxmoxAPIZpoolResp{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return ProxmoxAPIZpoolResp{}, err
	}

	return respBody, nil
}

func (api *ProxmoxAPI) refreshTicket() {
	ticker := time.NewTicker(time.Hour)
	for {
		newTicket, err := api.GetAPITicket()
		if err != nil {
			log.Println("Could not retrieve new ticket. Retry on next check... Error:", err)
		} else {
			log.Println("Refreshed ticket")
			api.setTicket(newTicket)
		}
		<-ticker.C
	}
}

func (api *ProxmoxAPI) waitForTicket() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		<-ticker.C
		if api.getTicket() != "" {
			break
		} else {
			log.Println("Waiting to get ticket...")
		}
	}
}
