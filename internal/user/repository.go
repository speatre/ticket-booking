package user

import "gorm.io/gorm"

type Repository interface {
	ByEmail(email string) (*User, error)
	ByID(id string) (*User, error)
	Create(u *User) error
	Update(u *User) error
}

type repo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repo{db} }

func (r *repo) ByEmail(email string) (*User, error) {
	var u User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *repo) ByID(id string) (*User, error) {
	var u User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *repo) Create(u *User) error { return r.db.Create(u).Error }
func (r *repo) Update(u *User) error { return r.db.Save(u).Error }
