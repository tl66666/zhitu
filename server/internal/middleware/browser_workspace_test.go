package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

const (
	testBrowserTokenA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testBrowserTokenB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestBrowserWorkspaceScopeIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.BrowserWorkspace{}); err != nil {
		t.Fatal(err)
	}

	workspaceID := func(accountID uint, token string) (int, uint) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(ContextUserID, accountID)
			c.Next()
		}, BrowserWorkspaceScope(db))
		var got uint
		r.GET("/", func(c *gin.Context) {
			got = c.GetUint(ContextUserID)
			c.Status(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderBrowserToken, token)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		return resp.Code, got
	}

	codeA1, idA1 := workspaceID(7, testBrowserTokenA)
	codeA2, idA2 := workspaceID(7, testBrowserTokenA)
	codeB, idB := workspaceID(7, testBrowserTokenB)
	codeOtherAccount, idOtherAccount := workspaceID(8, testBrowserTokenA)

	if codeA1 != http.StatusNoContent || codeA2 != http.StatusNoContent || codeB != http.StatusNoContent || codeOtherAccount != http.StatusNoContent {
		t.Fatalf("unexpected status codes: %d %d %d %d", codeA1, codeA2, codeB, codeOtherAccount)
	}
	if idA1 == 0 || idA1 != idA2 {
		t.Fatalf("same browser should reuse workspace: %d != %d", idA1, idA2)
	}
	if idA1 == idB || idA1 == idOtherAccount {
		t.Fatalf("different browser/account must have different workspaces: %d %d %d", idA1, idB, idOtherAccount)
	}
}

func TestBrowserWorkspaceScopeRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.BrowserWorkspace{}); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(ContextUserID, uint(7)); c.Next() }, BrowserWorkspaceScope(db))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, token := range []string{"", "short", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderBrowserToken, token)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("token %q returned %d", token, resp.Code)
		}
	}
}
