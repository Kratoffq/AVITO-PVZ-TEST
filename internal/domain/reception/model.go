package reception

import (
	"time"

	"github.com/google/uuid"
)

// Status представляет статус приемки
type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClose      Status = "close"
)

// Reception представляет собой приемку товаров
type Reception struct {
	ID       uuid.UUID `db:"id"`
	DateTime time.Time `db:"date_time"`
	PVZID    uuid.UUID `db:"pvz_id"`
	Status   Status    `db:"status"`
}

// New создает новый экземпляр Reception
func New(pvzID uuid.UUID) *Reception {
	return &Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   StatusInProgress,
	}
}

// Close закрывает приемку
func (r *Reception) Close() {
	r.Status = StatusClose
}
