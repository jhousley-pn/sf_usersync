package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
)

type SFUser struct {
	ID        string `json:"Id"`
	Username  string `json:"Username"`
	Email     string `json:"Email"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Phone     string `json:"Phone"`
	Pin       string `json:"Pin__c"`
	CreatedAt string `json:"CreatedDate"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Heroku Go API is running ✅"))
	})

	http.HandleFunc("/sync-users", syncUsersHandler)
	http.HandleFunc("/get-pin", getPinHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func syncUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []SFUser
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Println("JSON Decode Error:", err)
		return
	}

	for _, u := range users {
		normalizedPhone := normalizePhone(u.Phone)

		_, err := db.Exec(`
			INSERT INTO sf_users (
				id, username, email, first_name, last_name, phone, pin, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				username = EXCLUDED.username,
				email = EXCLUDED.email,
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				phone = EXCLUDED.phone,
				pin = EXCLUDED.pin;
		`, u.ID, u.Username, u.Email, u.FirstName, u.LastName, normalizedPhone, u.Pin, u.CreatedAt)

		if err != nil {
			log.Printf("DB error for user %s: %v\n", u.ID, err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User sync complete"))
}

func getPinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	phoneRaw := r.URL.Query().Get("phone")
	if phoneRaw == "" {
		http.Error(w, "Missing phone parameter", http.StatusBadRequest)
		return
	}

	normalized := normalizePhone(phoneRaw)
	if len(normalized) != 10 {
		http.Error(w, "Phone number must be 10 digits", http.StatusBadRequest)
		return
	}

	var pin sql.NullString
	err := db.QueryRow(`SELECT pin FROM sf_users WHERE REGEXP_REPLACE(phone, '\D', '', 'g') = $1`, normalized).Scan(&pin)
	if err == sql.ErrNoRows {
		http.Error(w, "PIN not found for phone number", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"phone": normalized,
		"pin":   pin.String,
	})
}

func normalizePhone(raw string) string {
	re := regexp.MustCompile(`\D`)
	digitsOnly := re.ReplaceAllString(raw, "")
	if len(digitsOnly) > 10 {
		digitsOnly = digitsOnly[len(digitsOnly)-10:]
	}
	return digitsOnly
}
