package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductService_CreateProduct(t *testing.T) {
	productName := utils.RandomString(10)
	skuAttr := `{"color": "red"}`

	type fields struct {
		mockSetup func(mockRepo *mocks.MockProductRepository, mockCache *mocks.MockCache, req *service.ProductCreateReq)
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
						{Attributes: json.RawMessage(skuAttr), Price: decimal.NewFromInt(100), Stock: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockProductRepository, mockCache *mocks.MockCache, req *service.ProductCreateReq) {
					mockRepo.EXPECT().CreateSPU(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, spu *model.SPU) error {
						spu.ID = 101
						assert.Equal(t, req.Name, spu.Name)
						require.Len(t, spu.SKUs, 1)
						assert.Equal(t, model.JSONB{"color": "red"}, spu.SKUs[0].Attributes)
						assert.True(t, decimal.NewFromInt(100).Equal(spu.SKUs[0].Price))
						assert.Equal(t, 10, spu.SKUs[0].Stock)
						return nil
					})
				},
			},
			wantErr:  false,
			wantResp: true,
			checkResp: func(t *testing.T, resp *service.ProductCreateResp) {
				assert.Equal(t, uint64(101), resp.SPUID)
			},
		},
		{
			name: "CreationError",
			args: args{
				req: &service.ProductCreateReq{
					Name: productName,
					SKUs: []service.SKUCreateReq{
						{Attributes: json.RawMessage(skuAttr)},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockProductRepository, mockCache *mocks.MockCache, req *service.ProductCreateReq) {
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
			mockCache := mocks.NewMockCache(ctrl)
			productService := service.NewProductService(mockRepo, mockCache)
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, mockCache, tt.args.req)
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

func TestProductService_GetProduct(t *testing.T) {
	spuID := uint64(101)
	cacheKey := fmt.Sprintf("mall:product:spu:%d", spuID)

	t.Run("CacheHit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockProductRepository(ctrl)
		mockCache := mocks.NewMockCache(ctrl)
		productService := service.NewProductService(mockRepo, mockCache)
		ctx := context.Background()

		cachedResp := &service.ProductResp{ID: spuID, Name: "Cached Product"}
		bytes, _ := json.Marshal(cachedResp)

		mockCache.EXPECT().Get(ctx, cacheKey).Return(string(bytes), nil)
		// Repo should NOT be called

		resp, err := productService.GetProduct(ctx, spuID)
		require.NoError(t, err)
		assert.Equal(t, "Cached Product", resp.Name)
	})

	t.Run("CacheMiss_DBHit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockProductRepository(ctrl)
		mockCache := mocks.NewMockCache(ctrl)
		productService := service.NewProductService(mockRepo, mockCache)
		ctx := context.Background()

		mockCache.EXPECT().Get(ctx, cacheKey).Return("", nil) // Cache miss
		mockRepo.EXPECT().GetSPUByID(ctx, spuID).Return(&model.SPU{Base: model.Base{ID: spuID}, Name: "DB Product"}, nil)
		// Expect Set Cache
		mockCache.EXPECT().Set(ctx, cacheKey, gomock.Any(), time.Hour).Return(nil)

		resp, err := productService.GetProduct(ctx, spuID)
		require.NoError(t, err)
		assert.Equal(t, "DB Product", resp.Name)
	})
}