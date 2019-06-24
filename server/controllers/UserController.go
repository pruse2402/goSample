package controllers

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"

	"gosample/server/dbcon"
	"gosample/server/models"
)

func ListUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	users := []models.User{}

	db.C("user").Find(nil).All(&users)

	renderJSON(w, http.StatusOK, users)
}

func ActiveUsers(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	users := []models.User{}

	db.C("user").Find(bson.M{"activeStatus": true}).All(&users)

	renderJSON(w, http.StatusOK, users)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params, _ := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	renderJSON(w, http.StatusOK, user)
}

func SaveUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}

	if !parseJson(w, r.Body, &user) {
		return
	}

	user.DOB, _ = time.Parse("02-Jan-2006", user.DOBStr)
	user.DateCreated = time.Now().UTC()
	user.LastUpdated = time.Now().UTC()

	hasErr, errs := user.Validate(db)
	if hasErr {
		e := errorData{}
		e.Message = "Error in validation"
		e.ValidationErrors = errs
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	rec, err := db.C("user").Upsert(bson.M{"upsert": ""}, &user)
	if err != nil {
		e := errorData{}
		e.Message = "Error in saving user"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	if rec.UpsertedId != nil {
		user.ID, _ = rec.UpsertedId.(bson.ObjectId)
	}

	res := struct {
		Message string      `json:"message"`
		User    models.User `json:"user"`
	}{}

	res.Message = "User saved successfully"
	res.User = user
	renderJSON(w, http.StatusOK, res)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params, _ := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	if !parseJson(w, r.Body, &user) {
		return
	}

	user.DOB, _ = time.Parse("02-Jan-2006", user.DOBStr)
	user.LastUpdated = time.Now().UTC()

	hasErr, errs := user.Validate(db)
	if hasErr {
		e := errorData{}
		e.Message = "Error in validation"
		e.ValidationErrors = errs
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err = db.C("user").UpdateId(user.ID, &user)
	if err != nil {
		e := errorData{}
		e.Message = "Error in updating user"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	res := struct {
		Message string      `json:"message"`
		User    models.User `json:"user"`
	}{}

	res.Message = "User updated successfully"
	res.User = user
	renderJSON(w, http.StatusOK, res)
}

func ActivateUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params, _ := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	user.ActiveStatus = true
	user.LastUpdated = time.Now().UTC()

	setter := bson.M{"activeStatus": user.ActiveStatus, "lastUpdated": user.LastUpdated}
	err = db.C("user").UpdateId(user.ID, bson.M{"$set": setter})
	if err != nil {
		e := errorData{}
		e.Message = "Error in activating user"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	res := struct {
		Message string        `json:"message"`
		ID      bson.ObjectId `json:"id"`
	}{}

	res.Message = "User activated successfully"
	res.ID = user.ID
	renderJSON(w, http.StatusOK, res)
}

func InactivateUser(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params, _ := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	user.ActiveStatus = false
	user.LastUpdated = time.Now().UTC()

	setter := bson.M{"activeStatus": user.ActiveStatus, "lastUpdated": user.LastUpdated}
	err = db.C("user").UpdateId(user.ID, bson.M{"$set": setter})
	if err != nil {
		e := errorData{}
		e.Message = "Error in inactivating user"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	res := struct {
		Message string        `json:"message"`
		ID      bson.ObjectId `json:"id"`
	}{}

	res.Message = "User inactivated successfully"
	res.ID = user.ID
	renderJSON(w, http.StatusOK, res)
}

func SaveUserImage(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params, _ := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	file, fh, err := r.FormFile("userImage")
	if err != nil {
		e := errorData{}
		e.Message = "Error in reading file"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	defer file.Close()

	if content := fh.Header.Get("Content-Type"); !strings.HasPrefix(content, "image/") {
		e := errorData{}
		e.Message = "Upload valid image file"
		renderJSON(w, http.StatusUnsupportedMediaType, e)
		return
	}

	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, file)
	if err != nil {
		e := errorData{}
		e.Message = "Error in uploading image"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	user.Image = buf.Bytes()
	user.LastUpdated = time.Now().UTC()

	setter := bson.M{"image": user.Image, "lastUpdated": user.LastUpdated}
	err = db.C("user").UpdateId(user.ID, bson.M{"$set": setter})
	if err != nil {
		e := errorData{}
		e.Message = "Error in uploading image"
		e.Error = err.Error()
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	res := errorData{Message: "Image uploaded successfully"}
	renderJSON(w, http.StatusOK, res)
}

func GetUserImage(w http.ResponseWriter, r *http.Request) {
	db := dbcon.CopyDB()
	defer db.Session.Close()

	user := models.User{}
	params := r.Context().Value("params").(httprouter.Params)
	id := params.ByName("id")

	if !bson.IsObjectIdHex(id) {
		e := errorData{}
		e.Message = "Invalid user ID"
		renderJSON(w, http.StatusBadRequest, e)
		return
	}

	err := db.C("user").FindId(bson.ObjectIdHex(id)).One(&user)
	if err != nil {
		e := errorData{}
		e.Message = "User not found"
		e.Error = err.Error()
		renderJSON(w, http.StatusNotFound, e)
		return
	}

	w.Write(user.Image)
}
