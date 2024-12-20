package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
	
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	// Подключение к базе данных 
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание экземпляра ParcelStore для работы с БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавление посылки
	id, err := store.Add(parcel)
	require.NoError(t, err)  // Проверка, что ошибки нет
	require.NotZero(t, id)  // Проверка, что идентификатор посылки не равен нулю
	
	// get
	// Получение добавленной посылки
	storedParcel, err := store.Get(id)
	require.NoError(t, err)  // Проверка, что ошибки нет
	parcel.Number = id  // Устанавливаем идентификатор добавленной посылки в исходной структуре, чтобы структуры стали идентичными
	require.Equal(t, parcel, storedParcel)  // Сравнение всей структуры целиком

	// delete
	// Удаление добавленной посылки
	err = store.Delete(id)
	require.NoError(t, err)  // Проверка, что ошибки нет

	// Проверка, что посылка удалена
	_, err = store.Get(id)
	require.Error(t, err)  // Ожидаем ошибку при попытке получить удалённую запись
	require.ErrorIs(t, err, sql.ErrNoRows)  // Проверяем, что ошибка — отсутствие записи в БД
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// Подключение к базе данных
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание экземпляра ParcelStore
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавление посылки
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set address
	// Обновление адреса посылки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// Проверка, что адрес обновился
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	if err != nil {
		require.ErrorIs(t, err, sql.ErrNoRows)
    	return
	}
	require.Equal(t, newAddress, storedParcel.Address)  // Проверка нового адреса
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// Подключение к базе данных
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание экземпляра ParcelStore
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// Добавление посылки
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set status
	// Обновление статуса посылки
	newStatus := "sent"
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	// Проверка, что статус обновился
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	if err != nil {
		require.ErrorIs(t, err, sql.ErrNoRows)
		return
	}
	require.Equal(t, newStatus, storedParcel.Status)  // Проверка нового статуса
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// Подключение к базе данных
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание экземпляра ParcelStore
	store := NewParcelStore(db)

	// Создаём несколько посылок для одного клиента
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := map[int]Parcel{}

	// Генерируем уникальный идентификатор клиента для всех посылок
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	// Добавление всех посылок в базу данных
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // Добавляем посылку
		require.NoError(t, err)
		parcels[i].Number = id  // Обновляем номер добавленной посылки
		parcelMap[id] = parcels[i]  // Сохраняем посылку в map для последующей проверки
	}

	// get by client
	// Получение всех посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client) 
	require.NoError(t, err)
	if err != nil {
		require.ErrorIs(t, err, sql.ErrNoRows)
		return
	}
	require.Len(t, storedParcels, 3)  // Проверка, что получено 3 записи

	// check
	// Проверка совпадения добавленных и полученных посылок
	for _, parcel := range storedParcels {
		assert.Contains(t, parcelMap, parcel.Number)  // Убедитесь, что посылка есть в map
		if assert.Contains(t, parcelMap, parcel.Number) {  // Дополнительная проверка перед сравнением
			assert.Equal(t, parcelMap[parcel.Number], parcel)
		}
	}
}