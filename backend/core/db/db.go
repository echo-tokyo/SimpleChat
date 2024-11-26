package db

import (
	"strings"

	echo "github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"github.com/google/uuid"

	"SimpleChat/backend/core/services"
	"SimpleChat/backend/core/db/models"
)


// структура для запросов к БД
type DB struct {
	dbConnect *gorm.DB
}
func NewDB() *DB {
	return &DB{
		dbConnect: dbConnection,
	}
}


// создание нового юзера
func (db *DB) CreateUser(username, password string) (models.User, error) {
	var newUser models.User

	// генерим новый uuid для записи
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return newUser, echo.NewHTTPError(500, map[string]string{"createUser": "failed to create uuid for user"})
	}
	// делаем из пароля хэш
	passwordHash, err := services.EncodePassword(password)
	if err != nil {
		return newUser, err
	}

	newUser = models.User{
		ID: newUUID,
		Username: username,
		Password: passwordHash,
	}

	createResult := db.dbConnect.Create(&newUser)
	if err := createResult.Error; err != nil {
		// если юзер с таким юзернеймом уже есть
		if strings.HasSuffix(err.Error(), "(2067)") {
			return models.User{}, echo.NewHTTPError(409, map[string]string{"username": "user with such username already exists"})
		}
		return models.User{}, echo.NewHTTPError(500, map[string]string{"createUser": "failed to create user: " + err.Error()})
	}
	return newUser, nil
}

// получение юзера по его ID
func (db *DB) GetUserByID(id uuid.UUID) (models.User, error) {
	var userFromDB models.User

	selectResult := db.dbConnect.First(&userFromDB, id)
	if err := selectResult.Error; err != nil {
		// если ошибка в ненахождении записи
		if err.Error() == "record not found" {
			return models.User{}, echo.NewHTTPError(404, map[string]string{"getUser": "user with such id was not found"})
		}
		return models.User{}, echo.NewHTTPError(500, map[string]string{"getUser": "failed to get user by id: " + err.Error()})
	}
	return userFromDB, nil
}

// получение юзера по его username'у
func (db *DB) GetUserByUsername(username string) (models.User, error) {
	var userFromDB models.User

	selectResult := db.dbConnect.Where("username = ?", username).First(&userFromDB)
	if err := selectResult.Error; err != nil {
		// если ошибка в ненахождении записи
		if err.Error() == "record not found" {
			return models.User{}, echo.NewHTTPError(404, map[string]string{"getUser": "user with such username was not found"})
		}
		return models.User{}, echo.NewHTTPError(500, map[string]string{"getUser": "failed to get user by username: " + err.Error()})
	}
	return userFromDB, nil
}


// получение чата (с подгрузкой его участников и сообщений) по его ID
func (db *DB) GetFullChatByID(id uuid.UUID) (models.Chat, error) {
	var chatFromDB models.Chat

	selectResult := db.dbConnect.Preload("Users").Preload(
		"Messages",
		// добавление сортировки сообщений по времени от старых к новым
		func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at ASC")
		},
	).Preload("Messages.Sender").First(&chatFromDB, id)
	
	if err := selectResult.Error; err != nil {
		// если ошибка в ненахождении записи
		if err.Error() == "record not found" {
			return models.Chat{}, echo.NewHTTPError(404, map[string]string{"getChat": "chat with such id was not found"})
		}
		return models.Chat{}, echo.NewHTTPError(500, map[string]string{"getChat": "failed to get chat by id: " + err.Error()})
	}
	return chatFromDB, nil
}

// получение чата (с подгрузкой его участников) по его ID
func (db *DB) GetChatParticipantsByID(id uuid.UUID) (models.Chat, error) {
	var chatFromDB models.Chat

	selectResult := db.dbConnect.Preload("Users").First(&chatFromDB, id)
	if err := selectResult.Error; err != nil {
		// если ошибка в ненахождении записи
		if err.Error() == "record not found" {
			return models.Chat{}, echo.NewHTTPError(404, map[string]string{"getChat": "chat with such id was not found"})
		}
		return models.Chat{}, echo.NewHTTPError(500, map[string]string{"getChat": "failed to get chat by id: " + err.Error()})
	}
	return chatFromDB, nil
}

// создание нового чата
func (db *DB) createChat(firstUserFromDB, secondUserFromDB models.User) (models.Chat, error) {
	var chat models.Chat

	// генерим новый uuid для записи
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return chat, echo.NewHTTPError(500, map[string]string{"createChat": "failed to create uuid for chat"})
	}
	// создаём чат с привязкой к нему двух юзеров
	chat = models.Chat{
		ID: newUUID,
		Users: []models.User{firstUserFromDB, secondUserFromDB},
	}
	createResult := db.dbConnect.Create(&chat)
	if err := createResult.Error; err != nil {
		return models.Chat{}, echo.NewHTTPError(500, map[string]string{"createChat": "failed to create chat"})
	}
	return chat, nil
}

// получение чата или создание нового, если такого ещё нет 
func (db *DB)  GetOrCreateChat(firstUserID, secondUserID uuid.UUID) (models.Chat, error) {
	var chat models.Chat

	var firstUserFromDB, secondUserFromDB models.User
	// получение чатов первого юзера
	selectResult := db.dbConnect.Preload("Chats").First(&firstUserFromDB, firstUserID)
	if err := selectResult.Error; err != nil {
		return chat, echo.NewHTTPError(500, map[string]string{"getUser": "failed to get user by id: " + err.Error()})
	}
	// получение чатов второго юзера
	selectResult = db.dbConnect.Preload("Chats").First(&secondUserFromDB, secondUserID)
	if err := selectResult.Error; err != nil {
		return chat, echo.NewHTTPError(500, map[string]string{"getUser": "failed to get user by id: " + err.Error()})
	}

	// совместные чаты юзеров (в контексте этого проекта он должен быть один)
	joinChats := services.IntersectUserChats(firstUserFromDB.Chats, secondUserFromDB.Chats)
	
	var err error
	// если пересечение не было найдено, то создаём новый чат
	if len(joinChats) == 0 {
		chat, err = db.createChat(firstUserFromDB, secondUserFromDB)
		return chat, err
	// если пересечение было найдено, то возвращаем чат (только его id без подгрузки участников и сообщений)
	} else {
		return joinChats[0], err
	}

	return chat, nil
}


// создание нового сообщения
func (db *DB) CreateMessage(chatID, senderID uuid.UUID, content string) (models.Message, error) {
	var newMessage models.Message

	// генерим новый uuid для записи
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return newMessage, echo.NewHTTPError(500, map[string]string{"createMessage": "failed to create uuid for message"})
	}

	// создаём новую запись сообщения
	newMessage = models.Message{
		ID: newUUID,
		ChatID: chatID,
		SenderID: senderID,
		Content: content,
	}

	createResult := db.dbConnect.Create(&newMessage)
	if err := createResult.Error; err != nil {
		return models.Message{}, echo.NewHTTPError(500, map[string]string{"createMessage": "failed to create message: " + err.Error()})
	}

	senderUser, err := db.GetUserByID(senderID)
	if err != nil {
		return newMessage, err
	}
	newMessage.Sender = senderUser
	
	return newMessage, nil
}
