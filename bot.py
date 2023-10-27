import requests
import os
from telegram import Bot, Update
from telegram.ext import MessageHandler, Filters, CallbackContext, Updater

# Загрузка переменных окружения из файла .env
TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
ALLOWED_USER_IDS = os.getenv("ALLOWED_USER_IDS").split(',')

def chat_with_gpt(user_message: str) -> str:
    headers = {
        'Authorization': f'Bearer {OPENAI_API_KEY}',
        'Content-Type': 'application/json'
    }
    data = {
        'prompt': user_message,
        'max_tokens': 150
    }
    response = requests.post('https://api.openai.com/v1/engines/davinci/completions', headers=headers, json=data)
    response_json = response.json()

    return response_json['choices'][0]['text'].strip()

def handle_text(update: Update, context: CallbackContext) -> None:
    user_id = str(update.message.from_user.id)
    
    if user_id not in ALLOWED_USER_IDS:
        update.message.reply_text("Извините, у вас нет доступа к этому боту.")
        return

    user_message = update.message.text
    gpt_response = chat_with_gpt(user_message)
    update.message.reply_text(gpt_response)

def main():
    updater = Updater(token=TELEGRAM_BOT_TOKEN)
    dp = updater.dispatcher
    dp.add_handler(MessageHandler(Filters.text & ~Filters.command, handle_text))
    updater.start_polling()
    updater.idle()

if __name__ == "__main__":
    main()
