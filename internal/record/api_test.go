package record

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccIsDoubleTrue(t *testing.T) {
	router, ms := setupRouter()

	ms.On("IsDouble", mock.AnythingOfType("*gin.Context"), UserID(1), UserID(1)).Return(true, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/1/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	ms.AssertExpectations(t)
}

func TestSuccIsDoubleFalse(t *testing.T) {
	router, ms := setupRouter()

	ms.On("IsDouble", mock.AnythingOfType("*gin.Context"), UserID(1), UserID(2)).Return(false, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/1/2", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	ms.AssertExpectations(t)
}

func TestFailIsDouble(t *testing.T) {
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/1/asdf", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/asdf/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func setupRouter() (*gin.Engine, *MockedService) {
	r := gin.Default()
	ms := new(MockedService)
	RegisterHandlers(r, ms)
	return r, ms
}
