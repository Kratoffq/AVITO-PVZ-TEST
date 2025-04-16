package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/handler"
	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/impl"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	// Загрузка .env файла
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Инициализация конфигурации
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	// Инициализация базы данных
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Очистка базы данных перед тестом
	_, err = db.Exec("TRUNCATE TABLE products, receptions, pvzs CASCADE")
	require.NoError(t, err)

	repo := postgres.NewRepository(db)
	service := impl.NewService(repo, cfg)
	h := handler.NewHandler(service, cfg)
	router := h.InitRoutes()

	// Получение токена сотрудника
	token, err := getEmployeeToken(router)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Создание ПВЗ
	pvzID, err := createPVZ(router, token)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, pvzID)

	// Создание приемки
	receptionID, err := createReception(router, token, pvzID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, receptionID)

	// Добавление 50 товаров
	productTypes := []models.ProductType{
		models.TypeElectronics,
		models.TypeClothing,
		models.TypeShoes,
	}
	for i := 0; i < 50; i++ {
		productType := productTypes[i%len(productTypes)]
		err := addProduct(router, token, pvzID, productType)
		require.NoError(t, err)
	}

	// Закрытие приемки
	err = closeReception(router, token, pvzID)
	require.NoError(t, err)

	// Проверка результатов
	pvzs, err := getPVZs(router, token)
	require.NoError(t, err)
	require.Len(t, pvzs, 1)
	require.Len(t, pvzs[0].Receptions, 1)
	require.Len(t, pvzs[0].Receptions[0].Products, 50)
	if pvzs[0].Receptions[0].Reception.Status != string(models.StatusClose) {
		t.Errorf("Expected reception status to be %v, got %v", string(models.StatusClose), pvzs[0].Receptions[0].Reception.Status)
	}

	// Проверка типов товаров
	productTypeCounts := make(map[string]int)
	for _, product := range pvzs[0].Receptions[0].Products {
		productTypeCounts[string(product.Type)]++
	}
	require.GreaterOrEqual(t, productTypeCounts[string(models.TypeElectronics)], 16) // Примерно 1/3 от 50
	require.GreaterOrEqual(t, productTypeCounts[string(models.TypeClothing)], 16)
	require.GreaterOrEqual(t, productTypeCounts[string(models.TypeShoes)], 16)
}

func getEmployeeToken(router *gin.Engine) (string, error) {
	reqBody := map[string]string{
		"role": string(models.RoleEmployee),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/dummyLogin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		return "", fmt.Errorf("failed to get token: status %d", w.Code)
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		return "", err
	}

	log.Printf("Generated token: %s", response.Token)
	return response.Token, nil
}

func createPVZ(router *gin.Engine, token string) (uuid.UUID, error) {
	reqBody := map[string]string{
		"city": "Москва",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/pvz", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to create PVZ: %s", errorResponse.Error)
		}
		return uuid.Nil, fmt.Errorf("failed to create PVZ: status %d", w.Code)
	}

	var pvz struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &pvz); err != nil {
		return uuid.Nil, err
	}

	return pvz.ID, nil
}

func createReception(router *gin.Engine, token string, pvzID uuid.UUID) (uuid.UUID, error) {
	reqBody := map[string]string{
		"pvz_id": pvzID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/receptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to create reception: %s", errorResponse.Error)
		}
		return uuid.Nil, fmt.Errorf("failed to create reception: status %d", w.Code)
	}

	var reception struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &reception); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal reception response: %v", err)
	}

	return reception.ID, nil
}

func addProduct(router *gin.Engine, token string, pvzID uuid.UUID, productType models.ProductType) error {
	reqBody := map[string]interface{}{
		"pvz_id": pvzID.String(),
		"type":   productType,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to add product: %s", errorResponse.Error)
		}
		return fmt.Errorf("failed to add product: status %d", w.Code)
	}

	return nil
}

