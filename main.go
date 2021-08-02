package main

import (
	"convert-bot/photopdf"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const BotToken = "" // Enter your Bot Token

var source = "./assets"
var tempDirectory string


func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	/// У каждого пользователя(чат айди) должна быть своя папка с доками. Проверяется условие на существование, если такой нет, то создается. | НЕТ. ДЕЛАЕМ АСИНХРОННОСТЬ!
	/// ДЕЛАЕМ БОТА ДЛЯ НАГРУЗОК В ОДНОМ ПОТОКЕ, ДАЛЕЕ РЕАЛИЗОВЫВАЕМ ТО, ЧТО ВЫШЕ

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch strings.ToLower(update.Message.Text) {
		case "хочу конвертировать":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Загрузи все фоточки одним сообщением. Напиши 'Летсконверт', когда будешь готов отправить на конвертацию!"))
			tempDirectory = tempDir()
		case "летсконверт":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Уно моменто, сейчас всё будет сделано :)"))
			log.Printf("Начинаю конвертацию")
			photopdf.Convert(tempDirectory)
			log.Printf("Закончил. Ждем 5с.")
		case "":
		default:
			if tempDirectory == "" {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Напиши 'Хочу конвертировать', чтобы я принял фотки)."))
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Напиши 'Летсконверт', чтобы я начал конвертацию."))
			}
		}

		if update.Message.Photo != nil {
			filePath := getTgFilePath(&update, bot)
			if tempDirectory == "" {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала напиши 'Хочу конвертировать"))
			} else {
				MakeRequest(filePath, tempDirectory)
			}
		}

		// Если файл существует, то отправляет его
		_, err := os.Stat(fmt.Sprintf("./%s/result.pdf", tempDirectory))
		if  !os.IsNotExist(err) {
			bot.Send(tgbotapi.NewDocumentUpload(update.Message.Chat.ID, fmt.Sprintf("./%s/result.pdf", tempDirectory)))
			os.RemoveAll(fmt.Sprintf("./%s",tempDirectory))
			os.Remove(fmt.Sprintf("./%s", tempDirectory))
			tempDirectory = ""
		}
	}
}

func getTgFilePath(update *tgbotapi.Update, bot *tgbotapi.BotAPI) string {
	myUpdate := *update
	file := myUpdate.Message.Photo
	fileInfo := *file
	fileId := fileInfo[3].FileID
	endfile, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		return fmt.Sprintf("error getFilePath")
	}
	return endfile.FilePath
}

func tempDir() string {

	// TempDir возвращает путь
	tDir, err := ioutil.TempDir(source, "ggwpletsconvert")
	if err != nil {
		panic(err)
	}
	return tDir

}

func MakeRequest(url string, tempDir string) {
	time.Sleep(1 * time.Second)
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", BotToken, url))
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalln("не 200!")
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	ioutil.WriteFile(fmt.Sprintf("./%s/%d.jpg", tempDir, time.Now().Unix()), body, 0o666)
}


