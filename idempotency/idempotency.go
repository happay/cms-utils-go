package idempotency

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

const (
	AppID      = "App-ID"
	RequestID  = "Request-ID"
)

type BaseModel struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

type Lock struct {
	BaseModel
	ReqId string `gorm:"unique_index:idx_req_id_app_id"`
	AppId string `gorm:"unique_index:idx_req_id_app_id"`
}


func CheckIdempotency(c *gin.Context, noRouteHandler string, db *gorm.DB) (bool, error) {
	var err error
	appId, reqId := c.GetHeader(AppID), c.GetHeader(RequestID)
	fmt.Println(c.HandlerName())
	fmt.Println(c.Request.Method)
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


