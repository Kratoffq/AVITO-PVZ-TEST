package monitoring

import (
	"testing"
)

func TestMetrics(t *testing.T) {
	t.Run("HTTP метрики", func(t *testing.T) {
		// Тестируем счетчик HTTP запросов
		IncHTTPRequest("GET", "/test", "200")
		IncHTTPRequest("POST", "/test", "400")

		// Тестируем длительность HTTP запросов
		ObserveHTTPRequestDuration("GET", "/test", 0.5)
		ObserveHTTPRequestDuration("POST", "/test", 1.0)
	})

	t.Run("Метрики базы данных", func(t *testing.T) {
		// Тестируем длительность запросов к БД
		ObserveDBQueryDuration("select", 0.1)
		ObserveDBQueryDuration("insert", 0.2)

		// Тестируем счетчик ошибок БД
		IncDBError("select")
		IncDBError("insert")
	})

	t.Run("Метрики ПВЗ", func(t *testing.T) {
		// Тестируем счетчик операций с ПВЗ
		IncPVZOperation("create", "success")
		IncPVZOperation("update", "error")

		// Тестируем длительность операций с ПВЗ
		ObservePVZOperationDuration("create", 0.3)
		ObservePVZOperationDuration("update", 0.4)
	})

	t.Run("Метрики товаров", func(t *testing.T) {
		// Тестируем счетчик операций с товарами
		IncProductOperation("add", "success")
		IncProductOperation("delete", "error")

		// Тестируем длительность операций с товарами
		ObserveProductOperationDuration("add", 0.1)
		ObserveProductOperationDuration("delete", 0.2)
	})

	t.Run("Метрики приемок", func(t *testing.T) {
		// Тестируем счетчик операций с приемками
		IncReceptionOperation("create", "success")
		IncReceptionOperation("close", "error")

		// Тестируем длительность операций с приемками
		ObserveReceptionOperationDuration("create", 0.2)
		ObserveReceptionOperationDuration("close", 0.3)
	})

	t.Run("Бизнес метрики", func(t *testing.T) {
		// Тестируем счетчики бизнес-метрик
		IncPVZCreated()
		IncReceptionCreated()
		IncProductAdded()
	})
}
