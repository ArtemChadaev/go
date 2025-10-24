package api

import (
	"net/http"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) {
	var input models.User

	if err := c.BindJSON(&input); err != nil {
		handleError(c, models.NewInvalidRequestError(err))
		return
	}

	_, err := h.services.CreateUser(input)
	if err != nil {
		handleError(c, err)
		return
	}

	tokens, err := h.services.GenerateTokens(input.Email, input.Password)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) signIn(c *gin.Context) {
	var input models.User

	if err := c.BindJSON(&input); err != nil {
		handleError(c, models.NewInvalidRequestError(err))
		return
	}

	tokens, err := h.services.GenerateTokens(input.Email, input.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) updateToken(c *gin.Context) {
	var input models.ResponseTokens

	if err := c.BindJSON(&input); err != nil {
		handleError(c, models.NewInvalidRequestError(err))
		return
	}

	tokens, err := h.services.GetAccessToken(input.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}
