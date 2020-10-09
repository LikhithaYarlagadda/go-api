package models




import (
	"gorm.io/gorm"
	// "time"
)

type User struct{
	gorm.Model
	ID uint `gorm:"primaryKey"`
	Username  string
	Password  string
	Post []Post `gorm:"constraint:OnDelete:CASCADE;foreignKey:PostedBy"`
	Comment []Comment `gorm:"constraint:OnDelete:CASCADE;foreignKey:CommentedBy"`
	Reaction []Reaction `gorm:"constraint:OnDelete:CASCADE;foreignKey:ReactedBy"`

}


type Post struct {
	gorm.Model
	Content string
	PostedBy uint
	Comment []Comment `gorm:"constraint:OnDelete:CASCADE;foreignKey:PostID"`
	Reaction []Reaction `gorm:"contraints:OnDelete:CASCADE;forienKey:PostID"`

}

type Comment struct {
	gorm.Model
	Content string
	CommentedBy uint
	PostID uint
	CommentID uint
	Replies []Comment `gorm:"constraint:OnDelete:CASCADE;foreignKey:CommentID"`
	Reaction []Reaction `gorm:"contraints:OnDelete:CASCADE;forienKey:CommentID"`
}

type Reaction struct {
	gorm.Model
	PostID uint
	CommentID uint
	ReactionType string
	ReactedBy uint
}
