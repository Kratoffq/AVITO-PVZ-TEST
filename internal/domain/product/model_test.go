package product

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	receptionID := uuid.New()

	tests := []struct {
		name        string
		receptionID uuid.UUID
		productType Type
		want        *Product
	}{
		{
			name:        "создание электронного товара",
			receptionID: receptionID,
			productType: TypeElectronics,
			want: &Product{
				ReceptionID: receptionID,
				Type:        TypeElectronics,
			},
		},
		{
			name:        "создание товара одежды",
			receptionID: receptionID,
			productType: TypeClothing,
			want: &Product{
				ReceptionID: receptionID,
				Type:        TypeClothing,
			},
		},
		{
			name:        "создание продуктового товара",
			receptionID: receptionID,
			productType: TypeFood,
			want: &Product{
				ReceptionID: receptionID,
				Type:        TypeFood,
			},
		},
		{
			name:        "создание прочего товара",
			receptionID: receptionID,
			productType: TypeOther,
			want: &Product{
				ReceptionID: receptionID,
				Type:        TypeOther,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.receptionID, tt.productType)

			// Проверяем, что ID сгенерирован
			assert.NotEmpty(t, got.ID)

			// Проверяем, что время создания установлено корректно
			assert.WithinDuration(t, time.Now(), got.DateTime, time.Second)

			// Проверяем остальные поля
			assert.Equal(t, tt.want.ReceptionID, got.ReceptionID)
			assert.Equal(t, tt.want.Type, got.Type)
		})
	}
}

func TestType_Constants(t *testing.T) {
	// Проверяем константы типов товаров
	assert.Equal(t, Type("electronics"), TypeElectronics)
	assert.Equal(t, Type("clothing"), TypeClothing)
	assert.Equal(t, Type("food"), TypeFood)
	assert.Equal(t, Type("other"), TypeOther)
}

func TestProduct_Fields(t *testing.T) {
	// Тест для проверки полей структуры
	receptionID := uuid.New()
	product := &Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        TypeElectronics,
		ReceptionID: receptionID,
	}

	// Проверяем, что поля доступны и имеют правильные типы
	assert.IsType(t, uuid.UUID{}, product.ID)
	assert.IsType(t, time.Time{}, product.DateTime)
	assert.IsType(t, Type(""), product.Type)
	assert.IsType(t, uuid.UUID{}, product.ReceptionID)
}
