package main

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	regionKeyBoard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("EU"),
			tgbotapi.NewKeyboardButton("US"),
			tgbotapi.NewKeyboardButton("KR"),
			tgbotapi.NewKeyboardButton("TW"),
		),
	)
	mainKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Change region"),
			tgbotapi.NewKeyboardButton("Affixes"),
		),
	)
	localeKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("en"),
			tgbotapi.NewKeyboardButton("de"),
			tgbotapi.NewKeyboardButton("it"),
			tgbotapi.NewKeyboardButton("fr"),
			tgbotapi.NewKeyboardButton("es"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ru"),
			tgbotapi.NewKeyboardButton("cn"),
			tgbotapi.NewKeyboardButton("pt"),
			tgbotapi.NewKeyboardButton("ko"),
			tgbotapi.NewKeyboardButton("tw"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Return"),
		),
	)
)

func MainHandler(db *sql.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, _ *http.Request) {
		// stub
		_, _ = db.Exec("CREATE TABLE IF NOT EXISTS chats_to_commands (chat_id INTEGER UNIQUE, command VARCHAR(10) )")
		resp.Write([]byte("Raider IO Telegram Bot"))
	}
}

func checkPlayer(region, text string, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) tgbotapi.MessageConfig {

	i := strings.Index(text, " ")
	if text == "" || i == -1 {
		msg.Text = "Please send correct name and realm"
		return msg
	}

	name := text[:i]
	realm := text[i+1:]
	if realm == "" {
		msg.Text = "Please send correct name and realm"
		return msg
	}

	info, err := GetScore(region, name, realm)
	if err != nil {
		if err.Error() == "" {
			msg.Text = "Player info not found "
		} else {
			msg.Text = err.Error()

		}
	} else {
		scores := info.Season[0].Scores
		heartLevel := info.Gear.ArtifactTraits / 65535
		txt := "*Name*: %s \n" +
			"*Race*: %s \n" +
			"*Class*: %s \n" +
			"*Profile*: [Raider.io](%s) \n" +
			"*Gear Item level Equiped*: %v \n" +
			"*Gear Item level Total*: %v \n" +
			"*Heart of Azeroth level*: %v \n" +
			"*Guild*: %s \n" +
			"*Score Current Season*:\n" +
			"    *All*: %v \n" +
			"    *DPS*: %v \n" +
			"    *Healer*: %v \n" +
			"    *Tank*: %v \n"

		msg.Text = fmt.Sprintf(txt, info.Name, info.Race, info.Class, info.ProfileURL, info.Gear.ItemLevelEquiped,
			info.Gear.ItemLevelTotal, heartLevel, info.Guild.Name, scores.All, scores.Dps, scores.Healer, scores.Tank)
		msg.ParseMode = "markdown"
	}
	return msg
}

func insertOrUpdate(db *sql.DB, chatID int64, command string) error {
	_, err := db.Exec(fmt.Sprintf(
		`INSERT INTO chats_to_commands 
		     ( chat_id
		     , command
		     )
		VALUES (%v, '%s')
		ON CONFLICT (chat_id) 
		DO UPDATE 
		      SET command = EXCLUDED.command`,
		chatID, command))
	return err
}

func selectCommand(db *sql.DB, chatID int64) (string, error) {
	var region string
	err := db.QueryRow(fmt.Sprintf(
		`SELECT command 
          FROM chats_to_commands
         WHERE chat_id = %v`,
		chatID)).Scan(&region)
	return region, err
}

func main() {

	// подключаемся к боту с помощью токена
	token := os.Getenv("TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}
	// инициализируем канал, куда будут прилетать обновления от API
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	//updates, err := bot.GetUpdatesChan(u)
	updates := bot.ListenForWebhook("/" + bot.Token)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	http.HandleFunc("/", MainHandler(db))
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.ReplyMarkup = regionKeyBoard
				msg.Text = "Now choose region"
			case "help":
				msg.Text = "Choose player region by telegram command and then send his name and realm: Niko Altar of Storms"
			default:
				msg.Text = "Command not found"
			}
		} else {
			txt := update.Message.Text
			switch txt {
			case "EU", "US", "KR", "TW":
				if err := insertOrUpdate(db, chatID, txt); err != nil {
					log.Println(err)
					msg.Text = "Wooops Error"
					_, _ = bot.Send(msg)
					continue
				}
				msg.Text = "Now you can send player name and realm: Niko Altar of Storms"
				msg.ReplyMarkup = mainKeyboard
			case "en", "ru", "ko", "cn", "pt", "it", "fr", "es", "de", "tw":
				region, err := selectCommand(db, chatID)
				if err != nil {
					log.Println(err)
					msg.Text = "Wooops Error"
					_, _ = bot.Send(msg)
					continue
				}

				affixes, err := GetAffixes(region, txt)
				if err != nil {
					msg.Text = "Wooops Error"
				} else {

					affTxt := "*%s*:\n" +
						"    *Info*: [Wowhead](%s) \n" +
						"    *Description*: %s \n\n"

					txt := fmt.Sprintf("*Affixes*: %s \n*Leader board*:[Raider.io](%s) \n\n",
						affixes.Title, affixes.LeaderboardURL)

					for _, aff := range affixes.AffixDetails {
						txt = txt + fmt.Sprintf(affTxt, aff.Name, aff.WowheadURL, aff.Description)
					}
					msg.Text = txt
					msg.ParseMode = "markdown"
				}
				msg.ReplyMarkup = mainKeyboard
			case "Change region":
				msg.Text = "Choose new region"
				msg.ReplyMarkup = regionKeyBoard
			case "Affixes":
				msg.Text = "Choose language"
				msg.ReplyMarkup = localeKeyboard
			case "Return":
				msg.Text = "Return to main menu"
				msg.ReplyMarkup = mainKeyboard
			default:
				region, err := selectCommand(db, chatID)
				if err != nil {
					log.Println(err)
					msg.Text = "You must choose region"
					_, _ = bot.Send(msg)
					continue
				}

				msg = checkPlayer(region, text, bot, msg)

			}
		}
		_, _ = bot.Send(msg)
	}
}
