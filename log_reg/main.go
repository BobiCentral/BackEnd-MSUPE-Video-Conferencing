package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings" // Для проверки email
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var db *sql.DB
var tmpl *template.Template

// Структура для передачи данных в шаблоны
type TemplateData struct {
	ErrorMessage string
	Username     string // Будет хранить введенное имя/ФИО при регистрации
	Email        string // Будет хранить введенный email
}

const (
	dbDSN     = "postgres://postgres:2440894@localhost:5432/Mixa?sslmode=disable" // ЗАМЕНИТЕ (postgres:2440894@localhost:5432/Mixa)
	dbTimeout = 5 * time.Second
)

func init() {
	tmpl = template.Must(template.New("").ParseGlob("templates/*.html"))
}

func main() {
	var err error
	db, err = initDB(dbDSN)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()
	log.Println("Пул соединений с БД успешно создан.")

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	port := ":8080"
	fmt.Printf("Сервер запущен на http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

func initDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия соединения с БД: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	err = conn.PingContext(ctx)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка проверки соединения с БД (ping): %w", err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)
	return conn, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, "home.html", nil)
}

// registerHandler - регистрация по Имени, Email, Паролю
func registerHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{}

	if r.Method == http.MethodPost {
		// Получаем имя (из поля username), email и пароль
		usernameInput := r.FormValue("username") // Имя пользователя
		emailInput := r.FormValue("email")       // Email
		passwordInput := r.FormValue("password") // Пароль

		// Сохраняем введенные значения для шаблона
		data.Username = usernameInput
		data.Email = emailInput

		// Валидация
		if usernameInput == "" || emailInput == "" || passwordInput == "" {
			data.ErrorMessage = "Имя, Email и Пароль обязательны для заполнения"
			executeTemplate(w, "register.html", data)
			return
		}
		// Простая проверка формата email (можно улучшить)
		if !strings.Contains(emailInput, "@") || !strings.Contains(emailInput, ".") {
			data.ErrorMessage = "Введите корректный Email адрес"
			executeTemplate(w, "register.html", data)
			return
		}

		// --- Проверка существования EMAIL в БД ---
		// Теперь уникальность проверяем по email, т.к. он используется для входа
		queryCheck := `SELECT 1 FROM users WHERE email = $1 LIMIT 1`
		ctxCheck, cancelCheck := context.WithTimeout(r.Context(), dbTimeout)
		defer cancelCheck()

		var exists int
		// Используем emailInput для проверки
		err := db.QueryRowContext(ctxCheck, queryCheck, emailInput).Scan(&exists)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Ошибка проверки существования email '%s': %v", emailInput, err)
			data.ErrorMessage = "Ошибка сервера при проверке данных"
			executeTemplate(w, "register.html", data)
			return
		}
		if err == nil { // Email уже существует
			data.ErrorMessage = "Пользователь с таким Email уже существует"
			executeTemplate(w, "register.html", data)
			return
		}
		// Если err == sql.ErrNoRows, email свободен

		// --- Хеширование пароля ---
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordInput), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Ошибка хеширования пароля для email '%s': %v", emailInput, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		// --- Вставка пользователя в БД ---
		// Вставляем имя (usernameInput) в столбец username, emailInput в email
		queryInsert := `
            INSERT INTO users (username, email, password_hash)
            VALUES ($1, $2, $3);
        `
		ctxInsert, cancelInsert := context.WithTimeout(r.Context(), dbTimeout)
		defer cancelInsert()

		// Передаем имя, email, хеш
		_, err = db.ExecContext(ctxInsert, queryInsert, usernameInput, emailInput, string(hashedPassword))
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				// Ошибка уникальности (скорее всего, email, т.к. мы его проверили)
				log.Printf("Конфликт при вставке email '%s' (вероятно, гонка запросов): %v", emailInput, err)
				data.ErrorMessage = "Пользователь с таким Email уже существует"
				executeTemplate(w, "register.html", data)
			} else {
				log.Printf("Ошибка вставки пользователя с email '%s' в БД: %v", emailInput, err)
				http.Error(w, "Внутренняя ошибка сервера при регистрации", http.StatusInternalServerError)
			}
			return
		}

		log.Printf("Пользователь '%s' (email: %s) успешно зарегистрирован.", usernameInput, emailInput)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	executeTemplate(w, "register.html", data)
}

// loginHandler - вход по Email и Паролю
func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{}

	if r.Method == http.MethodPost {
		// Получаем email и пароль
		emailInput := r.FormValue("email") // Используем name="email" из формы
		passwordInput := r.FormValue("password")

		// Сохраняем введенный email для повторного отображения
		data.Email = emailInput

		if emailInput == "" || passwordInput == "" {
			data.ErrorMessage = "Email и пароль обязательны"
			executeTemplate(w, "login.html", data)
			return
		}

		// --- Получение имени пользователя и хеша пароля из БД по EMAIL ---
		var storedHash string
		var storedUsername string // Имя пользователя для приветствия
		// Ищем по email, получаем username и password_hash
		query := `SELECT username, password_hash FROM users WHERE email = $1`
		ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
		defer cancel()

		// Сканируем имя и хеш
		err := db.QueryRowContext(ctx, query, emailInput).Scan(&storedUsername, &storedHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Email не найден
				log.Printf("Попытка входа с несуществующим email: %s", emailInput)
				data.ErrorMessage = "Неверный Email или пароль" // Общее сообщение
				executeTemplate(w, "login.html", data)
			} else {
				log.Printf("Ошибка получения данных для email '%s': %v", emailInput, err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			}
			return
		}

		// --- Сравнение хеша из БД и введенного пароля ---
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(passwordInput))
		if err != nil {
			// Пароль не совпадает
			log.Printf("Неудачная попытка входа (неверный пароль) для email: %s", emailInput)
			data.ErrorMessage = "Неверный Email или пароль" // Общее сообщение
			executeTemplate(w, "login.html", data)
			return
		}

		// --- Успешный вход ---
		log.Printf("Пользователь '%s' (email: %s) успешно вошел.", storedUsername, emailInput)
		// Передаем ИМЯ пользователя (storedUsername) в шаблон welcome.html
		executeTemplate(w, "welcome.html", storedUsername)
		return
	}

	executeTemplate(w, "login.html", data)
}

func executeTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона %s: %v", templateName, err)
		http.Error(w, "Внутренняя ошибка сервера при отображении страницы", http.StatusInternalServerError)
	}
}
