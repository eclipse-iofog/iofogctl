package iofog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/util"
	"net/http"
	"strings"
)

type Controller struct {
	baseURL string
}

func NewController(endpoint string) *Controller {
	return &Controller{
		baseURL: fmt.Sprintf("http://%s/api/v3/", endpoint),
	}
}

func (ctrl *Controller) GetStatus() (status string, timestamp string, err error) {
	url := ctrl.baseURL + "status"
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
		return
	}

	var respMap map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &respMap)
	if err != nil {
		return
	}
	status, exists := respMap["status"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get status from Controller")
		return
	}
	//timestamp, exists = respMap["timestamp"].(string)
	//if !exists {
	//	err = util.NewInternalError("Failed to get timestamp from Controller")
	//	return
	//}
	return
}

func (ctrl *Controller) CreateUser(user User) error {
	contentType := "application/json"
	userString := fmt.Sprintf("{ \"firstName\": \"%s\", \"lastName\": \"%s\", \"email\": \"%s\", \"password\": \"%s\" }", user.Name, user.Surname, user.Email, user.Password)
	signupBody := strings.NewReader(userString)
	url := ctrl.baseURL + "user/signup"
	resp, err := http.Post(url, contentType, signupBody)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
	}
	return nil
}

func (ctrl *Controller) GetAuthToken(user User) (token string, err error) {
	contentType := "application/json"
	userString := fmt.Sprintf("{\"email\":\"%s\",\"password\":\"%s\"}", user.Email, user.Password)
	loginBody := strings.NewReader(userString)
	url := ctrl.baseURL + "user/login"
	resp, err := http.Post(url, contentType, loginBody)
	if err != nil {
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
		return
	}

	// Read access token from HTTP response
	var auth map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &auth)
	if err != nil {
		return
	}
	token, exists := auth["accessToken"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get auth token from Controller")
		return
	}
	return
}

func (ctrl *Controller) CreateAgent(authToken, agentName string) (uuid string, err error) {
	contentType := "application/json"
	bodyString := fmt.Sprintf("{\"name\":\"%s\",\"fogType\":0}", agentName)
	body := strings.NewReader(bodyString)
	url := ctrl.baseURL + "iofog"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", contentType)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
		return
	}

	// Read uuid from response
	var respMap map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &respMap)
	if err != nil {
		return
	}
	uuid, exists := respMap["uuid"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get new Agent UUID from Controller")
		return
	}
	return
}

func (ctrl *Controller) GetAgentProvisionKey(authToken, agentUUID string) (key string, err error) {
	contentType := "application/json"
	url := ctrl.baseURL + "iofog/" + agentUUID + "/provisioning-key"
	body := strings.NewReader("")
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", contentType)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
		return
	}

	var respMap map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &respMap)
	if err != nil {
		return
	}
	key, exists := respMap["key"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get provisioning key from Controller")
		return
	}
	return
}

func (ctrl *Controller) DeleteAgent(authToken, agentUUID string) error {
	contentType := "application/json"
	url := ctrl.baseURL + "iofog/" + agentUUID
	body := strings.NewReader("")
	req, err := http.NewRequest("DELETE", url, body)
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", contentType)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = util.NewInternalError(fmt.Sprintf("Received %d from Controller", resp.StatusCode))
		return err
	}

	return nil
}
