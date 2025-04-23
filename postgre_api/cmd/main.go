package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://user:password@localhost:5432/mydb")
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	// Пример: вставим одного пользователя
	id := uuid.New()
	teachers := []uuid.UUID{uuid.New(), uuid.New()}
	students := []uuid.UUID{uuid.New()}
	requests := []uuid.UUID{}

	_, err = conn.Exec(ctx, `
		INSERT INTO users (
			id, fio, username, pass, role, age, specialty, price, rating, teachers, students, requests
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`, id, "Иванов Иван", "ivanov", "pass123", "student", 20, "math", 1000, 4.5, teachers, students, requests)
	if err != nil {
		log.Fatalf("failed to insert user: %v\n", err)
	}

	fmt.Println("Пользователь добавлен")

	// Пример: получим пользователей
	rows, err := conn.Query(ctx, "SELECT id, fio, username, age FROM users")
	if err != nil {
		log.Fatalf("failed to query users: %v\n", err)
	}
	defer rows.Close()

	fmt.Println("Список пользователей:")
	for rows.Next() {
		var id uuid.UUID
		var fio, username string
		var age uint8

		err := rows.Scan(&id, &fio, &username, &age)
		if err != nil {
			log.Printf("scan error: %v\n", err)
			continue
		}

		fmt.Printf("- %s (%s), %d лет\n", fio, username, age)
	}
}
