package controllers

import (
"encoding/json"
"io"
"io/ioutil"
"net/http"

"github.com/gorilla/sessions"

"gosample/server/conf"
)

type errorData struct {
	Message          string                 `json:"message"`
	Error            string                 `json:"error,omitempty"`
	ValidationErrors map[string]interface{} `json:"validationErrors,omitempty"`
}

var SessionStore = sessions.NewCookieStore([]byte(conf.SecretKey))

func renderJSON(w http.ResponseWriter, statusCode int, resp interface{}) {
	resByte, _ := json.MarshalIndent(resp, "", "	")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(resByte)
}

func parseJson(w http.ResponseWriter, body io.ReadCloser, model interface{}) bool {
	defer body.Close()

	b, _ := ioutil.ReadAll(body)
	err := json.Unmarshal(b, model)
	if err != nil {
		e := errorData{}
		e.Message = "Error in parsing json"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return false
	}

	return true
}
