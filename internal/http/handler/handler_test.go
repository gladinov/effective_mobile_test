package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	handlermocks "github.com/gladinov/effective_mobile_test_assignment/internal/http/handler/mocks"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		body := `{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025","end_date":"09-2025"}`
		req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		wantSub := domain.Subscription{
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
			EndDate:     yearMonthPtrHandler(2025, time.September),
		}
		createdID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")

		service.On("Create", mock.Anything, wantSub).Return(createdID, nil).Once()

		err := h.Create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp CreateResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, createdID, resp.SubID)
	})

	t.Run("invalid body returns bad request", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		body := `{"service_name":"Yandex Plus","price":400,"user_id":"","start_date":"07-2025"}`
		req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.Create(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusBadRequest, httpErr.Code)
		require.Equal(t, "user_id must not be empty", httpErr.Message)
	})

	t.Run("service error returns internal error", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		body := `{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}`
		req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		wantSub := domain.Subscription{
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
			EndDate:     nil,
		}
		service.On("Create", mock.Anything, wantSub).Return(uuid.UUID{}, assertErr()).Once()

		err := h.Create(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func TestHandlerGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("GetByID", mock.Anything, subID).Return(domain.Subscription{
			ID:          subID,
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
			EndDate:     yearMonthPtrHandler(2025, time.September),
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Get(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), `"start_date":"07-2025"`)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("GetByID", mock.Anything, subID).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Get(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/not-uuid", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("not-uuid")

		err := h.Get(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("GetByID", mock.Anything, subID).Return(domain.Subscription{}, assertErr()).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Get(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func TestHandlerUpdate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		body := `{"service_name":"Netflix","price":1200,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026","end_date":"03-2026"}`
		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+subID.String(), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		wantSub := domain.Subscription{
			ServiceName: "Netflix",
			Price:       1200,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
			EndDate:     yearMonthPtrHandler(2026, time.March),
		}
		service.On("UpdateByID", mock.Anything, subID, wantSub).Return(nil).Once()

		err := h.Update(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("invalid body returns bad request", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		body := `{"service_name":"Netflix","price":1200,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"03-2026","end_date":"01-2026"}`
		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+subID.String(), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Update(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusBadRequest, httpErr.Code)
		require.Equal(t, "end_date must not be before start_date", httpErr.Message)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		body := `{"service_name":"Netflix","price":1200,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026"}`
		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+subID.String(), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		wantSub := domain.Subscription{
			ServiceName: "Netflix",
			Price:       1200,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
			EndDate:     nil,
		}
		service.On("UpdateByID", mock.Anything, subID, wantSub).Return(domain.ErrSubscriptionNotFound).Once()

		err := h.Update(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		body := `{"service_name":"Netflix","price":1200,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026"}`
		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+subID.String(), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		wantSub := domain.Subscription{
			ServiceName: "Netflix",
			Price:       1200,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
			EndDate:     nil,
		}
		service.On("UpdateByID", mock.Anything, subID, wantSub).Return(assertErr()).Once()

		err := h.Update(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func TestHandlerDelete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("DeleteByID", mock.Anything, subID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Delete(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("DeleteByID", mock.Anything, subID).Return(domain.ErrSubscriptionNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Delete(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/not-uuid", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("not-uuid")

		err := h.Delete(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
		service.On("DeleteByID", mock.Anything, subID).Return(assertErr()).Once()

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+subID.String(), nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(subID.String())

		err := h.Delete(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func TestHandlerList(t *testing.T) {
	t.Parallel()

	t.Run("returns empty list with 200", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)
		service.On("List", mock.Anything).Return([]domain.Subscription{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.List(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[]`, rec.Body.String())
	})

	t.Run("service error returns 500", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)
		service.On("List", mock.Anything).Return(([]domain.Subscription)(nil), assertErr()).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.List(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func TestHandlerTotal(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
		filter := domain.FilterTotal{
			UserID:      &userID,
			ServiceName: stringPtrHandler("Yandex Plus"),
			From:        yearMonthPtrHandler(2025, time.July),
			To:          yearMonthPtrHandler(2025, time.September),
		}
		service.On("GetTotal", mock.Anything, filter).Return(1200, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?user_id="+userID.String()+"&service_name="+url.QueryEscape("Yandex Plus")+"&from=07-2025&to=09-2025", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.Total(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `{"total":1200}`, rec.Body.String())
	})

	t.Run("invalid query returns bad request", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=13-2025", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.Total(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusBadRequest, httpErr.Code)
		require.Equal(t, "month must be between 1 and 12", httpErr.Message)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		t.Parallel()

		service := handlermocks.NewService(t)
		h := newTestHandler(service)

		userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
		filter := domain.FilterTotal{
			UserID:      &userID,
			ServiceName: stringPtrHandler("Yandex Plus"),
			From:        yearMonthPtrHandler(2025, time.July),
			To:          yearMonthPtrHandler(2025, time.September),
		}
		service.On("GetTotal", mock.Anything, filter).Return(0, assertErr()).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?user_id="+userID.String()+"&service_name="+url.QueryEscape("Yandex Plus")+"&from=07-2025&to=09-2025", nil)
		rec := httptest.NewRecorder()
		c := newTestEcho().NewContext(req, rec)

		err := h.Total(c)
		httpErr := requireHTTPError(t, err)
		require.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}

func newTestHandler(service Service) *handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(logger, service, time.Second)
}

func newTestEcho() *echo.Echo {
	e := echo.New()
	return e
}

func requireHTTPError(t *testing.T, err error) *echo.HTTPError {
	t.Helper()
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	return httpErr
}

func yearMonthPtrHandler(year int, month time.Month) *domain.YearMonth {
	return &domain.YearMonth{Year: year, Month: month}
}

func stringPtrHandler(v string) *string {
	return &v
}

func assertErr() error {
	return context.DeadlineExceeded
}
