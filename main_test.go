package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUser(t *testing.T) {
	initDB()

	user := User{Name: "Test User", Balance: 100}
	userBytes, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/create_user", bytes.NewBuffer(userBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createUser)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var createdUser User
	err = json.NewDecoder(rr.Body).Decode(&createdUser)
	if err != nil {
		t.Fatal(err)
	}

	if createdUser.Name != user.Name {
		t.Errorf("handler returned unexpected body: got name %v want name %v",
			createdUser.Name, user.Name)
	}
}

func TestCreateQuest(t *testing.T) {
	initDB()

	quest := Quest{Name: "Test Quest", Cost: 50, Steps: 1}
	questBytes, _ := json.Marshal(quest)
	request, err := http.NewRequest("POST", "/create_quest", bytes.NewBuffer(questBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createQuest)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var createdQuest Quest
	err = json.NewDecoder(rr.Body).Decode(&createdQuest)
	if err != nil {
		t.Fatal(err)
	}

	if createdQuest.Name != quest.Name {
		t.Errorf("handler returned unexpected body: got name %v want name %v",
			createdQuest.Name, quest.Name)
	}
}

func TestCompleteQuest(t *testing.T) {
	initDB()

	user := User{Name: "Test User for Complete", Balance: 100}
	_, err := db.Exec("INSERT INTO users (name, balance) VALUES (?, ?)", user.Name, user.Balance)
	if err != nil {
		t.Fatal(err)
	}
	userID := getLastInsertID("users")

	quest := Quest{Name: "Test Quest for Complete", Cost: 50, Steps: 1}
	_, err = db.Exec("INSERT INTO quests (name, cost, steps) VALUES (?, ?, ?)", quest.Name, quest.Cost, quest.Steps)
	if err != nil {
		t.Fatal(err)
	}
	questID := getLastInsertID("quests")

	userQuest := UserQuest{UserID: int(userID), QuestID: int(questID)}
	userQuestBytes, _ := json.Marshal(userQuest)
	request, err := http.NewRequest("POST", "/complete_quest", bytes.NewBuffer(userQuestBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(completeQuest)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var balance int
	err = db.QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	if balance != user.Balance+quest.Cost {
		t.Errorf("Expected user balance to be %d, got %d", user.Balance+quest.Cost, balance)
	}
}
func TestUserHistory(t *testing.T) {
	initDB()

	user := User{Name: "Test User for History", Balance: 100}
	_, err := db.Exec("INSERT INTO users (name, balance) VALUES (?, ?)", user.Name, user.Balance)
	if err != nil {
		t.Fatal(err)
	}
	userID := getLastInsertID("users")

	quest := Quest{Name: "Test Quest for History", Cost: 50, Steps: 1}
	_, err = db.Exec("INSERT INTO quests (name, cost, steps) VALUES (?, ?, ?)", quest.Name, quest.Cost, quest.Steps)
	if err != nil {
		t.Fatal(err)
	}
	questID := getLastInsertID("quests")

	_, err = db.Exec("INSERT INTO user_quests (user_id, quest_id) VALUES (?, ?)", userID, questID)
	if err != nil {
		t.Fatal(err)
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("/user_history?user_id=%d", userID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(userHistory)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response struct {
		Quests  []Quest `json:"quests"`
		Balance int     `json:"balance"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Quests) == 0 {
		t.Error("Expected at least one quest in user history")
	}

	if response.Balance != user.Balance {
		t.Errorf("Expected user balance to be %d, got %d", user.Balance, response.Balance)
	}
}

func getLastInsertID(tableName string) int64 {
	var id int64
	query := fmt.Sprintf("SELECT last_insert_rowid() FROM %s", tableName)
	err := db.QueryRow(query).Scan(&id)
	if err != nil {
		fmt.Printf("%v", err)
	}
	return id
}
