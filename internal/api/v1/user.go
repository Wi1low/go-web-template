// Package v1 contains HTTP handlers for the /api/v1 surface. Handlers bind
// input, translate service errors into the response envelope, and carry the
// swag annotations that generate the OpenAPI document.
package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Wi1low/go-web-template/internal/auth"
	"github.com/Wi1low/go-web-template/internal/middleware"
	"github.com/Wi1low/go-web-template/internal/model"
	"github.com/Wi1low/go-web-template/internal/service"
	"github.com/Wi1low/go-web-template/pkg/response"
	"github.com/gin-gonic/gin"
)

// User aliases the demo entity so OpenAPI annotations can resolve it.
type User = model.User

// UserHandler wires the user service to gin.
type UserHandler struct {
	svc    service.UserService
	parser auth.TokenParser // nil when jwt.signing-key is empty
}

// NewUserHandler builds the handler. parser may be nil, disabling the demo
// auth routes.
func NewUserHandler(svc service.UserService, parser auth.TokenParser) *UserHandler {
	return &UserHandler{svc: svc, parser: parser}
}

// createUserRequest is the POST /users payload.
type createUserRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=64" example:"alice"`
	Email string `json:"email" binding:"required,email,max=128" example:"alice@example.com"`
}

// updateUserRequest is the PUT /users/:id payload; fields are optional.
type updateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=1,max=64"`
	Email *string `json:"email" binding:"omitempty,email,max=128"`
}

type listQuery struct {
	Page int `form:"page" binding:"omitempty,gte=1" example:"1"`
	Size int `form:"size" binding:"omitempty,gte=1,lte=100" example:"20"`
}

// LoginRequest is the demo credentials payload.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"demo@example.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// Create godoc
//
//	@Summary		Create a user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createUserRequest	true	"payload"
//	@Success		200		{object}	response.Response{data=model.User}
//	@Failure		400		{object}	response.Response
//	@Failure		409		{object}	response.Response
//	@Router			/api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.Create(c.Request.Context(), service.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	switch {
	case err == nil:
		response.Success(c, u)
	case errors.Is(err, service.ErrEmailTaken):
		response.Fail(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, "")
	}
}

// Get godoc
//
//	@Summary		Fetch a user by id
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"user id"
//	@Success		200	{object}	response.Response{data=model.User}
//	@Failure		404	{object}	response.Response
//	@Router			/api/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeServiceErr(c, err)
		return
	}
	response.Success(c, u)
}

// List godoc
//
//	@Summary		List users (paginated)
//	@Tags			users
//	@Produce		json
//	@Param			page	query		int	false	"page, 1-based"	default(1)
//	@Param			size	query		int	false	"page size"		default(20)
//	@Success		200		{object}	response.Response{data=[]model.User}
//	@Router			/api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q listQuery
	_ = c.ShouldBindQuery(&q) // optional params; defaults below
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Size == 0 {
		q.Size = 20
	}
	users, total, err := h.svc.List(c.Request.Context(), q.Page, q.Size)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPaging) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "")
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	response.Success(c, users)
}

// Update godoc
//
//	@Summary		Update a user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"user id"
//	@Param			body	body		updateUserRequest	true	"fields to change"
//	@Success		200		{object}	response.Response{data=model.User}
//	@Router			/api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, service.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeServiceErr(c, err)
		return
	}
	response.Success(c, u)
}

// Delete godoc
//
//	@Summary		Delete a user
//	@Tags			users
//	@Produce		json
//	@Param			id	path	int	true	"user id"
//	@Success		200	{object}	response.Response
//	@Router			/api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// Login godoc
//
//	@Summary		Demo login (issues a JWT)
//	@Description	Replace with real credential verification. The demo accepts
//	@Description	the seeded user demo@example.com / 123456.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"credentials"
//	@Success		200		{object}	response.Response{data=object}
//	@Router			/api/v1/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// DEMO ONLY: a real project verifies passwords against a hash. Accept the
	// seeded demo credentials so the /me flow can be exercised out of the box.
	if req.Password != "123456" {
		response.Unauthorized(c, "invalid credentials")
		return
	}
	u, err := h.svc.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.Unauthorized(c, "invalid credentials")
		return
	}
	token, err := h.parser.Issue(u.ID)
	if err != nil {
		response.InternalServerError(c, "issue token")
		return
	}
	response.Success(c, gin.H{"token": token})
}

// Me godoc
//
//	@Summary		Current user (requires bearer token)
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=model.User}
//	@Router			/api/v1/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	id, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeServiceErr(c, err)
		return
	}
	response.Success(c, u)
}

func pathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "invalid id parameter")
		return 0, false
	}
	return id, true
}

func writeServiceErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		response.Fail(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, "")
	}
}
