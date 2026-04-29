package handler

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGetQueryForTotal(t *testing.T) {
	t.Parallel()

	t.Run("parses domain filter", func(t *testing.T) {
		t.Parallel()

		userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
		req := httptest.NewRequest("GET", "/subscriptions/total?user_id="+userID.String()+"&service_name="+url.QueryEscape("Yandex Plus")+"&from=07-2025&to=09-2025", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		got, err := getQueryForTotal(c)
		require.NoError(t, err)
		require.NotNil(t, got.UserID)
		require.Equal(t, userID, *got.UserID)
		require.NotNil(t, got.ServiceName)
		require.Equal(t, "Yandex Plus", *got.ServiceName)
		require.Equal(t, 2025, got.From.Year)
		require.Equal(t, time.July, got.From.Month)
		require.Equal(t, 2025, got.To.Year)
		require.Equal(t, time.September, got.To.Month)
	})

	t.Run("returns error when to is before from", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total?from=09-2025&to=07-2025", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := getQueryForTotal(c)
		require.ErrorIs(t, err, domain.ErrEndDateBeforeStart)
	})

	t.Run("returns error for invalid user id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total?user_id=not-uuid", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := getQueryForTotal(c)
		require.ErrorIs(t, err, errInvalidUUID)
	})
}

func TestGetPaginationQuery(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults when query params are absent", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		got, err := getPaginationQuery(c)
		require.NoError(t, err)
		require.Equal(t, defaultLimit, got.Limit)
		require.Equal(t, defaultOffset, got.Offset)
	})

	t.Run("parses limit and offset", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions?limit=25&offset=50", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		got, err := getPaginationQuery(c)
		require.NoError(t, err)
		require.Equal(t, uint64(25), got.Limit)
		require.Equal(t, uint64(50), got.Offset)
	})

	t.Run("returns error when limit is zero", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions?limit=0", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := getPaginationQuery(c)
		require.ErrorIs(t, err, errLimitInvalid)
	})

	t.Run("returns error when offset is negative", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions?offset=-1", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := getPaginationQuery(c)
		require.ErrorIs(t, err, errOffsetInvalid)
	})
}

func TestStringFromQueryParam(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when param is absent", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		got, err := stringFromQueryParam("user_id", c, errUserIDEmpty, errUserIDMultipleValues)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("returns trimmed value", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total?service_name=%20Yandex%20Plus%20", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		got, err := stringFromQueryParam("service_name", c, errServiceNameEmpty, errServiceNameMultipleValues)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "Yandex Plus", *got)
	})

	t.Run("returns multiple values error", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total?user_id=1&user_id=2", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := stringFromQueryParam("user_id", c, errUserIDEmpty, errUserIDMultipleValues)
		require.ErrorIs(t, err, errUserIDMultipleValues)
	})

	t.Run("returns empty error for blank value", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/subscriptions/total?service_name=%20", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())

		_, err := stringFromQueryParam("service_name", c, errServiceNameEmpty, errServiceNameMultipleValues)
		require.ErrorIs(t, err, errServiceNameEmpty)
	})
}
