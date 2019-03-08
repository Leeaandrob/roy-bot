package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"

	tb "gopkg.in/tucnak/telebot.v2"
)

type RoyTranscoderStatus struct {
	QueueSize int    `json:"queue_size"`
	Current   string `json:"current"`
}

func healthCheckRoyTranscoder() (status bool) {
	resp, err := http.Get("https://transcoder.roy.video/ping")
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		status = false
	} else {
		status = true
	}

	defer resp.Body.Close()
	return
}

func statusRoyTranscoder() (status RoyTranscoderStatus) {
	resp, err := http.Get("https://transcoder.roy.video/transcode/proccess")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(body, &status)

	defer resp.Body.Close()
	return
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Init the env
func InitEnv() {
	if len(os.Getenv("DIR")) != 0 {
		err := godotenv.Load(os.Getenv("DIR"))
		checkError(err)

	} else {
		dir, errr := os.Getwd()
		checkError(errr)

		err := godotenv.Load(path.Join(dir, ".env"))
		checkError(err)
	}
}

func init() {
	InitEnv()
}

func main() {
	// telegram bot
	b, _ := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_KEY"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	// Roy Transcoder healthcheck
	replyBtnRoyTranscoderHealthCheck := tb.ReplyButton{Text: "@Transcoder - healthcheck"}
	b.Handle(&replyBtnRoyTranscoderHealthCheck, func(m *tb.Message) {
		go func() {
			if !healthCheckRoyTranscoder() {
				b.Send(m.Chat, fmt.Sprintf("%s Parece que tem alguma coisa errada no Roy Transcoder 😞 Consegue mandar uma mensagem pros humanos que programaram?", m.Chat.Username))
			} else {
				b.Send(m.Chat, fmt.Sprintf("%s Parece que está tudo ok 😄", m.Chat.Username))
			}
		}()
	})

	// Roy Transcoder healthcheck
	replyBtnRoyTranscoderStatus := tb.ReplyButton{Text: "@Transcoder - status"}
	b.Handle(&replyBtnRoyTranscoderStatus, func(m *tb.Message) {
		go func() {
			status := statusRoyTranscoder()
			b.Send(m.Chat, fmt.Sprintf("%s O tamanho da fila é %d o video que está sendo processado no momento é %s 😄", m.Chat.Username, status.QueueSize, status.Current))
		}()
	})

	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{replyBtnRoyTranscoderHealthCheck},
		[]tb.ReplyButton{replyBtnRoyTranscoderStatus},
	}

	b.Handle(tb.OnText, func(m *tb.Message) {
		b.Send(m.Chat, "Oi! Eis a lista de coisas que posso te responder:\n\n @Transcoder: healthcheck, status do servidor.", &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
	})

	b.Start()
}
