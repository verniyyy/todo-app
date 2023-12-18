package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/verniyyy/todo-app/model"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Route("/todo", func(r chi.Router) {
		r.Get("/", ListTOTO)
		r.Post("/", AddTODO)
		r.Post("/{id}", DoneTODO)
		r.Delete("/{id}", DeleteTODO)
	})
	http.ListenAndServe(":8080", r)
}

func ListTOTO(w http.ResponseWriter, r *http.Request) {
	conn, err := getDBConnection(nil)
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}
	defer conn.Close()

	var todos []model.TODO
	if err := WithTx(conn, func(tx *sql.Tx) error {
		rows, err := tx.Query("SELECT * FROM todo")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var todo model.TODO
			if err := rows.Scan(
				&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt,
			); err != nil {
				return err
			}
			todos = append(todos, todo)
		}
		return nil
	}); err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(todos)
}
func AddTODO(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		fmt.Fprintf(w, "Please enter a title")
		return
	}

	conn, err := getDBConnection(nil)
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}
	defer conn.Close()

	if err := WithTx(conn, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO todo (title, done, created_at) VALUES ($1, $2, $3)", title, false, time.Now())
		return err
	}); err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}

	fmt.Fprintf(w, "add todo ok")
}
func DoneTODO(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	conn, err := getDBConnection(nil)
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}
	defer conn.Close()

	if err := WithTx(conn, func(tx *sql.Tx) error {
		result, err := tx.Exec("UPDATE todo SET done = true WHERE id = $1", id)
		if err != nil {
			return err
		}

		aff, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if aff == 0 {
			return fmt.Errorf("id %v is not found", id)
		}

		return nil
	}); err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}

	fmt.Fprintf(w, "done todo %s", id)
}
func DeleteTODO(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	conn, err := getDBConnection(nil)
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}
	defer conn.Close()

	if err := WithTx(conn, func(tx *sql.Tx) error {
		result, err := tx.Exec("DELETE FROM todo where id = $1", id)
		if err != nil {
			return err
		}

		aff, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if aff == 0 {
			return fmt.Errorf("id %v is not found", id)
		}

		return nil
	}); err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
		return
	}

	fmt.Fprintf(w, "delete todo %s", id)
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func (cfg DBConfig) DNS() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
}

var defaultDBCfg = &DBConfig{
	Host:     "localhost",
	Port:     5432,
	User:     "postgres",
	Password: "postgres",
	DBName:   "todo",
}

func getDBConnection(cfg *DBConfig) (*sql.DB, error) {
	if cfg == nil {
		cfg = defaultDBCfg
	}

	conn, err := sql.Open("postgres", cfg.DNS())
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, err
}

func WithTx(conn *sql.DB, f func(*sql.Tx) error) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
