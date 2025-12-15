package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductService_CreateProduct(t *testing.T) {
	// Random data
	productName := utils.RandomString(10)
	skuAttr := `{"color": "red"}`

	type fields struct {
		mockSetup func(mockRepo *mocks.MockProductRepository, req *service.ProductCreateReq)
	}
	type args struct {
		req *service.ProductCreateReq
	}
	tests := []struct {
		name      string
		args      args
		fields    fields
		wantErr   bool
		errStr    string
		wantResp  bool
		checkResp func(t *testing.T, resp *service.ProductCreateResp)
	}{
		{
			name: "Success",
			args: args{
				req: &service.ProductCreateReq{
					Name:        productName,
					Description: "Test Description",
					CategoryID:  1,
					SKUs: []service.SKUCreateReq{
						{Attributes: skuAttr, Price: 100, Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockProductRepository, req *service.ProductCreateReq) {
					// Expect CreateSPU to be called ONCE with SPU containing nested SKUs
					mockRepo.EXPECT().CreateSPU(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, spu *model.SPU) error {
						spu.ID = 1 // Simulate DB ID generation
						assert.Equal(t, req.Name, spu.Name)
						
						// Verify nested SKUs
						require.Len(t, spu.SKUs, 1)
						assert.Equal(t, skuAttr, spu.SKUs[0].Attributes)
						assert.Equal(t, float64(100), spu.SKUs[0].Price)
						return nil
					})
				},
			},
			wantErr:  false,
			wantResp: true,
			checkResp: func(t *testing.T, resp *service.ProductCreateResp) {
				assert.Equal(t, uint(1), resp.SPUID)
			},
		},
		{
			name: "CreationError",
			args: args{
				req: &service.ProductCreateReq{
					Name: productName,
					SKUs: []service.SKUCreateReq{
						{Attributes: skuAttr},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockProductRepository, req *service.ProductCreateReq) {
					mockRepo.EXPECT().CreateSPU(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
				},
			},
			wantErr: true,
			errStr:  "failed to create product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockProductRepository(ctrl)
			productService := service.NewProductService(mockRepo)
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, tt.args.req)
			}

			resp, err := productService.CreateProduct(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errStr != "" {
					assert.Contains(t, err.Error(), tt.errStr)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.wantResp {
				require.NotNil(t, resp)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			} else {
				require.Nil(t, resp)
			}
		})
	}
}