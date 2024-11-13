// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

// This file is borrowed and further adjusted from https://github.com/goharbor/terraform-provider-harbor/blob/main/client/client.go
package provider

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NewTokenClient creates common settings
func NewTokenClient(url string, bearerToken string, insecure bool) *S3GridClient {
	gridClient := &S3GridClient{
		address:    url,
		token:      bearerToken,
		insecure:   insecure,
		httpClient: &http.Client{},
	}

	return gridClient
}

// NewUsernamePasswordClient creates common settings
func NewUsernamePasswordClient(url string, username string, password string, tenant string, insecure bool) *S3GridClient {
	gridClient := &S3GridClient{
		address:    url,
		username:   username,
		password:   password,
		tenant:     tenant,
		insecure:   insecure,
		httpClient: &http.Client{},
	}

	return gridClient
}

// SendRequest send a http request
func (c *S3GridClient) SendAuthorizeRequest(statusCode int) (tokenValue string, respCode int, err error) {
	var jsonD S3GridClientReturnJson

	address := c.address + api_auth
	client := &http.Client{}

	postRequest := &S3GridClientJson{
		AccountId: c.tenant,
		Password:  c.password,
		Username:  c.username,
		Cookie:    true,
		CsrfToken: false,
	}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(postRequest)
	if err != nil {
		return "", 0, err
	}

	if c.insecure {
		tr := &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	req, err := http.NewRequest("POST", address, b)
	if err != nil {
		return "", 0, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return "", resp.StatusCode, err
		} else {
			return "", http.StatusBadGateway, err
		}
	}

	body, err := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &jsonD); err != nil {
		return "", 0, err
	}

	if err != nil {
		return "", resp.StatusCode, err
	}
	resp.Body.Close()

	if statusCode != 0 && resp.StatusCode != statusCode {
		return "", resp.StatusCode, fmt.Errorf("[ERROR] unexpected status code got: %v expected: %v \n %v", statusCode, resp.StatusCode, statusCode)
	}

	return jsonD.Data, resp.StatusCode, nil
}

// SendRequest send a http request
func (c *S3GridClient) SendRequest(method string, path string, payload interface{}, statusCode int) (value []byte, respheaders string, respCode int, err error) {
	address := c.address + path
	bearer := "Bearer " + c.token
	client := &http.Client{}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(payload)
	if err != nil {
		return nil, "", 0, err
	}

	if c.insecure {
		tr := &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	req, err := http.NewRequest(method, address, b)
	if err != nil {
		return nil, "", 0, err
	}

	// Use access token authentification if bearer Token is specified
	if c.token != "" {
		req.Header.Add("Authorization", bearer)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return nil, "", resp.StatusCode, err
		} else {
			return nil, "", http.StatusBadGateway, err
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", resp.StatusCode, err
	}
	resp.Body.Close()
	strbody := string(body)
	respHeaders := resp.Header
	headers, err := json.Marshal(respHeaders)
	if err != nil {
		return nil, "", resp.StatusCode, err
	}

	if statusCode != 0 && resp.StatusCode != statusCode {
		return nil, "", resp.StatusCode, fmt.Errorf("[ERROR] unexpected status code got: %v expected: %v \n %v", resp.StatusCode, statusCode, strbody)
	}

	return body, string(headers), resp.StatusCode, nil
}
