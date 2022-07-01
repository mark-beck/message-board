package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type AuthServer struct {
	url string
}

type UserInfo struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	Image string   `json:"image"`
}

func init_auth() AuthServer {
	auth_name := os.Getenv("AUTH_NAME")
	auth_port := os.Getenv("AUTH_PORT")
	if auth_name == "" {
		auth_name = "localhost"
	}
	if auth_port == "" {
		auth_port = "8080"
	}

	url := fmt.Sprintf("http://%s:%s", auth_name, auth_port)
	return AuthServer{url}
}

func (auth AuthServer) get_info(token string) (UserInfo, error) {
	req, err := http.NewRequest("GET", auth.url+"auth/user/info", nil)
	if err != nil {
		return UserInfo{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}

	if resp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("auth server returned %d", resp.StatusCode)
	}

	bytes := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(bytes)
	if err != nil {
		return UserInfo{}, err
	}

	var info UserInfo

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return UserInfo{}, err
	}

	return info, nil

}
