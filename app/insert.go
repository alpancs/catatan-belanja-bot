package app

import (
	"fmt"
	"strings"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	NewItemsText = "apa saja yang pengen dicatat, bos?"
	SaveTemplate = "*%s %s* dicatat ya bos 👌 #catatan"
)

var (
	forceReply = `{"force_reply":true,"selective":true}`
)

func commandInsert(msg *telegram.Message) (bool, error) {
	if msg.Command() != "catat" {
		return false, nil
	}

	_, err := sendMessageCustom(msg.Chat.ID, NewItemsText, msg.MessageID, forceReply)
	return true, err
}

func insert(msg *telegram.Message) (bool, error) {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.Text != NewItemsText {
		return false, nil
	}

	for _, text := range strings.Split(msg.Text, "\n") {
		err := insertOneLine(msg, text)
		if err != nil {
			return true, err
		}
	}
	return true, nil
}

func insertOneLine(msg *telegram.Message, text string) error {
	priceText := patternPrice.FindString(text)
	item := strings.TrimSpace(text[:len(text)-len(priceText)])
	if item == "" || priceText == "" {
		return nil
	}

	price := ParsePrice(priceText)
	resp, err := sendMessage(msg.Chat.ID, fmt.Sprintf(SaveTemplate, item, price), 0)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO items VALUES ($1, $2, $3, $4, $5);", resp.Chat.ID, resp.MessageID, item, price, time.Now().In(time.UTC))
	if err != nil {
		revertReport(msg, resp, item, price)
	}
	return err
}

func revertReport(req, resp *telegram.Message, item string, price Price) {
	deleteMessage(resp.Chat.ID, resp.MessageID)
	sendMessage(req.Chat.ID, fmt.Sprintf("%s %s gagal dicatat bos 😔 #gagalmaningsonson", item, price), req.MessageID)
}
