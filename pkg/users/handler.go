package users

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/model"
)

type Handler struct {
	ur model.UserRepository
}

func NewHandler(ur model.UserRepository) Handler {
	return Handler{
		ur: ur,
	}
}

func (h Handler) FindById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	user, err := h.ur.FindById(id)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewUserResponse(user))
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	activeParam := r.URL.Query().Get("active")
	var (
		users []model.User
		err   error
	)
	if activeParam == "" {
		users, err = h.ur.AllUsers()
		if err != nil {
			_ = render.Render(w, r, ErrRender(err))
			return
		}
	} else {
		isActive, perr := strconv.ParseBool(activeParam)
		if perr != nil {
			_ = render.Render(w, r, ErrRender(perr))
			return
		}
		if isActive {
			users, err = h.ur.AllActive()
		} else {
			users, err = h.ur.AllInactive()
		}
	}
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	if err := render.Render(w, r, NewUserListResponse(users)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	data := &UserRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	created, err := h.ur.CreateUser(model.User{Name: data.Name})
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
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Balance    int        `json:"balance"`
	IsActive   bool       `json:"isActive"`
	IsDisabled bool       `json:"isDisabled"`
	Created    time.Time  `json:"created"`
	Updated    *time.Time `json:"updated"`
}

func MapUser(u model.User) User {
	resp := User{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		Balance:    u.Balance,
		IsDisabled: u.Disabled,
		Created:    u.Created,
		Updated:    u.Updated,
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
