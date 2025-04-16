package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	mock.ExpectExec("INSERT INTO pvz").
		WithArgs(pvz.ID, pvz.RegistrationDate, pvz.City).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreatePVZ(context.Background(), pvz)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	pvzID := uuid.New()
	expectedPVZ := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
		AddRow(expectedPVZ.ID, expectedPVZ.RegistrationDate, expectedPVZ.City)

	mock.ExpectQuery("SELECT id, registration_date, city FROM pvzs WHERE id = \\$1").
		WithArgs(pvzID).
		WillReturnRows(rows)

	pvz, err := repo.GetPVZByID(context.Background(), pvzID)
	require.NoError(t, err)
	require.Equal(t, expectedPVZ, pvz)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   models.StatusInProgress,
	}

	mock.ExpectExec("INSERT INTO receptions").
		WithArgs(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateReception(context.Background(), reception)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: uuid.New(),
	}

	mock.ExpectExec("INSERT INTO products").
		WithArgs(product.ID, product.DateTime, product.Type, product.ReceptionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateProduct(context.Background(), product)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	pvzID := uuid.New()
	expectedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
		AddRow(expectedReception.ID, expectedReception.DateTime, expectedReception.PVZID, expectedReception.Status)

	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(rows)

	reception, err := repo.GetLastOpenReception(context.Background(), pvzID)
	require.NoError(t, err)
	require.Equal(t, expectedReception, reception)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	receptionID := uuid.New()

	mock.ExpectExec("UPDATE receptions SET status = \\$1 WHERE id = \\$2").
		WithArgs(models.StatusClose, receptionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CloseReception(context.Background(), receptionID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	productID := uuid.New()

	mock.ExpectExec("DELETE FROM products WHERE id = \\$1").
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeleteProduct(context.Background(), productID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.Password, user.Role, user.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	email := "test@example.com"
	expectedUser := &models.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  "password123",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.CreatedAt)

	mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	email := "nonexistent@example.com"

	mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	require.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		Password:  "password123",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.CreatedAt)

	mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.GetUserByID(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.New()

	mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(context.Background(), userID)
	require.NoError(t, err)
	require.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZsWithReceptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()
	page := 1
	limit := 10

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"pvz_id", "registration_date", "city",
		"reception_id", "reception_date_time", "pvz_id", "status",
		"product_id", "product_date_time", "product_type",
	}).
		AddRow(
			pvzID, time.Now(), "Москва",
			receptionID, time.Now(), pvzID, models.StatusClose,
			productID, time.Now(), models.TypeElectronics,
		)

	mock.ExpectQuery("WITH filtered_receptions AS").
		WithArgs(startDate, endDate, limit, (page-1)*limit).
		WillReturnRows(rows)

	pvzs, err := repo.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)
	require.NoError(t, err)
	require.Len(t, pvzs, 1)
	require.Len(t, pvzs[0].Receptions, 1)
	require.Len(t, pvzs[0].Receptions[0].Products, 1)
	require.Equal(t, models.StatusClose, pvzs[0].Receptions[0].Reception.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	receptionID := uuid.New()
	expectedReception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   models.StatusInProgress,
	}

	rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
		AddRow(expectedReception.ID, expectedReception.DateTime, expectedReception.PVZID, expectedReception.Status)

	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE id = \\$1").
		WithArgs(receptionID).
		WillReturnRows(rows)

	reception, err := repo.GetReceptionByID(context.Background(), receptionID)
	require.NoError(t, err)
	require.Equal(t, expectedReception, reception)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	receptionID := uuid.New()

	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE id = \\$1").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)

	reception, err := repo.GetReceptionByID(context.Background(), receptionID)
	require.NoError(t, err)
	require.Nil(t, reception)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	receptionID := uuid.New()
	expectedProduct := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: receptionID,
	}

	rows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
		AddRow(expectedProduct.ID, expectedProduct.DateTime, expectedProduct.Type, expectedProduct.ReceptionID)

	mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products WHERE reception_id = \\$1 ORDER BY date_time DESC LIMIT 1").
		WithArgs(receptionID).
		WillReturnRows(rows)

	product, err := repo.GetLastProduct(context.Background(), receptionID)
	require.NoError(t, err)
	require.Equal(t, expectedProduct, product)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastProduct_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	receptionID := uuid.New()

	mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products WHERE reception_id = \\$1 ORDER BY date_time DESC LIMIT 1").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetLastProduct(context.Background(), receptionID)
	require.NoError(t, err)
	require.Nil(t, product)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	products := []*models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.TypeElectronics,
			ReceptionID: uuid.New(),
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.TypeClothing,
			ReceptionID: uuid.New(),
		},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO products \\(id, date_time, type, reception_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)")
	for _, product := range products {
		mock.ExpectExec("INSERT INTO products \\(id, date_time, type, reception_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
			WithArgs(product.ID, product.DateTime, product.Type, product.ReceptionID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err = repo.CreateProducts(context.Background(), products)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZsWithReceptions_MultipleProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()
	page := 1
	limit := 10

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()

	rows := sqlmock.NewRows([]string{
		"pvz_id", "registration_date", "city",
		"reception_id", "reception_date_time", "pvz_id", "status",
		"product_id", "product_date_time", "product_type",
	}).
		AddRow(
			pvzID, time.Now(), "Москва",
			receptionID, time.Now(), pvzID, models.StatusClose,
			productID1, time.Now(), models.TypeElectronics,
		).
		AddRow(
			pvzID, time.Now(), "Москва",
			receptionID, time.Now(), pvzID, models.StatusClose,
			productID2, time.Now(), models.TypeClothing,
		)

	mock.ExpectQuery("WITH filtered_receptions AS").
		WithArgs(startDate, endDate, limit, (page-1)*limit).
		WillReturnRows(rows)

	pvzs, err := repo.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)
	require.NoError(t, err)
	require.Len(t, pvzs, 1)
	require.Len(t, pvzs[0].Receptions, 1)
	require.Len(t, pvzs[0].Receptions[0].Products, 2)
	require.Equal(t, models.StatusClose, pvzs[0].Receptions[0].Reception.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProducts_TransactionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	products := []*models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.TypeElectronics,
			ReceptionID: uuid.New(),
		},
	}

	mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction error"))

	err = repo.CreateProducts(context.Background(), products)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZByID_Cache(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	pvzID := uuid.New()
	expectedPVZ := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	// Первый вызов - данные из БД
	rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
		AddRow(expectedPVZ.ID, expectedPVZ.RegistrationDate, expectedPVZ.City)

	mock.ExpectQuery("SELECT id, registration_date, city FROM pvzs WHERE id = \\$1").
		WithArgs(pvzID).
		WillReturnRows(rows)

	// Второй вызов - данные из кэша
	pvz, err := repo.GetPVZByID(context.Background(), pvzID)
	require.NoError(t, err)
	require.Equal(t, expectedPVZ, pvz)

	pvz, err = repo.GetPVZByID(context.Background(), pvzID)
	require.NoError(t, err)
	require.Equal(t, expectedPVZ, pvz)
	require.NoError(t, mock.ExpectationsWereMet())
}
