package event

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventRepository interface {
	List() ([]Event, error)
	ListPage(limit, offset int) ([]Event, error)
	Get(id string) (*Event, error)
	Create(e *Event) error
	Update(e *Event) error
	Delete(id string) error
	Reserve(tx *gorm.DB, eventID string, qty int) (bool, error)   // legacy atomic
	ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error) // new explicit tx reservation
}

type repo struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) EventRepository { return &repo{db} }

func (r *repo) List() ([]Event, error) {
	var out []Event
	return out, r.db.Order("starts_at asc").Find(&out).Error
}

func (r *repo) ListPage(limit, offset int) ([]Event, error) {
	var out []Event
	q := r.db.Order("starts_at asc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return out, q.Find(&out).Error
}

func (r *repo) Get(id string) (*Event, error) {
	var e Event
	if err := r.db.First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repo) Create(e *Event) error  { return r.db.Create(e).Error }
func (r *repo) Update(e *Event) error  { return r.db.Save(e).Error }
func (r *repo) Delete(id string) error { return r.db.Delete(&Event{}, "id = ?", id).Error }

// atomic reservation (used in legacy code)
func (r *repo) Reserve(tx *gorm.DB, eventID string, qty int) (bool, error) {
	res := tx.Exec(`UPDATE events 
        SET remaining = remaining - ? 
        WHERE id = ? AND remaining >= ?`, qty, eventID, qty)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

// ReserveTx: explicit tx version (DB-first fallback)
func (r *repo) ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error) {
	// dùng row-level locking để tránh race condition
	var e Event
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&e, "id = ?", eventID).Error; err != nil {
		return false, err
	}
	if e.Remaining < qty {
		return false, nil
	}
	e.Remaining -= qty
	if err := tx.Save(&e).Error; err != nil {
		return false, err
	}
	return true, nil
}
