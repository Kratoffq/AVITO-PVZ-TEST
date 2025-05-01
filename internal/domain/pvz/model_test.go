package pvz

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		city string
		want *PVZ
	}{
		{
			name: "успешное создание PVZ",
			city: "Москва",
			want: &PVZ{
				City: "Москва",
			},
		},
		{
			name: "создание PVZ с пустым городом",
			city: "",
			want: &PVZ{
				City: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.city)

			// Проверяем, что ID сгенерирован
			assert.NotEmpty(t, got.ID)

			// Проверяем, что время создания установлено корректно
			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)

			// Проверяем город
			assert.Equal(t, tt.want.City, got.City)
		})
	}
}

func TestPVZ_Fields(t *testing.T) {
	// Тест для проверки полей структуры
	pvz := &PVZ{
		City: "Санкт-Петербург",
	}

	// Проверяем, что поля доступны и имеют правильные типы
	assert.IsType(t, "", pvz.City)
	assert.IsType(t, time.Time{}, pvz.CreatedAt)
}
