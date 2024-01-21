package rest

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/users"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

type Handler struct {
	svc users.Service
}

func NewHandler(svc users.Service) Handler {
	return Handler{
		svc: svc,
	}
}

func (h Handler) FindById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}

	user, err := h.svc.FindById(id)
	if err != nil {
		var derr domain.UserNotFoundError
		if errors.As(err, &derr) {
			_ = render.Render(w, r, ErrUserNotFound(derr))
			return
		}
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
	_ = render.Render(w, r, NewUserResponse(user))
}

type getAllRequest struct {
	active bool
}

func (req getAllRequest) Active() bool {
	return req.active
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	activeParam := r.URL.Query().Get("active")
	var (
		users []domain.User
		err   error
	)
	if activeParam == "" {
		users, err = h.svc.GetAll()
		if err != nil {
			_ = render.Render(w, r, ErrServerError(err))
			return
		}
	} else {
		isActive, perr := strconv.ParseBool(activeParam)
		if perr != nil {
			_ = render.Render(w, r, ErrServerError(perr))
			return
		}
		users, err = h.svc.FindByState(getAllRequest{active: isActive})
	}
	if err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	if err := render.Render(w, r, NewUserListResponse(users)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	uReq := &UserRequest{}
	if err := render.Bind(r, uReq); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	created, err := h.svc.CreateUser(uReq)
	if err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
	render.JSON(w, r, NewUserResponse(created))
}

func NewUserListResponse(users []domain.User) UserListResponse {
	list := UserListResponse{Users: []User{}}
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
	Balance    int64      `json:"balance"`
	IsActive   bool       `json:"isActive"`
	IsDisabled bool       `json:"isDisabled"`
	Created    time.Time  `json:"created"`
	Updated    *time.Time `json:"updated"`
}

func MapUser(u domain.User) User {
	resp := User{
		ID:         u.ID,
		Name:       u.Name,
		Balance:    u.Balance,
		IsActive:   u.IsActive,
		IsDisabled: u.IsDisabled,
		Created:    u.Created,
		Updated:    u.Updated,
	}
	if u.Email != nil {
		resp.Email = *u.Email
	}
	return resp
}

func NewUserResponse(u domain.User) UserResponse {
	return UserResponse{User: MapUser(u)}
}

type UserResponse struct {
	User User `json:"user"`
}

func (ur UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Error struct {
	Class   string `json:"class"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrResponse struct {
	Error Error `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Error.Code)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Error: Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		},
	}
}

func ErrUserNotFound(err domain.UserNotFoundError) render.Renderer {
	return &ErrResponse{
		Error: Error{
			Class:   "App\\Exception\\UserNotFoundException",
			Code:    http.StatusNotFound,
			Message: err.Error(),
		},
	}
}

func ErrServerError(err error) render.Renderer {
	return &ErrResponse{
		Error: Error{
			Message: "Internal Server Error",
			Code:    http.StatusInternalServerError,
		},
	}
}

type UserRequest struct {
	NameParam  string  `json:"name"`
	EmailParam *string `json:"email"`
}

func (u UserRequest) Bind(r *http.Request) error {
	if u.NameParam == "" {
		return errors.New("missing required User fields.")
	}
	return nil
}

func (u UserRequest) Name() string {
	return u.NameParam
}

func (u UserRequest) Email() string {
	if u.HasEmail() {
		return *u.EmailParam
	}
	return ""
}

func (u UserRequest) HasEmail() bool {
	return u.EmailParam != nil
}
