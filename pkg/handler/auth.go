package handler

import (
	"net/http"

	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) {
	var input rest.User

	if err := c.BindJSON(&input); err != nil {
		handleError(c, rest.NewInvalidRequestError(err))
		return
	}

	_, err := h.services.Autorization.CreateUser(input)
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
	var input rest.User

	if err := c.BindJSON(&input); err != nil {
		handleError(c, rest.NewInvalidRequestError(err))
		return
	}

	tokens, err := h.services.Autorization.GenerateTokens(input.Email, input.Password)
	if err != nil {
		handleError(c, err)
		return
	}


	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) updateToken(c *gin.Context) {
	var input rest.ResponseTokens

	if err := c.BindJSON(&input); err != nil {
		handleError(c, rest.NewInvalidRequestError(err))
		return
	}

	tokens, err := h.services.GetAccessToken(input.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}
