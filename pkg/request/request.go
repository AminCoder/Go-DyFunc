package req

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Trigger_client struct {
	Base_url string
	Username string
	Password string
	client   *http.Client
}

func New_trigger_client(base_url, username, password string) *Trigger_client {
	return &Trigger_client{
		Base_url: base_url,
		Username: username,
		Password: password,
		client:   &http.Client{},
	}
}

func (tc *Trigger_client) Send_request(id string, function_name string, args interface{}) ([]byte, error) {
	payload := []struct {
		Id   string      `json:"id"`
		Func string      `json:"func"`
		Args interface{} `json:"args"`
	}{
		{
			Id:   id,
			Func: function_name,
			Args: args,
		},
	}
	json_data, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshalling payload:", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", tc.Base_url, bytes.NewBuffer(json_data))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(tc.Username+":"+tc.Password))
	req.Header.Set("Authorization", basicAuth)

	resp, err := tc.client.Do(req)
	if err != nil {
		log.Println("Error sending POST:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}
