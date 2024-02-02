package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/users"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

type ParameterMissingError struct {
	Name string
}

func (err ParameterMissingError) Error() string {
	return "" // we don't really need the Error function.
}

type ParameterInvalidError struct {
	Name string
}

func (err ParameterInvalidError) Error() string {
	return "" // we don't really need the Error function.
}

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
		_ = render.Render(w, r, ErrInvalidParamter("id"))
		return
	}

	user, err := h.svc.FindById(id)
	if err != nil {
		h.renderError(w, r, err)
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
			_ = render.Render(w, r, ErrInvalidParamter("active"))
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
	uReq := new(UserRequest)
	if err := render.Bind(r, uReq); err != nil {
		h.renderError(w, r, err)
		return
	}

	created, err := h.svc.CreateUser(uReq)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	render.JSON(w, r, NewUserResponse(created))
}

func (h Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("id"))
		return
	}

	uReq := new(UserRequest)
	if err := render.Bind(r, uReq); err != nil {
		h.renderError(w, r, err)
		return
	}

	updated, err := h.svc.UpdateUser(id, uReq)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	render.JSON(w, r, NewUserResponse(updated))
}

func (h Handler) renderError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		unfErr domain.UserNotFoundError
		aeErr  domain.UserAlreadyExistsError
		pmErr  ParameterMissingError
		piErr  ParameterInvalidError
	)

	switch {
	case errors.As(err, &unfErr):
		_ = render.Render(w, r, ErrUserNotFound(unfErr))
	case errors.As(err, &aeErr):
		_ = render.Render(w, r, ErrUserAlreadyExists(aeErr))
	case errors.As(err, &pmErr):
		_ = render.Render(w, r, ErrMissingParameter(pmErr.Name))
	case errors.As(err, &piErr):
		_ = render.Render(w, r, ErrInvalidParamter(piErr.Name))
	default:
		_ = render.Render(w, r, ErrServerError(err))
	}
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
	StatusCode int   `json:"-"`
	Error      Error `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

func ErrInvalidParamter(name string) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusBadRequest,
		Error: Error{
			Code:    http.StatusBadRequest,
			Class:   "App\\Exception\\ParameterInvalidException",
			Message: fmt.Sprintf("Parameter '%s' is invalid", name),
		},
	}
}

func ErrMissingParameter(name string) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusBadRequest,
		Error: Error{
			Code:    http.StatusBadRequest,
			Class:   "App\\Exception\\ParameterMissingException",
			Message: fmt.Sprintf("Parameter '%s' is missing", name),
		},
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusBadRequest,
		Error: Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		},
	}
}

func ErrUserNotFound(err domain.UserNotFoundError) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusNotFound,
		Error: Error{
			Class:   "App\\Exception\\UserNotFoundException",
			Code:    http.StatusNotFound,
			Message: err.Error(),
		},
	}
}

func ErrUserAlreadyExists(err domain.UserAlreadyExistsError) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusInternalServerError,
		Error: Error{
			Class:   "App\\Exception\\UserAlreadyExistsException",
			Code:    209,
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
	NameParam       *string `json:"name"`
	EmailParam      *string `json:"email"`
	IsDisabledParam *bool   `json:"isDisabled"`
}

func (u *UserRequest) Bind(r *http.Request) error {
	if u.NameParam == nil {
		return ParameterMissingError{Name: "name"}
	}
	// Trim and remove non-printable characters
	*u.NameParam = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, strings.TrimSpace(*u.NameParam))

	if *u.NameParam == "" || len(*u.NameParam) > 64 {
		return ParameterInvalidError{Name: "name"}
	}

	if u.EmailParam != nil {
		if _, err := mail.ParseAddress(*u.EmailParam); err != nil {
			return ParameterInvalidError{Name: "email"}
		}
	}

	return nil
}

func (u UserRequest) Name() string {
	return *u.NameParam
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

func (u UserRequest) IsDisabled() bool {
	return u.IsDisabledParam != nil && *u.IsDisabledParam
}
