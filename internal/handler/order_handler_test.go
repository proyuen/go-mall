package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/token"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type fields struct {
		mockSetup func(mockService *mocks.MockOrderService)
	}
	type args struct {
		userID  uint64
		reqBody interface{}
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			args: args{
				userID: 1,
				reqBody: CreateOrderRequest{
					Items: []CreateOrderItemRequest{
						{SKUID: 101, Quantity: 2},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockOrderService) {
					mockService.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(&service.OrderCreateResp{
						OrderID:     1,
						OrderNumber: "ORD123",
						TotalAmount: decimal.NewFromFloat(100.0),
					}, nil)
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   "Order created successfully",
		},
		{
			name: "InvalidInput_EmptyItems",
			args: args{
				userID: 1,
				reqBody: CreateOrderRequest{
					Items: []CreateOrderItemRequest{},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockOrderService) {
					// No call
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation", // Validator error
		},
		{
			name: "InvalidInput_InvalidQuantity",
			args: args{
				userID: 1,
				reqBody: CreateOrderRequest{
					Items: []CreateOrderItemRequest{
						{SKUID: 101, Quantity: 0},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockOrderService) {
					// No call
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation",
		},
		{
			name: "ServiceError",
			args: args{
				userID: 1,
				reqBody: CreateOrderRequest{
					Items: []CreateOrderItemRequest{
						{SKUID: 101, Quantity: 2},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockOrderService) {
					mockService.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(nil, errors.New("out of stock"))
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "out of stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockOrderService(ctrl)
			handler := NewOrderHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate AuthMiddleware setting context
			claims := &token.Payload{UserID: tt.args.userID}
			c.Set(utils.AuthorizationPayloadKey, claims)

			jsonBody, err := json.Marshal(tt.args.reqBody)
			require.NoError(t, err)

			c.Request, err = http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockService)
			}

			handler.CreateOrder(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}