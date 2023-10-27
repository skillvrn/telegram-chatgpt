package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Отправка текста на ChatGPT API и получение ответа
		response, err := getChatGPTResponse(update.Message.Text, client)
		if err != nil {
			log.Println("Ошибка при получении ответа от ChatGPT:", err)
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		bot.Send(msg)
	}
}

func getChatGPTResponse(inputText string, client *http.Client) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API ключ ChatGPT не установлен")
	}

	url := "https://api.openai.com/v1/engines/davinci/completions"
	data := fmt.Sprintf(`{"prompt": "%s", "max_tokens": 50}`, inputText)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ошибка при запросе к ChatGPT API. Код состояния: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if output, ok := result["choices"].([]interface{}); ok && len(output) > 0 {
		if text, ok := output[0].(map[string]interface{})["text"].(string); ok {
			return text, nil
		}
	}

	return "", fmt.Errorf("Не удалось получить ответ от ChatGPT API")
}
