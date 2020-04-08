package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

const ServerError500ResponseString = "server failed to complete request. please try after some time."
const UnauthorisedError401ResponseString = "The authorization credentials provided for the request are invalid."

// SendOK200Response sets a ok status (HTTP 200) on the gin context and
// send responseBody parameter passed as JSON response.
func SendOK200Response(c *gin.Context, response interface{}) {
	c.Render(http.StatusOK, render.JSON{Data: response})
	c.Abort()
}

func SendCreated201Response(c *gin.Context, response interface{}) {
	c.Render(http.StatusCreated, render.JSON{Data: response})
	c.Abort()
}

func SendNoContent204Response(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{})
	c.Abort()
}

func SendPaymentRequired402Response(c *gin.Context, response interface{}) {
	c.Render(http.StatusPaymentRequired, render.JSON{Data: response})
	c.Abort()
}

// BadRequestResponse sets a "bad request" (HTTP 400) on the gin context and
// send userErrMessage to user. Finally it also aborts any other handlers in-line by calling Abort.
func SendBadRequest400Response(c *gin.Context, userErrMessage string) {
	c.JSON(http.StatusBadRequest,
		gin.H{
			"msg": userErrMessage,
		})
	c.Abort()
}

func SendBadRequestPlatform400Response(c *gin.Context, userErrMessage string) {
	c.JSON(http.StatusBadRequest,
		gin.H{
			"res_str":  userErrMessage,
			"res_data": gin.H{},
		})
	c.Abort()
}

// SendNotFound404Response sets a "resource not found" (HTTP 404) on the gin context and
// send userErrMessage to user. Finally it also aborts any other handlers in-line by calling Abort.
func SendNotFound404Response(c *gin.Context, userErrMessage string) {
	c.JSON(http.StatusNotFound,
		gin.H{
			"msg": userErrMessage,
		})
	c.Abort()
}

// SendUnauthorised401Response sets the response status code to Unauthorised (http 401).
// Finally it also aborts any other handlers in-line by calling Abort.
func SendUnauthorised401Response(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"msg": UnauthorisedError401ResponseString,
	})
	c.Abort()
}

func SendResourceConflict409Response(c *gin.Context, userErrMessage string) {
	c.JSON(http.StatusConflict,
		gin.H{
			"msg": userErrMessage,
		})
	c.Abort()
}

// BadRequestResponse sets a "bad request" (HTTP 400) on the gin context and
// send userErrMessage to user. Finally it also aborts any other handlers in-line by calling Abort.
func SendServerError500Response(c *gin.Context) {
	c.JSON(http.StatusInternalServerError,
		gin.H{
			"msg": ServerError500ResponseString,
		})
	c.Abort()
}

func SendStatusProcessing(c *gin.Context, userErrMessage string) {
	c.JSON(http.StatusProcessing,
		gin.H{
			"msg": userErrMessage,
		})
	c.Abort()
}
