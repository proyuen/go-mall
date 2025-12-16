package model

import (
	"github.com/proyuen/go-mall/pkg/snowflake"
	"gorm.io/gorm"
)

// BeforeCreate is a GORM hook that runs before inserting a new record.
// It generates a distributed ID using Snowflake algorithm if one is not provided.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == 0 {
		b.ID = snowflake.GenID()
	}
	return nil
}