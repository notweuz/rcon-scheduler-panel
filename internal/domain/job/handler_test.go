package job_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/job"
	"rcon-scheduler-panel/internal/domain/job/mocks"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/validation"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupRouter(mockSVC *mocks.MockService) *gin.Engine {
	cfg := &config.Config{JWTSecret: "test"}
	h := job.NewHandler(mockSVC, cfg)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h.RegisterRoutes(r.Group("/api"))
	return r
}

func mustJSON(t *testing.T, v any) []byte {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func newRequest(t *testing.T, method, url string, body []byte, withToken bool) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	if withToken {
		token, err := auth.GenerateToken(1, "test")
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

var validJobReq = request.CreateJob{
	Name:       "Restart warning",
	Cron:       "0 4 * * *",
	LogResults: true,
	Steps: []request.Step{
		{Command: "say Restart in 1 minute"},
		{Wait: "1m"},
		{Command: "quit"},
	},
}

var expectedSteps = []model.Step{
	{Command: "say Restart in 1 minute"},
	{Wait: "1m"},
	{Command: "quit"},
}

func TestJobHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	const url = "/api/servers/3/jobs"

	t.Run("201 - Job Created", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(j *model.Job) bool {
				return j.ServerID == 3 &&
					j.Name == validJobReq.Name &&
					j.Cron == validJobReq.Cron &&
					j.LogResults == validJobReq.LogResults &&
					assert.ObjectsAreEqual(expectedSteps, j.Steps)
			})).
			Run(func(_ context.Context, j *model.Job) {
				j.ID = 10
			}).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), true))

		require.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool         `json:"success"`
			Data    response.Job `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(10), resp.Data.ID)
		assert.Equal(t, int64(3), resp.Data.ServerID)
		assert.Equal(t, validJobReq.Name, resp.Data.Name)
		assert.Equal(t, validJobReq.Cron, resp.Data.Cron)
		assert.True(t, resp.Data.LogResults)
		assert.Equal(t, []response.Step{
			{Command: "say Restart in 1 minute"},
			{Wait: "1m"},
			{Command: "quit"},
		}, resp.Data.Steps)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid server ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		for _, id := range []string{"abc", "0"} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, newRequest(t, http.MethodPost, "/api/servers/"+id+"/jobs", mustJSON(t, validJobReq), true))

			assert.Equal(t, http.StatusBadRequest, w.Code, "server id %q", id)
		}
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, []byte(`{"name":`), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := validJobReq
		invalid.Name = "ab"
		invalid.Cron = ""
		invalid.Steps = []request.Step{}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, invalid), true))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Fields, "name")
		assert.Contains(t, resp.Error.Fields, "cron")
		assert.Contains(t, resp.Error.Fields, "steps")
	})

	t.Run("400 - Bad Request (invalid steps)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		tests := []struct {
			name string
			step request.Step
		}{
			{"empty step", request.Step{}},
			{"command and wait together", request.Step{Command: "say hi", Wait: "1m"}},
		}

		for _, tt := range tests {
			invalid := validJobReq
			invalid.Steps = []request.Step{tt.step}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, invalid), true))

			require.Equal(t, http.StatusBadRequest, w.Code, tt.name)

			var resp response.Response
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.NotNil(t, resp.Error, tt.name)
			assert.Contains(t, resp.Error.Fields, "command", tt.name)
			assert.Contains(t, resp.Error.Fields, "wait", tt.name)
		}
	})

	t.Run("400 - Bad Request (rejected by service)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errs.ErrBadRequest.WithMessage("invalid cron expression")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), true))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Equal(t, "invalid cron expression", resp.Error.Message)
	})

	t.Run("404 - Server Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errs.ErrNotFound.WithMessage("server not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("409 - Conflict (job already exists)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errs.ErrConflict.WithMessage("job already exist")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPost, url, mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestJobHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedJob := &model.Job{
		ID:       10,
		ServerID: 3,
		Name:     "Restart warning",
		Cron:     "0 4 * * *",
		Steps:    expectedSteps,
	}

	t.Run("200 - Job Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(10)).
			Return(expectedJob, nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/jobs/10", nil, true))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool         `json:"success"`
			Data    response.Job `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, response.JobFromModel(*expectedJob), resp.Data)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/jobs/10", nil, false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/jobs/abc", nil, true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("404 - Job Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(999)).
			Return(nil, errs.ErrNotFound.WithMessage("job not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/jobs/999", nil, true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestJobHandler_GetByServerID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedJobs := []model.Job{
		{ID: 10, ServerID: 3, Name: "Restart warning", Cron: "0 4 * * *", Steps: expectedSteps},
		{ID: 11, ServerID: 3, Name: "Save world", Cron: "*/30 * * * *", Steps: []model.Step{{Command: "save-all"}}},
	}

	t.Run("200 - Jobs Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByServerID(mock.Anything, int64(3)).
			Return(expectedJobs, nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/servers/3/jobs", nil, true))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool           `json:"success"`
			Data    []response.Job `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		require.Len(t, resp.Data, len(expectedJobs))
		for i := range expectedJobs {
			assert.Equal(t, response.JobFromModel(expectedJobs[i]), resp.Data[i])
		}
	})

	t.Run("200 - No Jobs (empty array, not null)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByServerID(mock.Anything, int64(3)).
			Return([]model.Job{}, nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/servers/3/jobs", nil, true))

		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"data":[]`)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/servers/3/jobs", nil, false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid server ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/servers/abc/jobs", nil, true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByServerID(mock.Anything, int64(3)).
			Return(nil, errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodGet, "/api/servers/3/jobs", nil, true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestJobHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	const url = "/api/jobs/10"

	t.Run("200 - Job Updated", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.MatchedBy(func(j *model.Job) bool {
				return j.ID == 10 &&
					j.Name == validJobReq.Name &&
					j.Cron == validJobReq.Cron &&
					j.LogResults == validJobReq.LogResults &&
					assert.ObjectsAreEqual(expectedSteps, j.Steps)
			})).
			Run(func(_ context.Context, j *model.Job) {
				j.ServerID = 3
			}).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, url, mustJSON(t, validJobReq), true))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool         `json:"success"`
			Data    response.Job `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(10), resp.Data.ID)
		assert.Equal(t, int64(3), resp.Data.ServerID)
		assert.Equal(t, validJobReq.Name, resp.Data.Name)
		assert.Len(t, resp.Data.Steps, len(validJobReq.Steps))
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, url, mustJSON(t, validJobReq), false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, "/api/jobs/abc", mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := validJobReq
		invalid.Steps = nil

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, url, mustJSON(t, invalid), true))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Fields, "steps")
	})

	t.Run("404 - Job Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.Anything).
			Return(errs.ErrNotFound.WithMessage("job not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, url, mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.Anything).
			Return(errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodPut, url, mustJSON(t, validJobReq), true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestJobHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("204 - Job Deleted", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Delete(mock.Anything, int64(10)).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodDelete, "/api/jobs/10", nil, true))

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodDelete, "/api/jobs/10", nil, false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodDelete, "/api/jobs/abc", nil, true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("404 - Job Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Delete(mock.Anything, int64(999)).
			Return(errs.ErrNotFound.WithMessage("job not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(t, http.MethodDelete, "/api/jobs/999", nil, true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
