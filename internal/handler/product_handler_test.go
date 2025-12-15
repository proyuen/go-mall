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
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Random data
	productName := utils.RandomString(10)

	type fields struct {
		mockSetup func(mockService *mocks.MockProductService)
	}
	type args struct {
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
				reqBody: CreateProductRequest{
					Name:        productName,
					Description: "Desc",
					CategoryID:  1,
					SKUs: []SKURequest{
						{Attributes: "{}", Price: 100, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					mockService.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(&service.ProductCreateResp{
						SPUID: 1,
					}, nil)
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   "Product created successfully",
		},
		{
			name: "InvalidInput_MissingName",
			args: args{
				reqBody: CreateProductRequest{
					CategoryID: 1,
					SKUs: []SKURequest{
						{Attributes: "{}", Price: 100, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					// Expect no call
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name: "InvalidInput_InvalidSKU",
			args: args{
				reqBody: CreateProductRequest{
					Name:       productName,
					CategoryID: 1,
					SKUs: []SKURequest{
						{Attributes: "{}", Price: -10, Stock: 10}, // Invalid price
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					// Expect no call
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation for 'Price' failed", // Exact message depends on validator
		},
		{
			name: "ServiceError",
			args: args{
				reqBody: CreateProductRequest{
					Name:       productName,
					CategoryID: 1,
					SKUs: []SKURequest{
						{Attributes: "{}", Price: 100, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					mockService.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(nil, errors.New("service failure"))
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "service failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockProductService(ctrl)
			handler := NewProductHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, err := json.Marshal(tt.args.reqBody)
			require.NoError(t, err)

			c.Request, err = http.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockService)
			}

			handler.CreateProduct(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
