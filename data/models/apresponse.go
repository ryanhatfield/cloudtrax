package models

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
)

//RejectCode string representation
const RejectCode string = "REJECT"

//AcceptCode string representation
const AcceptCode string = "ACCEPT"

//OKCode string representation
const OKCode string = "OK"

const responseTemplate string = `"CODE" "{{.ResponseCode}}"
"RA" "{{.ResponseAuthorization}}"
{{if eq .ResponseCode "ACCEPT" -}}
"SECONDS" "{{.Seconds}}"
"DOWNLOAD" "{{.Download}}"
"UPLOAD" "{{.Upload}}"
{{- else if eq .ResponseCode "REJECT" -}}
"BLOCKED_MSG" "{{.BlockedMessage}}"
{{end}}`

//APResponse is the generated object being sent back to the Access Point
type APResponse struct {
	ResponseCode          string
	Request               *APRequest
	ResponseAuthorization string
	Seconds               int32
	Download              int32
	Upload                int32
	BlockedMessage        string
	Secret                string
}

//Execute the APResposne object, and write the response to the io writer
func (apr *APResponse) Execute(w *http.ResponseWriter) error {
	var err error

	//Pull in our template
	t := template.Must(template.New("response").Parse(responseTemplate))

	if err != nil {
		return err
	}

	//Execute the response template, and write to the response
	err = t.Execute(*w, *apr)
	if err != nil {
		return err
	}

	//return nil; no news is good news.
	return nil
}

//GenerateRA takes the response CODE, the (un-decoded) RA field, and the site secret,
//and generates the Response Authentication token.
//NOTE: I don't like this method, it will be updated/changed/mamed at some point.
func GenerateRA(code string, ra string, secret string) (string, error) {
	var buffer bytes.Buffer
	var err error
	hasher := md5.New()

	decodedRa, err := hex.DecodeString(ra)
	if err != nil {
		return "", fmt.Errorf(
			"An error has occured while decoding the hex string.\n%s", err.Error())
	}
	buffer.WriteString(code)
	buffer.WriteString(string(decodedRa))
	buffer.WriteString(secret)
	_, err = hasher.Write(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf(
			"An error has occured while writing to the md5 hasher.\n %s", err.Error())
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

//NewAPResponse generates a new APResponse from the APRequest
func NewAPResponse(req *APRequest) *APResponse {
	res := &APResponse{Request: req}
	switch req.RequestType {
	case AccountingRequest:
		res.ResponseCode = OKCode
	case StatusRequest:
		res.ResponseCode = RejectCode
		res.BlockedMessage = "Your session has expired."
	case LoginRequest:
		//TODO: Check login credentials here
		res.ResponseCode = AcceptCode
		res.Seconds = 3600
		res.Download = 2000
		res.Upload = 800
	default:
		panic(fmt.Errorf("Error: %v, URL: %v", "incorrect request type", req.RequestType))
	}
	return res
}
