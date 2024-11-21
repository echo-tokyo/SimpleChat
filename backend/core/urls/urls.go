package urls

import (
	echo "github.com/labstack/echo/v4"

	userUrls "SimpleChat/backend/app_user/urls"
	chatUrls "SimpleChat/backend/app_chat/urls"
)


// подгрузка urls каждого микроприложения и их общая настройка
func InitUrlRouters(echoApp *echo.Echo) {
	apiUserGroup := echoApp.Group("/api/user")
	userUrls.RouterGroup(apiUserGroup)

	apiChatGroup := echoApp.Group("/api/chat")
	chatUrls.RouterGroup(apiChatGroup)
}