package app

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Item struct {
	Name      string
	Price     Price
	CreatedAt time.Time
}

const (
	ListText  = "pengen lihat daftar catatan yang mana bos? 👀"
	Today     = "hari ini"
	Yesterday = "kemarin"
	ThisWeek  = "pekan ini"
	PastWeek  = "pekan lalu"
	ThisMonth = "bulan ini"
	PastMonth = "bulan lalu"
)

var (
	monthNames      = []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	replyMarkupList = buildReplyMarkupList()
)

func buildReplyMarkupList() string {
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
		"text":                {ListText},
		"reply_to_message_id": {fmt.Sprintf("%d", msg.MessageID)},
		"reply_markup":        {replyMarkupList},
	})
	return true, err
}

func list(msg *telegram.Message) (bool, error) {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.Text != ListText {
		return false, nil
	}

	items, err := queryItems(msg.Chat.ID, msg.Text)
	if err != nil {
		return true, err
	}

	_, err = sendMessage(url.Values{
		"chat_id":      {fmt.Sprintf("%d", msg.Chat.ID)},
		"text":         {formatItems("catatan "+msg.Text, items)},
		"parse_mode":   {"Markdown"},
		"reply_markup": {`{"remove_keyboard": true}`},
	})
	return true, err
}

func queryItems(chatID int64, interval string) ([]Item, error) {
	query := "SELECT name, price, created_at FROM items WHERE chat_id = $1 AND created_at >= %s AND created_at < %s ORDER BY created_at;"
	today := "DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')"
	tomorrow := fmt.Sprintf("(%s + INTERVAL '1 DAY')", today)
	beginOfWeek := fmt.Sprintf("(%s - INTERVAL '%d DAY')", today, time.Now().In(time.FixedZone("Asia/Jakarta", 7*60*60)).Weekday())
	beginOfMonth := "DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')"

	switch interval {
	case Today:
		query = fmt.Sprintf(query, today, tomorrow)
	case Yesterday:
		query = fmt.Sprintf(query, fmt.Sprintf("(%s - INTERVAL '1 DAY')", today), today)
	case ThisWeek:
		query = fmt.Sprintf(query, beginOfWeek, tomorrow)
	case PastWeek:
		query = fmt.Sprintf(query, fmt.Sprintf("(%s - INTERVAL '7 DAYS')", beginOfWeek), beginOfWeek)
	case ThisMonth:
		query = fmt.Sprintf(query, beginOfMonth, tomorrow)
	case PastMonth:
		query = fmt.Sprintf(query, fmt.Sprintf("(%s - INTERVAL '1 MONTH')", beginOfMonth), beginOfMonth)
	default:
		return nil, nil
	}

	rows, err := db.Query(query, chatID)
	if err != nil {
		return nil, err
	}

	var items []Item
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.Name, &item.Price, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func formatItems(title string, items []Item) string {
	text := fmt.Sprintf("*==== %s ====*\n", strings.ToUpper(title))
	sum := Price(0)
	lastDay := 0
	for _, item := range items {
		if day := item.CreatedAt.Day(); day != lastDay {
			text += fmt.Sprintf("\n_%d %s_\n", day, monthNames[item.CreatedAt.Month()-1])
			lastDay = day
		}
		text += fmt.Sprintf("- %s %s\n", item.Name, item.Price)
		sum += item.Price
	}
	return fmt.Sprintf("%s\n*TOTAL: %s*", text, sum)
}
