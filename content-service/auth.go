package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
)

type AuthServer struct {
	url string
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

func (auth AuthServer) get_info(ctx echo.Context, token string) (UserInfo, error) {
	sp := jaegertracing.CreateChildSpan(ctx, "get_info")
	defer sp.Finish()

	req, err := jaegertracing.NewTracedRequest("GET", auth.url+"/auth/user/info", nil, sp)
	if err != nil {
		return UserInfo{}, err
	}

	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}

	if resp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("auth server returned %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
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

// get user by id
func (auth AuthServer) get_user(ctx echo.Context, id string) (UserInfo, error) {
	sp := jaegertracing.CreateChildSpan(ctx, "get_user")
	defer sp.Finish()

	req, err := jaegertracing.NewTracedRequest("GET", auth.url+"/auth/user/"+id, nil, sp)
	if err != nil {
		return UserInfo{}, err
	}

	req.Header.Set("Authorization", ctx.Request().Header["Authorization"][0])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}

	if resp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("auth server returned %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
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

func (auth AuthServer) get_user_batch(ctx echo.Context, ids []string) ([]UserInfo, error) {
	sp := jaegertracing.CreateChildSpan(ctx, "get_user_batch")
	defer sp.Finish()

	req, err := jaegertracing.NewTracedRequest("POST", auth.url+"/auth/user/get_batch", nil, sp)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf("[\"%s\"]", strings.Join(ids, "\",\"")))))

	req.Header.Set("Authorization", ctx.Request().Header["Authorization"][0])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("auth server returned %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	Info(sp, "response data", bytes)

	var info []UserInfo

	err = json.Unmarshal(bytes, &info)

	if err != nil {
		return nil, err
	}

	Info(sp, "response json", info)

	return info, nil
}
