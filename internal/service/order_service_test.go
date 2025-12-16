package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/shopspring/decimal" // Import decimal
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOrderService_CreateOrder(t *testing.T) {
	type fields struct {
		mockSetup func(
			mockOrderRepo *mocks.MockOrderRepository,
			mockProductRepo *mocks.MockProductRepository,
			mockTxManager *mocks.MockTransactionManager,
			req *service.OrderCreateReq,
		)
	}
	type args struct {
		req *service.OrderCreateReq
	}
	tests := []struct {
		name      string
		args      args
		fields    fields
		wantErr   bool
		errStr    string
		wantResp  bool
		checkResp func(t *testing.T, resp *service.OrderCreateResp)
	}{
		{
			name: "Success",
			args: args{
				req: &service.OrderCreateReq{
					UserID: 1, // Changed to uint64 in service DTO, but mock setup uses int literal
					Items: []service.OrderItemReq{
						{SKUID: 101, Quantity: 2}, // Changed to uint64 in service DTO
					},
				},
			},
			fields: fields{
				mockSetup: func(mockOrderRepo *mocks.MockOrderRepository, mockProductRepo *mocks.MockProductRepository, mockTxManager *mocks.MockTransactionManager, req *service.OrderCreateReq) {
					// 1. GetSKUByID (Check Price & Stock)
					mockProductRepo.EXPECT().GetSKUByID(gomock.Any(), uint64(101)).Return(&model.SKU{ // Changed to uint64
						Price: decimal.NewFromFloat(50.0), // Changed to decimal.Decimal
						Stock: 100,
					}, nil)

					// 2. Transaction Setup
					mockTxManager.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						// Execute the callback immediately to simulate transaction
						return fn(ctx)
					})

					// 3. UpdateSKUStock (Deduct)
					mockProductRepo.EXPECT().UpdateSKUStock(gomock.Any(), uint64(101), -2).Return(nil) // Changed to uint64

					// 4. CreateOrder
					mockOrderRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				},
			},
			wantErr:  false,
			wantResp: true,
			checkResp: func(t *testing.T, resp *service.OrderCreateResp) {
				assert.True(t, decimal.NewFromFloat(100.0).Equal(resp.TotalAmount)) // Changed to decimal.Decimal
				assert.NotEmpty(t, resp.OrderNumber)
			},
		},
		{
			name: "SKUNotFound",
			args: args{
				req: &service.OrderCreateReq{
					UserID: 1,
					Items: []service.OrderItemReq{
						{SKUID: 999, Quantity: 1},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockOrderRepo *mocks.MockOrderRepository, mockProductRepo *mocks.MockProductRepository, mockTxManager *mocks.MockTransactionManager, req *service.OrderCreateReq) {
					mockProductRepo.EXPECT().GetSKUByID(gomock.Any(), uint64(999)).Return(nil, errors.New("sku not found")) // Changed to uint64
				},
			},
			wantErr: true,
			errStr:  "failed to get SKU 999",
		},
		{
			name: "InsufficientStock",
			args: args{
				req: &service.OrderCreateReq{
					UserID: 1,
					Items: []service.OrderItemReq{
						{SKUID: 101, Quantity: 10},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockOrderRepo *mocks.MockOrderRepository, mockProductRepo *mocks.MockProductRepository, mockTxManager *mocks.MockTransactionManager, req *service.OrderCreateReq) {
					mockProductRepo.EXPECT().GetSKUByID(gomock.Any(), uint64(101)).Return(&model.SKU{ // Changed to uint64
						Price: decimal.NewFromFloat(50.0), // Changed to decimal.Decimal
						Stock: 5,                          // Less than 10
					}, nil)
				},
			},
			wantErr: true,
			errStr:  "not enough stock",
		},
		{
			name: "StockDeductionFailure",
			args: args{
				req: &service.OrderCreateReq{
					UserID: 1,
					Items: []service.OrderItemReq{
						{SKUID: 101, Quantity: 1},
					},
				},
			},
			fields: fields{
				mockSetup: func(mockOrderRepo *mocks.MockOrderRepository, mockProductRepo *mocks.MockProductRepository, mockTxManager *mocks.MockTransactionManager, req *service.OrderCreateReq) {
					mockProductRepo.EXPECT().GetSKUByID(gomock.Any(), uint64(101)).Return(&model.SKU{ // Changed to uint64
						Price: decimal.NewFromFloat(50.0), // Changed to decimal.Decimal
						Stock: 10,
					}, nil)

					mockTxManager.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

					mockProductRepo.EXPECT().UpdateSKUStock(gomock.Any(), uint64(101), -1).Return(errors.New("db lock error")) // Changed to uint64
				},
			},
			wantErr: true,
			errStr:  "failed to deduct stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
			mockProductRepo := mocks.NewMockProductRepository(ctrl)
			mockTxManager := mocks.NewMockTransactionManager(ctrl)

			orderService := service.NewOrderService(mockOrderRepo, mockProductRepo, mockTxManager)
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockOrderRepo, mockProductRepo, mockTxManager, tt.args.req)
			}

			resp, err := orderService.CreateOrder(ctx, tt.args.req)
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