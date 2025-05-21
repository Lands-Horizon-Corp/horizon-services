package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Feedback struct {
		ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Email        string     `gorm:"type:varchar(255)"`
		Description  string     `gorm:"type:text"`
		FeedbackType string     `gorm:"type:varchar(50);not null;default:'general'"`
		MediaID      *uuid.UUID `gorm:"type:uuid"`
		Media        *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
	}
)
