package cache_test

import (
	"context"
	"errors"
	"testing"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/pkg/cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResilientCache_Get(t *testing.T) {
	// Simulate a temporary error
	tempErr := errors.New("network flake")

	tests := []struct {
		name        string
		key         string
		mockSetup   func(mockCache *mocks.MockCache)
		wantVal     string
		wantErr     bool
		errContains string
	}{
		{
			name: "Happy Path",
			key:  "key1",
			mockSetup: func(mockCache *mocks.MockCache) {
				// Expect exactly 1 call
				mockCache.EXPECT().Get(gomock.Any(), "key1").Return("value1", nil).Times(1)
			},
			wantVal: "value1",
			wantErr: false,
		},
		{
			name: "Retry Success",
			key:  "key2",
			mockSetup: func(mockCache *mocks.MockCache) {
				// Fail twice, then succeed
				gomock.InOrder(
					mockCache.EXPECT().Get(gomock.Any(), "key2").Return("", tempErr),
					mockCache.EXPECT().Get(gomock.Any(), "key2").Return("", tempErr),
					mockCache.EXPECT().Get(gomock.Any(), "key2").Return("value2", nil),
				)
			},
			wantVal: "value2",
			wantErr: false,
		},
		{
			name: "Max Retries Exceeded",
			key:  "key3",
			mockSetup: func(mockCache *mocks.MockCache) {
				// Fail 3 times (MaxRetries)
				mockCache.EXPECT().Get(gomock.Any(), "key3").Return("", tempErr).Times(3)
			},
			wantVal:     "",
			wantErr:     true,
			errContains: "max retries exceeded", 
		},
		{
			name: "No Retry on Cache Miss",
			key:  "key4",
			mockSetup: func(mockCache *mocks.MockCache) {
				// Return empty string and nil error (Cache Miss)
				// Should be called ONLY ONCE
				mockCache.EXPECT().Get(gomock.Any(), "key4").Return("", nil).Times(1)
			},
			wantVal: "",
			wantErr: false,
		},
		{
			name: "Context Canceled (No Retry)",
			key:  "key5",
			mockSetup: func(mockCache *mocks.MockCache) {
				// Should not retry if context is canceled
				mockCache.EXPECT().Get(gomock.Any(), "key5").Return("", context.Canceled).Times(1)
			},
			wantVal:     "",
			wantErr:     true,
			errContains: "context canceled", 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCache := mocks.NewMockCache(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockCache)
			}

			// Initialize ResilientCache
			resilient := cache.NewResilientCache(mockCache)

			ctx := context.Background()
			val, err := resilient.Get(ctx, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, val)
			}
		})
	}
}