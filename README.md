# REST API Service for User Quests

## Overview
This REST API service manages users and quests, allowing users to complete quests once and earn rewards.

## API Endpoints

### Create User
- **Endpoint**: `/create_user`
- **Method**: POST
- **Body**:
  ```json
  {
    "name": "User Name",
    "balance": 100
  }
  ```
- **Description**: Creates a new user.

### Create Quest
- **Endpoint**: `/create_quest`
- **Method**: POST
- **Body**:
  ```json
  {
    "name": "Quest Name",
    "cost": 50,
    "steps": 1
  }
  ```
- **Description**: Creates a new quest.

### Complete Quest
- **Endpoint**: `/complete_quest`
- **Method**: POST
- **Body**:
  ```json
  {
    "user_id": 1,
    "quest_id": 1
  }
  ```
- **Description**: Marks a quest as completed by a user and updates the user's balance.

### User History
- **Endpoint**: `/user_history`
- **Method**: GET
- **Query Parameters**:
  - `user_id`: The ID of the user.
- **Description**: Returns the list of quests completed by the user and their current balance.

## Running the Server
Run the server with:
```sh
docker build -t quest-app . && docker run -p 8080:8080 quest-app
```

## Testing
Run unit tests with:
```sh
go test
```

## Example Curl Commands

### Create User
```shell
curl -X POST http://localhost:8080/create_user \
     -H "Content-Type: application/json" \
     -d '{"name": "New User", "balance": 100}'
```

### Create Quest
```shell
curl -X POST http://localhost:8080/create_quest \
     -H "Content-Type: application/json" \
     -d '{"name": "New Quest", "cost": 50, "steps": 1}'
```