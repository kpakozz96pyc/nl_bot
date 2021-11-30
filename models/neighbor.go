package models

type Neighbor struct {
	Id                int64  `db:"id" json:"id"`
	Name              string `db:"name" json:"name"`
	TelegramFirstName string `db:"telegram_first_name" json:"telegramFirstName"`
	TelegramLastName  string `db:"telegram_last_name" json:"telegramLastName"`
	TelegramUserName  string `db:"telegram_user_name" json:"telegramUserName"`
	TelegramUserId    int64  `db:"telegram_user_id" json:"telegramUserId"`
	Turn              int64  `db:"turn" json:"turn"`
	Section           int64  `db:"section" json:"section"`
	Building          int64  `db:"building" json:"building"`
	Flat              int64  `db:"flat" json:"flat"`
	Floor             int64  `db:"floor" json:"floor"`
}
