package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/heroku/huhu-backend-app/app/models"
	"github.com/heroku/huhu-backend-app/app/utils"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func InputUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var project models.User
	err := decoder.Decode(&project)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	// project.Id = utils.GenerateId()

	var user models.User
	models.OneUserLogin(db, project.Email, &user)
	if user == (models.User{}) {
		err = models.InsertUser(db, &project)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusCreated, project)
	} else {
		respondJSON(w, http.StatusNotAcceptable, nil)
	}
}

func UpdateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var users models.User
	id, _ := strconv.Atoi(vars["id"])
	err := models.OneUserGetting(db, id, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&users)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	err = models.UpdateUser(db, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, users)
}

func DeletedUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var users models.User
	id, _ := strconv.Atoi(vars["id"])
	err := models.OneUserGetting(db, id, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = models.DeletedUser(db, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, users)
}

func Login(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var project models.User
	err := decoder.Decode(&project)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var users models.User
	err = models.OneUserLogin(db, project.Email, &users)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Wrong email or password")
		return
	}
	if users.Password != "" && users.Password == project.Password {
		respondJSON(w, http.StatusAccepted, users)
		return
	}
	respondError(w, http.StatusUnauthorized, "Wrong email or password")
	return
}

func VerifyByEmail(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var project models.User
	err := decoder.Decode(&project)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var users models.User
	err = models.OneUserLogin(db, project.Email, &users)
	if err != nil {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	if users == (models.User{}) {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	respondJSON(w, http.StatusOK, users)
	return
}

func VerifyByResetKey(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var project models.User
	err := decoder.Decode(&project)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var users models.User
	err = models.OneUserResetKey(db, project.ResetKey, &users)
	if err != nil {
		println(err.Error())
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	if users == (models.User{}) {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	respondJSON(w, http.StatusOK, users)
	return
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	fmt.Println("status ", status)
	var res utils.ResponseData

	res.Status = status
	res.Meta = utils.ResponseMessage(status)
	res.Data = payload

	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func respondError(w http.ResponseWriter, status int, message string) {
	var res utils.ResponseData
	rescode := utils.ResponseMessage(status)
	res.Status = status
	res.Meta = rescode
	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))

}
