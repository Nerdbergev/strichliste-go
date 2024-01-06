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

func NewUserRepository(db *sql.DB) UserRepository {
	return UserRepository{db: db}
}

type UserRepository struct {
	db *sql.DB
}

func (r UserRepository) All(includeDeleted bool) ([]User, error) {
	rows, err := r.db.Query("SELECT * FROM user")
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

func (r UserRepository) CreateUser(u User) (User, error) {
	u.Created = time.Now()
	res, err := r.db.Exec("INSERT INTO user (name, email, created, balance, disabled, updated) VALUES ($1, $2, $3, 0, false, $4)", u.Name,
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

func (r UserRepository) FindByName(name string) (User, error) {
	row := r.db.QueryRow("SELECT * FROM user WHERE name = ?", name)
	return processRow(row)
}

func (r UserRepository) FindById(id int64) (User, error) {
	row := r.db.QueryRow("SELECT * FROM user WHERE id = ?", id)
	return processRow(row)
}

func processRow(r *sql.Row) (User, error) {
	var user User
	err := r.Scan(&user.ID, &user.Name, &user.Email, &user.Balance, &user.Disabled, &user.Created,
		&user.Updated)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
