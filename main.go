package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"personal-web/connection"

	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	connection.DatabaseConnect()
	route := mux.NewRouter()
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))
	route.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", home).Methods("GET")

	route.HandleFunc("/project", myProject).Methods("GET")
	
	route.HandleFunc("/form-project", myProjectForm).Methods("GET")
	route.HandleFunc("/add-project", myProjectData).Methods("POST")
	
	route.HandleFunc("/delete-project/{id}", myProjectDelete).Methods("GET")
	route.HandleFunc("/form-edit-project/{id}", myEdit).Methods("GET")
	route.HandleFunc("/edit-project/{id}", myProjectEdited).Methods("POST")
	
	route.HandleFunc("/form-register", formRegister).Methods(("GET"))
	route.HandleFunc("/register", register).Methods("POST")


	route.HandleFunc("/form-login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server running at localhost port 8000")
	http.ListenAndServe("localhost:8000", route)
}
type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

type StructInputDataForm struct {
	Id              int
	Title     string
	Content     string
	
	IsLogin  		 bool
	
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func home(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("views/index.html")
	if err != nil {
		panic(err)
	}
	template.Execute(w, nil)
}

func myProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/myProject.html")
	
	
//

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}



	fm := session.Flashes("message")
//perlu loping karena nanti ketika refresh si alertny masih ada
	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, f1 := range fm {
			// meamasukan flash message
			flashes = append(flashes, f1.(string))
		}
	}


	Data.FlashData = strings.Join(flashes, "")






//
	
	data, _ := connection.Conn.Query(context.Background(), "SELECT id, title, content FROM tb_blog")
	var result []StructInputDataForm
	for data.Next() {
		var each = StructInputDataForm{}
		err := data.Scan(&each.Id, &each.Title, &each.Content)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		
		result = append(result, each)
	}

	response := map[string]interface{}{
		"DataSession": Data,
		"Projects": result,
	}

	if err == nil {
		tmpl.Execute(w, response)
	} else {
		w.Write([]byte("Message: "))
		w.Write([]byte(err.Error()))
	}
}



func myProjectForm(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/myProjectForm.html")
	if err == nil {
		tmpl.Execute(w, nil)
	} else {
		panic(err)
	}
}


func myProjectData(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	
	

	

 
	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")


	
	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_blog (title, content  ) VALUES ($1, $2)", title, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	http.Redirect(w, r, "/project", http.StatusMovedPermanently)
}

func myEdit(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/myEdit.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectEdit := StructInputDataForm{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id, title, content FROM tb_blog WHERE id=$1", id).Scan(
		&ProjectEdit.Id, &ProjectEdit.Title, &ProjectEdit.Content )
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}
	

	response := map[string]interface{}{
		"Project": ProjectEdit,
	}

	if err == nil {
		tmpl.Execute(w, response)
	} else {
		panic(err)
	}
}

func myProjectEdited(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var title string
	var content string
	
	fmt.Println(r.Form)
	for i, values := range r.Form {
		for _, value := range values {
			if i == "title" {
				title = value
			}
			if i == "content" {
				content = value
			}
			
		}
	}
	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_blog SET title=$1, content=$2 WHERE id=$3", title, content, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	http.Redirect(w, r, "/project", http.StatusMovedPermanently)
}
func myProjectDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_blog WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}
	http.Redirect(w, r, "/project", http.StatusMovedPermanently)
}



func formRegister(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-register.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	// fmt.Println(passwordHash)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-login.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, f1 := range fm {
			// meamasukan flash message
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	tmpl.Execute(w, Data)
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	user := User{}

	// mengambil data email, dan melakukan pengecekan email
	err = connection.Conn.QueryRow(context.Background(),
		"SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {

		// fmt.Println("Email belum terdaftar")
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))// cookiedari browser
		session, _ := store.Get(r, "SESSION_KEY")

		//session = menyimpan data
		// _  = menampilkan data
		session.AddFlash("Email belum terdaftar!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		// w.WriteHeader(http.StatusBadRequest)
		// w.Write([]byte("message : Email belum terdaftar " + err.Error()))
		return
	}

	// melakukan pengecekan password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// fmt.Println("Password salah")
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Password Salah!", "message")
		session.Save(r, w)

		
		
		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		// w.WriteHeader(http.StatusBadRequest)
		// w.Write([]byte("message : Email belum terdaftar " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	// berfungsi untuk menyimpan data kedalam session browser
	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 10800 // 3 JAM expred

	
	
	session.Save(r, w)

	http.Redirect(w, r, "/project", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/form-login", http.StatusSeeOther)
}