func closeReception(router *gin.Engine, token string, pvzID uuid.UUID) error {
	req := httptest.NewRequest("POST", "/api/pvz/"+pvzID.String()+"/close_last_reception", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to close reception: %s", errorResponse.Error)
		}
		return fmt.Errorf("failed to close reception: status %d", w.Code)
	}

	return nil
}

func getPVZs(router *gin.Engine, token string) ([]struct {
	PVZ struct {
		ID               uuid.UUID `json:"id"`
		RegistrationDate time.Time `json:"registration_date"`
		City             string    `json:"city"`
	} `json:"pvz"`
	Receptions []struct {
		Reception struct {
			ID       uuid.UUID `json:"id"`
			DateTime time.Time `json:"date_time"`
			PVZID    uuid.UUID `json:"pvz_id"`
			Status   string    `json:"status"`
		} `json:"reception"`
		Products []struct {
			ID          uuid.UUID `json:"id"`
			DateTime    time.Time `json:"date_time"`
			Type        string    `json:"type"`
			ReceptionID uuid.UUID `json:"reception_id"`
		} `json:"products"`
	} `json:"receptions"`
}, error) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)

	// Получаем первую страницу
	params := url.Values{}
	params.Add("startDate", startDate.Format(time.RFC3339))
	params.Add("endDate", endDate.Format(time.RFC3339))
	params.Add("page", "1")
	params.Add("limit", "30")

	req := httptest.NewRequest("GET", "/api/pvz?"+params.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to get PVZs (page 1): %s", errorResponse.Error)
		}
		return nil, fmt.Errorf("failed to get PVZs (page 1): status %d", w.Code)
	}

	var pvzs []struct {
		PVZ struct {
			ID               uuid.UUID `json:"id"`
			RegistrationDate time.Time `json:"registration_date"`
			City             string    `json:"city"`
		} `json:"pvz"`
		Receptions []struct {
			Reception struct {
				ID       uuid.UUID `json:"id"`
				DateTime time.Time `json:"date_time"`
				PVZID    uuid.UUID `json:"pvz_id"`
				Status   string    `json:"status"`
			} `json:"reception"`
			Products []struct {
				ID          uuid.UUID `json:"id"`
				DateTime    time.Time `json:"date_time"`
				Type        string    `json:"type"`
				ReceptionID uuid.UUID `json:"reception_id"`
			} `json:"products"`
		} `json:"receptions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &pvzs); err != nil {
		return nil, err
	}

	// Получаем вторую страницу
	params.Set("page", "2")
	req = httptest.NewRequest("GET", "/api/pvz?"+params.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			log.Printf("Failed to get PVZs (page 2): %s", errorResponse.Error)
		}
		return nil, fmt.Errorf("failed to get PVZs (page 2): status %d", w.Code)
	}

	var page2Pvzs []struct {
		PVZ struct {
			ID               uuid.UUID `json:"id"`
			RegistrationDate time.Time `json:"registration_date"`
			City             string    `json:"city"`
		} `json:"pvz"`
		Receptions []struct {
			Reception struct {
				ID       uuid.UUID `json:"id"`
				DateTime time.Time `json:"date_time"`
				PVZID    uuid.UUID `json:"pvz_id"`
				Status   string    `json:"status"`
			} `json:"reception"`
			Products []struct {
				ID          uuid.UUID `json:"id"`
				DateTime    time.Time `json:"date_time"`
				Type        string    `json:"type"`
				ReceptionID uuid.UUID `json:"reception_id"`
			} `json:"products"`
		} `json:"receptions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page2Pvzs); err != nil {
		return nil, err
	}

	// Объединяем результаты
	if len(page2Pvzs) > 0 && len(page2Pvzs[0].Receptions) > 0 {
		pvzs[0].Receptions[0].Products = append(pvzs[0].Receptions[0].Products, page2Pvzs[0].Receptions[0].Products...)
	}

	return pvzs, nil
}
