package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const openAIURL = "https://api.openai.com/v1/engines/davinci/completions"

func chatWithGPT(userMessage string) string {
	reqBody := strings.NewReader(`{
		"prompt": "` + userMessage + `",
		"max_tokens": 150
	}`)

	req, err := http.NewRequest("POST", openAIURL, reqBody)
	if err != nil {
		return "Ошибка: " + err.Error()
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "Ошибка: " + err.Error()
	}
	defer resp.Body.Close()

	var responseJson map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseJson)

	choices := responseJson["choices"].([]interface{})
	if len(choices) > 0 {
		return choices[0].(map[string]interface{})["text"].(string)
	}

	return "Извините, я не могу обработать ваш запрос."
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	allowedUserIDs := strings.Split(os.Getenv("ALLOWED_USER_IDS"), ",")

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userID := update.Message.From.ID
		isAllowed := false
		for _, allowedID := range allowedUserIDs {
			if allowedID == userID {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Извините, у вас нет доступа к этому боту."))
			continue
		}

		response := chatWithGPT(update.Message.Text)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, response))
	}
}
