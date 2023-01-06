package netbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client

	token string
}

func NewClient(opts ...Option) (*Client, error) {
	x := Client{
		baseURL: &url.URL{
			Scheme: "http",
			Host:   "localhost",
		},
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
	for _, o := range opts {
		if err := o(&x); err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (nb *Client) ListDevices(site string) ([]Device, error) {
	queryURL := *nb.baseURL
	queryURL.Path = "/api/dcim/devices/"

	queryVals := url.Values{}
	queryVals.Add("tag", "pxe-enable")
	queryVals.Add("has_primary_ip", "yes")

	if site != "" {
		queryVals.Add("site", site)
	}
	queryURL.RawQuery = queryVals.Encode()

	queryHeaders := http.Header{}
	queryHeaders.Add("accept", "application/json")
	queryHeaders.Add("authorization", "token "+nb.token)

	req := &http.Request{
		URL:    &queryURL,
		Header: queryHeaders,
	}

	morepages := true
	devices := []Device{}
	for morepages {
		resp, err := nb.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)
		data := struct {
			Count    int
			Next     string
			Previous string
			Results  []Device
		}{}

		if err := dec.Decode(&data); err != nil {
			return nil, err
		}
		devices = append(devices, data.Results...)

		if data.Next != "" {
			req.URL, _ = url.Parse(data.Next)
		} else {
			morepages = false
		}
	}
	return devices, nil
}

func (nb *Client) ListInterfaces(deviceID int64) ([]Interface, error) {
	queryURL := *nb.baseURL
	queryURL.Path = "/api/dcim/interfaces/"

	queryHeaders := http.Header{}
	queryHeaders.Add("accept", "application/json")
	queryHeaders.Add("authorization", "token "+nb.token)

	queryVals := url.Values{}
	queryVals.Add("device_id", fmt.Sprintf("%d", deviceID))
	queryVals.Add("mac_address__n", "null")

	queryURL.RawQuery = queryVals.Encode()

	req := &http.Request{
		URL:    &queryURL,
		Header: queryHeaders,
	}
	morepages := true
	interfaces := []Interface{}
	for morepages {
		resp, err := nb.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)
		data := struct {
			Count    int
			Next     string
			Previous string
			Results  []Interface
		}{}

		if err := dec.Decode(&data); err != nil {
			return nil, err
		}
		interfaces = append(interfaces, data.Results...)

		if data.Next != "" {
			req.URL, _ = url.Parse(data.Next)
		} else {
			morepages = false
		}
	}
	return interfaces, nil
}
