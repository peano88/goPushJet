// copyright (c) 2016 geir54

package goPushJet

import (
	"encoding/json"
	"errors"
	"github.com/dghubble/sling"
	"net/http"
	"net/url"
	"strconv"
)

type servResp struct {
	Service Service `json:"service"`
}

type getServResp struct {
	Service Service  `json:"service"`
	Status  string   `json:"status"`
	Error   errorMsg `json:"error"`
}

type getServParams struct {
	Service string `url:"service,omitempty"`
	Secret  string `url:"secret,omitempty"`
}

type updResp struct {
	Status string   `json:"status"`
	Error  errorMsg `json:"error"`
}

type updParams struct {
	Name   string `url:"name,omitempty"`
	Secret string `url:"secret"`
	Icon   string `url:"icon,omitempty"`
}

type delParams struct {
	Secret string `url:"secret"`
}

type delResp struct {
	Status string   `json:"status"`
	Error  errorMsg `json:"error"`
}

type Service struct {
	Created int    `json:"created"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Public  string `json:"public"`
	Secret  string `json:"secret"`
}

type respStatus struct {
	Status string   `json:"status"`
	Error  errorMsg `json:"error"`
}

type errorMsg struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// GetQR - Get QR image
func (serv *Service) GetQR() string {
	return "https://chart.googleapis.com/chart?cht=qr&chl=" + serv.Public + "&choe=UTF-8&chs=200x200"
}

func (serv *Service) IsEmpty() bool {
	return serv.Public == ""
}

func serviceSlingBase() *sling.Sling {
	return sling.New().Base("https://api.pushjet.io/")
}

func checkRespStatus(s respStatus) error {
	if s.Error.Message != "" {
		return errors.New(s.Error.Message)
	}

	if s.Status != "ok" && s.Status != "" {
		return errors.New("Did not return status OK: " + s.Status)
	}
	return nil
}

func UpdateService(secret, newName, newIcon string) error {
	//Create the updateService Params

	usp := updParams{Secret: secret, Name: newName, Icon: newIcon}
	uspResp := new(updResp)
	sbb := serviceSlingBase().New()
	sb := sbb.Patch("service").BodyForm(&usp)
	_, err := sb.ReceiveSuccess(uspResp)
	if err != nil {
		return err
	}

	if err = checkRespStatus(respStatus{
		Status: uspResp.Status,
		Error:  uspResp.Error}); err != nil {

		return err
	}

	return nil

}

func DeleteService(secret string) error {
	dp := &delParams{Secret: secret}
	dResp := new(delResp)
	_, err := serviceSlingBase().Delete("service").BodyForm(dp).ReceiveSuccess(dResp)

	if err != nil {
		return err
	}

	return checkRespStatus(respStatus{
		Status: dResp.Status,
		Error:  dResp.Error})
}
func GetServiceInfo(service, secret string) (Service, error) {
	gsi := &getServParams{Secret: secret, Service: service}
	gsr := new(getServResp)
	path := "https://api.pushjet.io/service"
	// Sling
	_, err := sling.New().Get(path).QueryStruct(gsi).ReceiveSuccess(gsr)

	if err != nil {
		return Service{}, err
	}

	if err = checkRespStatus(respStatus{
		Status: gsr.Status,
		Error:  gsr.Error}); err != nil {
		return Service{}, err
	}
	return gsr.Service, nil
}

// CreateService - Create new service
func CreateService(name, icon string) (Service, error) {
	resp, err := http.PostForm("https://api.pushjet.io/service",
		url.Values{"name": {name}, "icon": {icon}})

	if err != nil {
		return Service{}, err
	}
	defer resp.Body.Close()

	ser := servResp{}
	err = json.NewDecoder(resp.Body).Decode(&ser)
	if err != nil {
		return Service{}, err
	}

	return ser.Service, nil
}

// SendMessage -
// secret: required stringd2d1820d56b862a6f5b1a69a7af730fa The service secret token
// message: required string Your server is on fire! The notification text
// title: string A custom message title
// level: integer 3 The importance level from 1(low) to 5(high)
// link: string http://i.imgur.com/TerUkQY.gif An optional link
func SendMessage(secret, message, title string, level int, link string) error {
	resp, err := http.PostForm("https://api.pushjet.io/message",
		url.Values{"secret": {secret}, "message": {message}, "title": {title}, "level": {strconv.Itoa(level)}, "link": {link}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	msg := respStatus{}
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		return err
	}

	if msg.Error.Message != "" {
		return errors.New(msg.Error.Message)
	}

	if msg.Status != "ok" {
		return errors.New("Did not return status OK")
	}

	return nil
}
