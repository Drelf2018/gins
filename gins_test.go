package gins_test

import (
	"net/http"
	"testing"

	"github.com/Drelf2018/gins"
	"github.com/gin-gonic/gin"
)

type Auth int

func (Auth) Use(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		uid = "visitor"
	}
	c.Set("uid", uid)
}

func (Auth) GetPing(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "hello " + c.GetString("uid")})
}

type Admin bool

func (*Admin) Use(c *gin.Context) {
	if c.GetString("uid") != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failure."})
		c.Abort()
	}
}

func (*Admin) GetData(c *gin.Context) {
	c.JSON(200, gin.H{"data": "some important data."})
}

func (*Admin) StaticFileCode() (string, string) {
	return "/code", "./gins.go"
}

type Main struct {
	Auth
	*Admin `router:"admin"`
}

func (Main) StaticFileIcon() (string, string) {
	return "favicon.ico", "./favicon.ico"
}

func TestGin(t *testing.T) {
	r, err := gins.Default(Main{})
	if err != nil {
		t.Fatal(err)
	}
	r.Run("localhost:9000")
}
