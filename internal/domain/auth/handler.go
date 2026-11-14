package auth

import (
	"context"
	jwt "rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/transport/http/respond"
	"rcon-scheduler-panel/internal/validation"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Register(ctx context.Context, req request.AuthCredentials) (string, error)
	Login(ctx context.Context, req request.AuthCredentials) (string, error)
}

type Handler struct {
	service Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Login godoc
//
//	@Summary		Login
//	@Description	Validates username and password, then returns a JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		request.AuthCredentials	true	"Username and password"
//	@Success		200			{object}	response.Token
//	@Failure		400			{object}	response.Problem	"Validation error"
//	@Failure		401			{object}	response.Problem	"Invalid credentials"
//	@Failure		500			{object}	response.Problem	"Internal server error"
//	@Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req request.AuthCredentials
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	token, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, response.NewBearerToken(token, jwt.TokenTTL))
}

// Register godoc
//
//	@Summary		Register
//	@Description	Creates a new user and returns a JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		request.AuthCredentials	true	"Username and password"
//	@Success		201			{object}	response.Token
//	@Failure		400			{object}	response.Problem	"Validation error"
//	@Failure		403			{object}	response.Problem	"Registration is closed"
//	@Failure		500			{object}	response.Problem	"Internal server error"
//	@Router			/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req request.AuthCredentials
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	token, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.Created(c, response.NewBearerToken(token, jwt.TokenTTL))
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/register", h.Register)
}
