package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}

type Quest struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Cost  int    `json:"cost"`
	Steps int    `json:"steps"`
}

type UserQuest struct {
	UserID  int `json:"user_id"`
	QuestID int `json:"quest_id"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./quests.db")
	if err != nil {
		log.Fatal(err)
	}

	createUserTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"name" TEXT,
		"balance" INTEGER
	  );`

	createQuestTableSQL := `CREATE TABLE IF NOT EXISTS quests (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"cost" INTEGER,
		"steps" INTEGER
	  );`

	createUserQuestTableSQL := `CREATE TABLE IF NOT EXISTS user_quests (
		"user_id" INTEGER,
		"quest_id" INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(quest_id) REFERENCES quests(id),
		PRIMARY KEY (user_id, quest_id)
	  );`

	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createQuestTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createUserQuestTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	statement, err := db.Prepare("INSERT INTO users (name, balance) VALUES (?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = statement.Exec(user.Name, user.Balance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func createQuest(w http.ResponseWriter, r *http.Request) {
	var quest Quest
	err := json.NewDecoder(r.Body).Decode(&quest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	statement, _ := db.Prepare("INSERT INTO quests (name, cost, steps) VALUES (?, ?, ?)")
	_, err = statement.Exec(quest.Name, quest.Cost, quest.Steps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(quest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func completeQuest(w http.ResponseWriter, r *http.Request) {
	var userQuest UserQuest
	err := json.NewDecoder(r.Body).Decode(&userQuest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_quests WHERE user_id=? AND quest_id=?)", userQuest.UserID, userQuest.QuestID).Scan(&exists)
	if err != nil || exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var cost int
	err = db.QueryRow("SELECT cost FROM quests WHERE id = ?", userQuest.QuestID).Scan(&cost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", cost, userQuest.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO user_quests (user_id, quest_id) VALUES (?, ?)", userQuest.UserID, userQuest.QuestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func userHistory(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	var quests []Quest
	rows, err := db.Query("SELECT q.id, q.name, q.cost, q.steps FROM quests q JOIN user_quests uq ON q.id = uq.quest_id WHERE uq.user_id = ?", userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var quest Quest
		if err := rows.Scan(&quest.ID, &quest.Name, &quest.Cost, &quest.Steps); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		quests = append(quests, quest)
	}

	var balance int
	err = db.QueryRow("SELECT balance FROM users WHERE id = ?", userId).Scan(&balance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := struct {
		Quests  []Quest `json:"quests"`
		Balance int     `json:"balance"`
	}{
		Quests:  quests,
		Balance: balance,
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func main() {
	initDB()

	http.HandleFunc("/create_user", createUser)
	http.HandleFunc("/create_quest", createQuest)
	http.HandleFunc("/complete_quest", completeQuest)
	http.HandleFunc("/user_history", userHistory)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
