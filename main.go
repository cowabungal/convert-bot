package main

import (
	"convert-bot/photopdf"
	"flag"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var BotToken *string

var source = "./assets"
var tempDirectory string


func main() {
	BotToken = flag.String("token", "", "telegram.org")
	flag.Parse()

	bot, err := tgbotapi.NewBotAPI(*BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		} else if update.Message.Text != "" && !update.Message.IsCommand() {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я умею конвертировать фотки в пдф. В левом нижнем углу есть меню со всеми командами)\nОтправь /go и загрузи все фоточки одним сообщением."))
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {

			switch update.Message.Command() {
			case "go":
				if tempDirectory != "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Загрузи все фоточки одним сообщением и ПОДОЖДИ 10сек.\nСкинь команду /convert, когда будешь готов отправить на конвертацию!\n\nЕсли что-то пошло не так, то нажми на /cancel и попробуй еще раз."))
				} else{
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Загрузи все фоточки одним сообщением и ПОДОЖДИ 10сек.\nСкинь команду /convert, когда будешь готов отправить на конвертацию!"))
					tempDirectory = tempDir()
				}

			case "convert":
				if tempDirectory == "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала отправь команду /go."))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Cейчас всё будет сделано :)"))
					log.Printf("Начинаю конвертацию")
					photopdf.Convert(tempDirectory)
					log.Printf("Закончил. Ждем 5с.")
				}
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды. Список доступных команд в меню в левом нижнем углу"))
			}
		}

		if update.Message.Photo != nil {
			filePath := getTgFilePath(&update, bot)
			if tempDirectory == "" {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала отправь команду /go"))
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
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", *BotToken, url))
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
