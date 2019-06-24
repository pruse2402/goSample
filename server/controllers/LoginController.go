package controllers

import (
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"gosample/server/conf"
	"gosample/server/dbcon"
	"gosample/server/models"
)

type Credential struct {
	EmailID  string `json:"emailID"`
	Password string `json:"password"`
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	credential := Credential{}

	if !parseJson(w, r.Body, &credential) {
		return
	}

	err := db.C("user").Find(bson.M{"email": credential.EmailID, "password": credential.Password}).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "Invalid username or password"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	// session storage
	session, _ := SessionStore.New(r, conf.CookieName)
	session.Values["loggedInUser"] = user.ID.Hex()
	session.Save(r, w)

	res := errorData{Message: "logged in successfully"}
	renderJSON(w, http.StatusOK, res)
}

func GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}

	session, _ := SessionStore.Get(r, conf.CookieName)

	if _, ok := session.Values["loggedInUser"]; !ok {
		e := errorData{}
		e.Message = "Logged in user not found"
		renderJSON(w, http.StatusUnauthorized, e)
		return
	}

	sessionUserID_s, _ := session.Values["loggedInUser"].(string)

	if !bson.IsObjectIdHex(sessionUserID_s) {
		e := errorData{}
		e.Message = "Invalid session user ID"
		renderJSON(w, http.StatusUnauthorized, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(sessionUserID_s)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "Logged in user not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusUnauthorized, e)
		return
	}

	renderJSON(w, http.StatusOK, user)
}
