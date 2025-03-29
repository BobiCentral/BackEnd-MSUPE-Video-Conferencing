// main.go
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password []byte
}

var users = make(map[string]User)                                // Пока вместо бд
var tmpl = template.Must(template.ParseGlob("templates/*.html")) // Просто примеры страниц, чтоб проверить

func homeHandler(w http.ResponseWriter, r *http.Request) { // первая страница с выбором рег или вход
	tmpl.ExecuteTemplate(w, "home.html", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) { // страница с рег
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if _, ok := users[username]; ok {
			tmpl.ExecuteTemplate(w, "register.html", "Пользователь уже существует") // проверка существует такой логин или нет
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // хэширование пароля (вроде я это даже понял)
		if err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		users[username] = User{username, hashedPassword}
		http.Redirect(w, r, "/login", http.StatusSeeOther) // сохраняет данные и показывает страницу входа
		return
	}

	tmpl.ExecuteTemplate(w, "register.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) { // страница входа
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, ok := users[username]
		if !ok {
			tmpl.ExecuteTemplate(w, "login.html", "Неверный логин или пароль") // проверка есть такой логин или нет
			return
		}

		err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
		if err != nil {
			tmpl.ExecuteTemplate(w, "login.html", "Неверный логин или пароль") // проверка пароль совподает с логином или нет
			return
		}
		// Написал "неверный логин или пароль", чтобы не писало, что именно не правильно (я думаю так будет лучше)
		tmpl.ExecuteTemplate(w, "welcome.html", user.Username) // Если всё ок, то показывает страницу успешного входа
		return
	}

	tmpl.ExecuteTemplate(w, "login.html", nil)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
