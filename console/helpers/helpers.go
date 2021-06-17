package helpers

import (
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/iostrovok/aura-test/response"
)

const host = "http://127.0.0.1:8080"

func List(printAlls ...bool) {
	printAll := false
	if len(printAlls) > 0 {
		printAll = printAlls[0]
	}

	resp, err := http.Get(host + "/sessions")
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return
	}

	out := make([]interface{}, 0)
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &out); err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return
	}

	if printAll || len(out) < 101 {
		fmt.Printf("%+v\n", out)
	} else if len(out) > 100 {
		fmt.Printf("%+v .... \n", out[:100])
	}
	fmt.Printf("Total: %d\n\n", len(out))
}

func DestroySession(client *http.Client, id string) (int, error) {
	req, err := http.NewRequest(http.MethodDelete, host+"/sessions/"+id, nil)
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return 0, err
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return 0, err
	}

	return resp.StatusCode, nil
}

func CreateSession(client *http.Client) string {
	req, err := http.NewRequest(http.MethodPost, host+"/sessions", nil)
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return ""
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return ""
	}

	out := &response.Response{}
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &out); err != nil {
		logrus.Errorf("err: %s\n", err.Error())

		return ""
	}

	return out.ID
}
