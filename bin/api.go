package bin

import (
	"bytes"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

var GlobalUrl, ServerToken, GlobalError, GlobalMessage string
var data = GetConfigData()
var Client = redis.NewClient(&redis.Options{})

// Оснвной универсальный метод отправки запросов в API
func Request(url string, typeReq string, token string, bodyReq any, bodyResp any) *http.Response {
	jsonReq, _ := json.Marshal(bodyReq)                                         //Преобразование тела запроса в json
	req, _ := http.NewRequest(typeReq, GlobalUrl+url, bytes.NewBuffer(jsonReq)) //Создание http запроса
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	response, _ := client.Do(req)
	if response != nil {
		body, _ := io.ReadAll(response.Body)
		_ = json.Unmarshal(body, bodyResp) //Обратное преобразование из json
	} else {
		SaveLog(log.Fields{
			"group": "server",
		}, log.ErrorLevel, "No response from API")
		panic("Нет ответа от Api")
	}
	return response
}

// Метод подключения и проверки доступа к API
func ConnectAPI() {
	SaveLog(log.Fields{
		"group": "server",
	}, log.TraceLevel, "Connecting to API and Redis...")
	db, err := strconv.Atoi(data["RedisDB"])
	CheckErr(err)
	var password string
	if data["RedisPassword"] == "-" {
		password = ""
	} else {
		password = data["RedisPassword"]
	}
	Client = redis.NewClient(&redis.Options{
		Addr:     data["RedisAddr"],
		Password: password,
		DB:       db,
	})
	GlobalUrl = data["Api"]
	user := map[string]string{
		"Email":    data["Login"],
		"Password": data["Password"],
	}
	var result map[string]string
	response := Request("/login", "POST", "", user, &result)
	if response.StatusCode == 200 {
		SaveLog(log.Fields{
			"group": "server",
		}, log.InfoLevel, "Received server token")
		ServerToken = result["token"]
	} else {
		SaveLog(log.Fields{
			"group": "server",
		}, log.ErrorLevel, "Server token was not received")
		panic("Серверный токен не был получен")
	}
}

func RefreshAPI() {
	response := Request("/check", "GET", ServerToken, nil, nil)
	if response.StatusCode != 200 {
		ConnectAPI()
	}
}
