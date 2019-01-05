package controllers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/heroku/huhu-backend-app/app/models"
	"github.com/jinzhu/gorm"
)

func ListUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	users := []models.User{}
	err := models.GetAllUser(db, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, users)
}

func OneUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var users models.User
	id, _ := strconv.Atoi(vars["id"])
	err := models.OneUserGetting(db, id, &users)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		respondJSON(w, http.StatusOK, users)
		return
	}
}

func Index(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, nil)
	return
}
