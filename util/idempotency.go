package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
)

const (
	AppID      = "App-ID"
	RequestID  = "Request-ID"
)


type Lock struct {
	BaseModel
	ReqId string `gorm:"unique_index:idx_req_id_app_id"`
	AppId string `gorm:"unique_index:idx_req_id_app_id"`
}


func CheckIdempotency(c *gin.Context, noRouteHandler string, db *gorm.DB) (bool, error) {
	var err error
	appId, reqId := c.GetHeader(AppID), c.GetHeader(RequestID)
	if c.HandlerName() == noRouteHandler || c.Request.Method == http.MethodGet {
		return false, err
	}
	lock := &Lock{
		ReqId: reqId,
		AppId: appId,
	}
	err = db.Create(lock).Error
	if err != nil {
		err := fmt.Errorf("error while inserting records in lock table reqId %s, app %s, err : %s", appId, reqId, err)
		fmt.Println(err)
		return false, err
	}
	return true, err
}
