package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID       int64
	Name     string
	Email    string
	Balance  int
	Disabled bool
	Created  time.Time
	Updated  *time.Time
}

func New(db *sql.DB) UserModel {
	return UserModel{db: db}
}

type UserModel struct {
	db *sql.DB
}

func (m UserModel) All(includeDeleted bool) ([]User, error) {
	rows, err := m.db.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user []User

	for rows.Next() {
		var u User

		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Balance, &u.Disabled, &u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		user = append(user, u)
	}
	return user, nil
}

func (m UserModel) CreateUser(u User) (User, error) {
	u.Created = time.Now()
	res, err := m.db.Exec("INSERT INTO user (name, email, created, balance, disabled, updated) VALUES ($1, $2, $3, 0, false, $4)", u.Name,
		u.Email, u.Created, nil)
	if err != nil {
		return User{}, err
	}
	u.ID, err = res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (m UserModel) FindByName(name string) (User, error) {
	var user User
	err := m.db.QueryRow("SELECT * FROM user where name = ?", name).
		Scan(&user.ID, &user.Name, &user.Email, &user.Balance, &user.Disabled, &user.Created,
			&user.Updated)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
