package lux_test

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/davidsbond/lux"
	"github.com/stretchr/testify/assert"
)

func TestRouter_HandlesRequests(t *testing.T) {
	tt := []struct {
		Request        events.APIGatewayProxyRequest
		Handlers       map[string]lux.HandlerFunc
		ExpectedError  string
		ExpectedStatus int
	}{
		// Scenario 1: Valid GET request with correct headers.
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{"content-type": "application/json"},
			},
			Handlers:       map[string]lux.HandlerFunc{"GET": getHandler},
			ExpectedStatus: http.StatusOK,
		},
		// Scenario 2: Invalid GET request with incorrect header value.
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{"content-type": "application/xml"},
			},
			Handlers:       map[string]lux.HandlerFunc{"GET": getHandler},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "cannot determine route for request, check your HTTP method & headers are valid",
		},
		// Scenario 3: Handler does not exist
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{"content-type": "application/json"},
			},
			Handlers:       map[string]lux.HandlerFunc{},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "cannot determine route for request, check your HTTP method & headers are valid",
		},
		// Scenario 4: Invalid GET request with no headers.
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{},
			},
			Handlers:       map[string]lux.HandlerFunc{"GET": getHandler},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "cannot determine route for request, check your HTTP method & headers are valid",
		},
		// Scenario 5: Valid DELETE request with only a GET handler registered.
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "DELETE",
				Headers:    map[string]string{"content-type": "application/json"},
			},
			Handlers:       map[string]lux.HandlerFunc{"GET": getHandler},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "cannot determine route for request, check your HTTP method & headers are valid",
		},
	}

	for _, tc := range tt {
		// GIVEN that we have a router
		router := lux.NewRouter()

		// AND that router has handlers registered
		for method, handler := range tc.Handlers {
			router.Handler(method, handler).Headers("content-type", "application/json")
		}

		// WHEN we perform the request
		resp, err := router.HandleRequest(tc.Request)

		// THEN any errors should be what we expect
		if err != nil {
			assert.Equal(t, tc.ExpectedError, err.Error())
		}

		// AND the status code should be what we expect.
		assert.Equal(t, tc.ExpectedStatus, resp.StatusCode)
	}
}

func TestRouter_Recovers(t *testing.T) {
	tt := []struct {
		Request        events.APIGatewayProxyRequest
		Handlers       map[string]lux.HandlerFunc
		ExpectedError  string
		ExpectedStatus int
	}{
		{
			Request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{"content-type": "application/json"},
			},
			Handlers: map[string]lux.HandlerFunc{"GET": panicHandler},
		},
	}

	for _, tc := range tt {
		// GIVEN that we have a router with a recovery handler.
		router := lux.NewRouter().Recovery(recoverHandler)

		// AND that router has handlers registered
		for method, handler := range tc.Handlers {
			router.Handler(method, handler).Headers("content-type", "application/json")
		}

		// WHEN we perform the request that will panic
		resp, err := router.HandleRequest(tc.Request)

		// THEN any errors should be what we expect
		if err != nil {
			assert.Equal(t, tc.ExpectedError, err.Error())
		}

		// AND the status code should be what we expect.
		assert.Equal(t, tc.ExpectedStatus, resp.StatusCode)
	}
}

func getHandler(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func recoverHandler(req events.APIGatewayProxyRequest, err error) {
	// Do nothing
}

func panicHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	panic("uh oh")
}
