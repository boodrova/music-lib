package main

import (
	"log"
	"net/http"

	"music-lib/api"
	"music-lib/config" // Подключаем config пакет
	"music-lib/db"
	"music-lib/pkg/client"

	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию из .env
	cfg := config.LoadConfig()

	// Формируем DSN для подключения к базе данных
	dsn := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"
	if dsn == "" {
		log.Fatal("[ERROR] DATABASE_URL is not set in .env file")
	}

	// Создаем репозиторий
	repo, err := db.NewRepository(dsn)
	if err != nil {
		log.Fatalf("[ERROR] Error connecting to the database: %v", err)
	}

	// Применяем миграции
	if err := applyMigrations(dsn); err != nil {
		log.Fatalf("[ERROR] Error applying migrations: %v", err)
	}

	// Создаем клиент Swagger
	client := client.NewSwaggerClient(cfg.SwaggerAPI)

	// Инициализируем обработчики API
	handler := &api.Handler{
		Repo:   repo,
		Client: client,
	}

	// Создаем маршруты
	r := mux.NewRouter()
	r.HandleFunc("/songs", handler.GetSongs).Methods("GET")
	r.HandleFunc("/songs", handler.AddSong).Methods("POST")
	r.HandleFunc("/songs/text/{id}", handler.GetSongText).Methods("GET")
	r.HandleFunc("/songs/{id}", handler.DeleteSong).Methods("DELETE")
	r.HandleFunc("/songs/{id}", handler.UpdateSong).Methods("PUT")

	// Используем порт из конфигурации
	port := ":" + cfg.APIPort
	log.Fatal(http.ListenAndServe(port, r))
}

func applyMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
