package handler

import (
	"net/http"

	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) {
	var input rest.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.services.Autorization.CreateUser(input)
	if err != nil {
		//if errors.Is(err, rest.ErrUserAlreadyExists) {
		//}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	tokens, err := h.services.GenerateTokens(input.Email, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"accessToken": tokens.AccessToken, "refreshToken": tokens.RefreshToken})
}

func (h *Handler) signIn(c *gin.Context) {
	var input rest.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.services.Autorization.GenerateTokens(input.Email, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"accessToken": tokens.AccessToken, "refreshToken": tokens.RefreshToken})
}
