package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Test structures to ensure correct JSON marshalling for the handler
type TestSKURequest struct {
	Attributes map[string]interface{} `json:"attributes"`
	Price      float64                `json:"price"`
	Stock      int                    `json:"stock"`
	Image      string                 `json:"image"`
}

type TestCreateProductRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CategoryID  uint64           `json:"category_id"`
	SKUs        []TestSKURequest `json:"skus"`
}

func TestProductHandler_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productName := utils.RandomString(10)
	skuAttrs := map[string]interface{}{"color": "red", "size": "M"}

	type fields struct {
		mockSetup func(mockService *mocks.MockProductService)
	}
	type args struct {
		reqBody interface{}
	}
	tests := []struct {
		name          string
		args          args
		fields        fields
		wantStatus    int
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			args: args{
				reqBody: TestCreateProductRequest{
					Name:        productName,
					Description: "Test Description",
					CategoryID:  1,
					SKUs: []TestSKURequest{
						{Attributes: skuAttrs, Price: 100.0, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					// Use AssignableToTypeOf + DoAndReturn for strict argument validation
					mockService.EXPECT().
						CreateProduct(gomock.Any(), gomock.AssignableToTypeOf(&service.ProductCreateReq{})).
						DoAndReturn(func(_ context.Context, req *service.ProductCreateReq) (*service.ProductCreateResp, error) {
							assert.Equal(t, productName, req.Name)
							assert.Equal(t, uint64(1), req.CategoryID)
							assert.Len(t, req.SKUs, 1)

							// Verify Price conversion (float64 -> decimal.Decimal)
							assert.True(t, decimal.NewFromFloat(100.0).Equal(req.SKUs[0].Price), "Price mismatch")

							// Verify Attributes (Map -> RawMessage)
							var receivedAttrs map[string]interface{}
							err := json.Unmarshal(req.SKUs[0].Attributes, &receivedAttrs)
							require.NoError(t, err)
							assert.Equal(t, skuAttrs, receivedAttrs)

							return &service.ProductCreateResp{SPUID: 101}, nil
						})
				},
			},
			wantStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, float64(http.StatusCreated), resp["code"])
				assert.Equal(t, "Product created successfully", resp["message"])

				data := resp["data"].(map[string]interface{})
				// SPUID is returned as string due to `json:",string"` tag on uint64
				assert.Equal(t, "101", data["spu_id"])
			},
		},
		{
			name: "InvalidInput_MissingName",
			args: args{
				reqBody: TestCreateProductRequest{
					CategoryID: 1,
					SKUs: []TestSKURequest{
						{Attributes: skuAttrs, Price: 100.0, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					// Expect NO call to service
				},
			},
			wantStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, float64(http.StatusBadRequest), resp["code"])
				assert.Contains(t, resp["message"], "Field validation for 'Name' failed")
			},
		},
		{
			name: "ServiceError",
			args: args{
				reqBody: TestCreateProductRequest{
					Name:       productName,
					CategoryID: 1,
					SKUs: []TestSKURequest{
						{Attributes: skuAttrs, Price: 100.0, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockProductService) {
					mockService.EXPECT().
						CreateProduct(gomock.Any(), gomock.AssignableToTypeOf(&service.ProductCreateReq{})).
						Return(nil, errors.New("db error"))
				},
			},
			wantStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, float64(http.StatusInternalServerError), resp["code"])
				// Security check: Ensure internal error details are NOT leaked
				assert.Equal(t, "Internal Server Error", resp["message"])
			},
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
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
