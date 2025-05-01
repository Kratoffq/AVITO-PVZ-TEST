package reception

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	pvzID := uuid.New()

	tests := []struct {
		name  string
		pvzID uuid.UUID
		want  *Reception
	}{
		{
			name:  "успешное создание приемки",
			pvzID: pvzID,
			want: &Reception{
				PVZID:  pvzID,
				Status: StatusInProgress,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.pvzID)

			// Проверяем, что ID сгенерирован
			assert.NotEmpty(t, got.ID)

			// Проверяем, что время создания установлено корректно
			assert.WithinDuration(t, time.Now(), got.DateTime, time.Second)

			// Проверяем остальные поля
			assert.Equal(t, tt.want.PVZID, got.PVZID)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestReception_Close(t *testing.T) {
	tests := []struct {
		name       string
		reception  *Reception
		wantStatus Status
	}{
		{
			name: "закрытие активной приемки",
			reception: &Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   StatusInProgress,
			},
			wantStatus: StatusClose,
		},
		{
			name: "закрытие уже закрытой приемки",
			reception: &Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   StatusClose,
			},
			wantStatus: StatusClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reception.Close()
			assert.Equal(t, tt.wantStatus, tt.reception.Status)
		})
	}
}

func TestStatus_Constants(t *testing.T) {
	// Проверяем константы статусов
	assert.Equal(t, Status("in_progress"), StatusInProgress)
	assert.Equal(t, Status("close"), StatusClose)
}

func TestReception_Fields(t *testing.T) {
	// Тест для проверки полей структуры
	pvzID := uuid.New()
	reception := &Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   StatusInProgress,
	}

	// Проверяем, что поля доступны и имеют правильные типы
	assert.IsType(t, uuid.UUID{}, reception.ID)
	assert.IsType(t, time.Time{}, reception.DateTime)
	assert.IsType(t, uuid.UUID{}, reception.PVZID)
	assert.IsType(t, Status(""), reception.Status)
}
