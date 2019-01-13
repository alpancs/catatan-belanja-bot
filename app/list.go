package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	Today     = "hari ini"
	Yesterday = "kemarin"
	ThisWeek  = "pekan ini"
	PastWeek  = "pekan lalu"
	ThisMonth = "bulan ini"
	PastMonth = "bulan lalu"
)

var keyboardList = buildKeyboardList()

func buildKeyboardList() string {
	raw, err := json.Marshal(telegram.ReplyKeyboardMarkup{
		Keyboard: [][]telegram.KeyboardButton{
			{{Text: Today}, {Text: Yesterday}},
			{{Text: ThisWeek}, {Text: PastWeek}},
			{{Text: ThisMonth}, {Text: PastMonth}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Selective:       true,
	})
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func commandList(msg *telegram.Message) (bool, error) {
	if msg.Command() != "lihat" {
		return false, nil
	}

	_, err := sendMessage(url.Values{
		"chat_id":             {fmt.Sprintf("%d", msg.Chat.ID)},
		"text":                {"pengen lihat daftar catatan yang mana bos? 👀"},
		"reply_to_message_id": {fmt.Sprintf("%d", msg.MessageID)},
		"reply_markup":        {keyboardList},
	})
	return true, err
}
