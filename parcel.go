package main

import (
	"database/sql"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Добавляет новую строку в таблицу parcel
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client), 
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
	sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	// Получаем идентификатор последней добавленной записи
	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastId), nil
}

// Читает строку из таблицы parcel по заданному number
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}

	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number))

	// Заполняем структуру данными из результата запроса
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

// Возвращает все строки из таблицы parcel для конкретного клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Создаём срез для хранения результата
	var res []Parcel

	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Итерируем по всем строкам результата
	for rows.Next() {
		var p Parcel
		// Считываем текущую строку в структуру Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Добавляем структуру в срез
		res = append(res, p)
	}

	// Проверяем наличие ошибок во время итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Jбновляет статус строки в таблице parcel
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		return err
	}		
	return nil
}

// Обновляет адрес строки в таблице parcel
// Менять адрес можно только если текущий статус — "registered"
func (s ParcelStore) SetAddress(number int, address string) error {
	// Выполняем UPDATE с условием проверки статуса
	_, err := s.db.Exec(
		"UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	if err != nil {
		return err
	}
	return nil
}

// Удаляет строку из таблицы parcel
// Удаление разрешено только если статус строки — "registered"
func (s ParcelStore) Delete(number int) error {
	// Выполняем DELETE с условием проверки статуса
	_, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	if err != nil {
		return err
	}
	return nil
}