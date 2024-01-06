package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/user/model"
)

type Handler struct {
	um model.UserModel
}

func NewHandler(um model.UserModel) Handler {
	return Handler{
		um: um,
	}
}

func (h Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	u, err := h.um.All(false)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.Render(w, r, NewUserListResponse(u)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	data := &UserRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	created, err := h.um.CreateUser(model.User{Name: data.Name})
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
	render.JSON(w, r, NewUserResponse(created))
}

func NewUserListResponse(users []model.User) UserListResponse {
	list := UserListResponse{}
	for _, u := range users {
		list.Users = append(list.Users, MapUser(u))
	}
	return list
}

type UserListResponse struct {
	Users []User `json:"users"`
}

func (ur UserListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type User struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Balance  int        `json:"balance"`
	Active   bool       `json:"isActive"`
	Disabled bool       `json:"isDisabled"`
	Created  time.Time  `json:"created"`
	Updated  *time.Time `json:"updated"`
}

func MapUser(u model.User) User {
	resp := User{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Balance:  u.Balance,
		Disabled: u.Disabled,
		Created:  u.Created,
		Updated:  u.Updated,
	}
	return resp
}

func NewUserResponse(u model.User) UserResponse {
	return UserResponse{User: MapUser(u)}
}

type UserResponse struct {
	User User `json:"user"`
}

func (ur UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *UserRequest) Bind(r *http.Request) error {
	if u.Name == "" {
		return errors.New("missing required User fields.")
	}

	return nil
}
