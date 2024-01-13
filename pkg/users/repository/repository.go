package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/nerdbergev/shoppinglist-go/pkg/database"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

type User struct {
	ID         int64
	Name       string
	Email      sql.NullString
	Balance    int64
	IsDisabled bool
	Created    time.Time
	Updated    sql.NullTime
}

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *sql.DB
}

func (r Repository) GetAll() ([]domain.User, error) {
	rows, err := r.db.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) FindActive(ut time.Time) ([]domain.User, error) {
	rows, err := r.db.Query("SELECT * FROM user WHERE updated > ?", ut)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) FindInactive(ut time.Time) ([]domain.User, error) {
	rows, err := r.db.Query("SELECT * FROM user WHERE updated < ? or updated is null", ut)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) AllActive() ([]User, error) {
	rows, err := r.db.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled, &u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		user = append(user, u)
	}
	return user, nil
}

func (r Repository) AllInactive() ([]User, error) {
	rows, err := r.db.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled, &u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		user = append(user, u)
	}
	return user, nil
}

func (r Repository) StoreUser(u domain.User) (domain.User, error) {
	u.Created = time.Now()
	res, err := r.db.Exec("INSERT INTO user (name, email, created, balance, disabled, updated) VALUES ($1, $2, $3, 0, false, $4)", u.Name,
		u.Email, u.Created, nil)
	if err != nil {
		return domain.User{}, err
	}
	u.ID, err = res.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}

	return u, nil
}

func (r Repository) FindByName(name string) (domain.User, error) {
	row := r.db.QueryRow("SELECT * FROM user WHERE name = ?", name)
	return processRow(row)
}

func (r Repository) FindById(ctx context.Context, id int64) (domain.User, error) {
	row := r.getDB(ctx).QueryRow("SELECT * FROM user WHERE id = ?", id)
	return processRow(row)
}

func (r Repository) UpdateUser(domain.User) error {
	return nil
}

func (r Repository) DeleteById(int64) error {
	return nil
}

type DB interface {
	QueryRow(string, ...any) *sql.Row
}

func (r Repository) getDB(ctx context.Context) DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}

func processRow(r *sql.Row) (domain.User, error) {
	var user User
	err := r.Scan(&user.ID, &user.Name, &user.Email, &user.Balance, &user.IsDisabled, &user.Created,
		&user.Updated)
	if err != nil {
		return domain.User{}, err
	}
	return mapToDomain(user), nil
}

func processRows(r *sql.Rows) ([]domain.User, error) {
	var users []domain.User
	for r.Next() {
		var u User
		err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled, &u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		users = append(users, mapToDomain(u))
	}
	return users, nil
}

func mapToDomain(u User) domain.User {
	du := domain.User{
		ID:         u.ID,
		Name:       u.Name,
		Balance:    u.Balance,
		IsDisabled: u.IsDisabled,
		Created:    u.Created,
	}
	if u.Email.Valid {
		du.Email = new(string)
		*du.Email = u.Email.String
	}
	if u.Updated.Valid {
		du.Updated = new(time.Time)
		*du.Updated = u.Updated.Time
	}
	return du
}
