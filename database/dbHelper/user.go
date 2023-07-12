package dbHelper

import (
	"database/sql"
	"simpleHttpRequest/database"
	"simpleHttpRequest/models"
)

func CreateUser(name, email, password string) (int, error) {
	SQL := `INSERT INTO users(name, email, password)
			VALUES ($1, TRIM(LOWER($2)), $3)
			RETURNING id` //asking for the result
	var userID int
	err := database.Todo.QueryRowx(SQL, name, email, password).Scan(&userID) //to make changes in the table
	if err != nil {
		return -1, err
	}

	return userID, nil
}

func CheckIfUserExists(email string) (bool, error) {
	var isExist bool
	SQL := `select count(*) > 0 as is_exist
			from users
			where email = $1`
	err := database.Todo.Get(&isExist, SQL, email)
	if err != nil {
		return isExist, err
	}

	return isExist, nil
}

func GetUserUsingToken(token string) (models.User, error) {
	var user models.User
	SQL := `SELECT name,email,password,id FROM users u 
            INNER JOIN user_session us on u.id = us.user_id 
            WHERE us.session_token = $1` //asking for the result

	err := database.Todo.Get(&user, SQL, token) //to get result form db
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserIdForTheUser(email, password string) (int, error) {
	var id int
	SQL := `SELECT id FROM users 
            WHERE password = $1 AND  email = $2`
	err := database.Todo.Get(&id, SQL, password, email)
	if err != nil {
		return -1, err
	}
	return id, nil

}

func GetPasswordUsingEmail(email string) (string, error, bool) {
	var password string
	SQL := `SELECT password FROM users 
            WHERE email = $1`
	err := database.Todo.Get(&password, SQL, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, false
		}
		return "", err, false
	}
	return password, nil, true

}

func CrateSessionForTheUser(userId int, token string) (int, error) {
	SQL := `INSERT INTO user_session(user_id,session_token) VALUES ($1,$2)`
	err := database.Todo.QueryRowx(SQL, userId, token).Err()
	if err != nil {
		return userId, err
	}

	return userId, nil
}

func DeleteSessionForTheUser(token string) error {
	SQL := `DELETE FROM user_session us WHERE us.session_token = $1`
	err := database.Todo.QueryRowx(SQL, token).Err()
	if err != nil {
		return err
	}
	return nil
}

func CreateTask(userId int, taskDb string) (string, error) {
	var task string
	SQL := `INSERT INTO task(task, user_id, is_task_completed) 
            VALUES($1, $2, false) RETURNING task
            `
	err := database.Todo.Get(&task, SQL, taskDb, userId)
	if err != nil {
		return task, err
	}

	return task, err
}

func GetAllTaskForTheUser(userId int, searchText string, isCompleted string) ([]models.Task, error) {
	task := make([]models.Task, 0)
	SQL := `SELECT task, is_task_completed FROM task 
            WHERE user_id = $1 AND task ILIKE  '%'||$2||'%' AND archived_at IS  NULL `

	//Whether we need to add another query or not depends upon the
	var err error
	if isCompleted == "true" || isCompleted == "false" {
		SQL = SQL + "AND is_task_completed = $3"
		isCompletedBool := isCompleted == "true"
		err = database.Todo.Select(&task, SQL, userId, searchText, isCompletedBool)
	} else {
		err = database.Todo.Select(&task, SQL, userId, searchText)
	}

	if err != nil {
		return task, err
	}
	return task, nil
}

func SetTaskStatus(taskId, userId int, isCompleted bool) (bool, error) {
	var task bool
	SQL := ` UPDATE task t SET is_task_completed = $1 
             WHERE user_id = $2 AND t.id = $3 AND t.archived_at IS NULL RETURNING is_task_completed`
	err := database.Todo.QueryRowx(SQL, isCompleted, userId, taskId).Scan(&task)
	if err != nil {
		return task, err
	}
	return task, nil
}

func DeleteTask(taskId int) (string, error) {
	var task string
	SQL := `UPDATE task SET archived_at = now()	WHERE id = $1 AND archived_at IS NULL RETURNING task`
	err := database.Todo.QueryRowx(SQL, taskId).Scan(&task)
	return task, err
}

// GetAllCompletedTask to get all completed task for the user
func GetAllCompletedTask(userId int, isCompleted bool) ([]models.Task, error) {
	//I am going to return the users name and the task associated with him
	var task []models.Task
	SQL := `SELECT t.task , is_task_completed  FROM task t  
            WHERE user_id = $1 AND is_task_completed = $2 AND archived_at IS NULL`
	err := database.Todo.Select(&task, SQL, userId, isCompleted)
	if err != nil {
		return task, err
	}
	return task, nil
}

//search from all the user

func SearchUser(searchText string) ([]models.SelectedUser, error) {
	//I am going to return the users name and the task associated with him
	var users []models.SelectedUser
	SQL := `SELECT u.name , u.email  FROM users u
            WHERE u.name like '%'||$1||'%' `
	err := database.Todo.Select(&users, SQL, searchText)
	if err != nil {
		return users, err
	}
	return users, nil
}

func CheckIfUserSessionIsValid(token string) (bool, error) {
	var isTokenPresent bool
	SQL := `SELECT count(*)> 0 FROM user_session WHERE session_token = $1 AND archived_at IS NULL`
	err := database.Todo.Get(&isTokenPresent, SQL, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return isTokenPresent, nil
}

func LogOutUser(token string) error {
	SQL := `UPDATE user_session SET archived_at = now() WHERE session_token = $1`
	_, err := database.Todo.Exec(SQL, token)
	if err != nil {
		return err
	}
	return nil
}
