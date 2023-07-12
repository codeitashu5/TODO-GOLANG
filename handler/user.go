package handler

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"net/http"
	"simpleHttpRequest/database/dbHelper"
	"simpleHttpRequest/middlewares"
	"simpleHttpRequest/models"
	"simpleHttpRequest/utils"
	"strconv"
	"time"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var body models.UserInfo
	err := utils.ParseBody(r.Body, &body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "CreateUser: failed to parse request body")
		return
	}

	result, err := dbHelper.CheckIfUserExists(body.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "CreateUser:check failed")
		return
	}

	if !result {
		password, errPassword := utils.HashPassword(body.Password)
		if errPassword != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "CreateUser:failed to hash password")
			return
		}
		id, createErr := dbHelper.CreateUser(body.Name, body.Email, password)
		if createErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "CreateUser:failed to crate request")
			return
		}
		createTokenAndForTheUser(body.Email, id, w)
		return
	}

	utils.RespondError(w, http.StatusBadRequest, err, "CreateUser: email already exists")

}

func Health(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "server is running"})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var body models.LoginInfo
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusUnauthorized, parseErr, "LoginUser: failed to parse request body")
		return
	}

	//get password for the email
	passwordDb, errEmail, emailPresent := dbHelper.GetPasswordUsingEmail(body.Email)

	if errEmail != nil {
		utils.RespondError(w, http.StatusInternalServerError, errEmail, "LoginUser: failed to get email")
		return
	}

	if !emailPresent {
		utils.RespondError(w, http.StatusBadRequest, nil, "LoginUser: email do not exist")
		return
	}

	//check if password is correct
	passwordErr := utils.CheckPassword(body.Password, passwordDb)
	if passwordErr != nil {
		utils.RespondError(w, http.StatusBadRequest, nil, "LoginUser: password is not correct")
		return
	}

	id, userErr := dbHelper.GetUserIdForTheUser(body.Email, passwordDb)
	if userErr != nil {
		utils.RespondError(w, http.StatusBadRequest, userErr, "LoginUser: failed to get user")
		return
	}
	createTokenAndForTheUser(body.Email, id, w)
}

func createTokenAndForTheUser(email string, id int, w http.ResponseWriter) {
	expirationTime := time.Now().Add(time.Minute * 20)
	sessionToken := utils.HashString(email + time.Now().String())

	//put-ing session into session token
	_, err := dbHelper.CrateSessionForTheUser(id, sessionToken)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "LoginUser: un-able to crate session")
		return
	}

	payLoad := models.Claims{
		UserId:         id,
		SessionToke:    sessionToken,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}

	//this will crate the token for us
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payLoad)
	tokenString, tokenErr := token.SignedString(models.JwtKey)
	if tokenErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, tokenErr, "LoginUser: failed to get user")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{tokenString})
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, nil, "LogoutUser: Context not found")
		return
	}

	err := dbHelper.LogOutUser(user.SessionToke)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, nil, "LogoutUser: Session not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"Log-out success"})

}

func CreateNewTask(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Task string `json:"task"`
	}{}
	user, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "CreateNewTask: failed to parse request body")
		return
	}

	task, err := dbHelper.CreateTask(user.UserId, body.Task)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "CreateNewTask: unable to create task")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string `json:"message"`
		Task    string `json:"task"`
	}{
		"task created successfully",
		task,
	})

}

func GetAllTaskForTheUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	searchText := r.URL.Query().Get("searchText")
	isTaskCompleted := r.URL.Query().Get("isCompleted")
	task, err := dbHelper.GetAllTaskForTheUser(user.UserId, searchText, isTaskCompleted)

	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "GetAllTaskForTheUser : unable to get task for the user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Task []models.Task `json:"task"`
	}{task})

}

func GetAllCompletedTaskForTheUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var task []models.Task
	var err error

	query := r.URL.Query().Get("isCompleted")
	var isCompleted = true

	if query == "false" {
		isCompleted = false
	}
	task, err = dbHelper.GetAllCompletedTask(user.UserId, isCompleted)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "GetAllCompletedTaskForTheUser: unable to get completed task for the user")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Task []models.Task `json:"task"`
	}{task})

}

func SearchUserWithName(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	searchText := r.URL.Query().Get("searchText")
	users, err := dbHelper.SearchUser(searchText)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "SearchUserWithName: unable to get completed task for the user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Task []models.SelectedUser `json:"selectedUser"`
	}{Task: users})

}

func SetTaskStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var body models.SetStatus

	parseErr := utils.ParseBody(r.Body, &body)
	if parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "SetTaskStatus: failed to parse request body")
		return
	}

	status, err := dbHelper.SetTaskStatus(body.TaskId, user.UserId, body.IsTaskCompleted)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, err, "SetTaskStatus: No row exist for the task")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "SetTaskStatus: Task not found")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, struct {
		Status bool `json:"status"`
	}{Status: status})

}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(middlewares.UserContext).(models.Claims)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	path := chi.URLParam(r, "taskId")
	taskId, errConv := strconv.Atoi(path)
	if errConv != nil {
		utils.RespondError(w, http.StatusBadRequest, errConv, "DeleteTask: Invalid path")
		return
	}

	status, err := dbHelper.DeleteTask(taskId)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, err, "DeleteTask: No row exist for the user")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "DeleteTask: Failed to execute")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
		Task    string `json:"task"`
	}{
		Task:    status,
		Message: "Deleted successfully",
	})
}
