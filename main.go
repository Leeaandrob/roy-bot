package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"

	tb "gopkg.in/tucnak/telebot.v2"
)

type RoyTranscoderStatus struct {
	QueueSize int   `json:"queue_size"`
	Video     Video `json:"video"`
}

type RoyAPI struct {
	Results []Tenant `json:"results"`
}

type Video struct {
	ID          string `json:"id"`
	Tenant      string `json:"tenant"`
	Filename    string `json:"filename"`
	File        string `json:"file"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	Owner       string `json:"owner"`
}

type Tenant struct {
	Hostname string `json:"hostname"`
	Schema   string `json:"schema"`
}

func getRoyTenants(envinroment string) (tenants []Tenant) {
	var data RoyAPI

	urls := make(map[string]string)

	urls["release"] = "https://api.roy.solutions/v1/tenants?page=1&size=100"
	urls["stage"] = "https://api.roystaging.com/v1/tenants?page=1&size=100"

	resp, err := http.Get(urls[envinroment])

	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(body, &data)
	defer resp.Body.Close()

	tenants = data.Results

	return
}

func healthCheckRoyTranscoder() (status bool) {
	resp, err := http.Get("http://transcoder.roy.solutions/ping")
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

func checkJobRoyTranscoder(payload ...string) (job string) {
	resp, err := http.Get(fmt.Sprintf("http://transcoder.roy.solutions/jobs?tenant=%s&id=%s&type=%s", payload[0], payload[1], payload[2]))
	if err != nil {
		log.Fatalln(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	job = string(body)
	defer resp.Body.Close()
	return
}

func statusRoyTranscoder() (status RoyTranscoderStatus) {
	resp, err := http.Get("http://transcoder.roy.solutions/transcode/proccess")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(body, &status)

	if status.Video.Environment != "" {
		tenants := getRoyTenants(status.Video.Environment)

		for _, t := range tenants {
			if t.Schema == status.Video.Tenant {
				status.Video.Owner = t.Hostname
			}
		}
	}

	defer resp.Body.Close()
	return
}

func createJobRoyTranscoder(payload ...string) bool {
	jsonStr := map[string]interface{}{
		"tenant":      fmt.Sprintf("%s", payload[0]),
		"id":          payload[1],
		"file":        fmt.Sprintf("%s/contents/videos/%s/%s.mp4", payload[0], payload[1], payload[2]),
		"filename":    fmt.Sprintf("%s.mp4", payload[2]),
		"type":        payload[3],
		"environment": payload[4],
	}
	bytesRepresentation, err := json.Marshal(jsonStr)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.Post("http://transcoder.roy.solutions/v1/transcode", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	} else {
		return false
	}
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

	// Roy Transcoder job transcode
	replyBtnRoyTranscoderJob := tb.ReplyButton{Text: "/transcoder_job"}
	b.Handle(&replyBtnRoyTranscoderJob, func(m *tb.Message) {
		go func(m *tb.Message) {
			if !healthCheckRoyTranscoder() {
				b.Send(m.Chat, fmt.Sprintf("%s Parece que tem alguma coisa errada no Roy Transcoder 😞 Consegue mandar uma mensagem pros humanos que programaram?", m.Chat.Username))
			} else {
				words := strings.Fields(m.Payload)
				createJobRoyTranscoder(words...)
				b.Send(m.Chat, fmt.Sprintf("%s seu video está sendo processado 😄.", m.Chat.Username))
			}
		}(m)
	})

	// Roy Transcoder job status
	replyBtnRoyTranscoderJobStatus := tb.ReplyButton{Text: "/transcoder_job_status"}
	b.Handle(&replyBtnRoyTranscoderJobStatus, func(m *tb.Message) {
		go func(m *tb.Message) {
			if !healthCheckRoyTranscoder() {
				b.Send(m.Chat, fmt.Sprintf("%s Parece que tem alguma coisa errada no Roy Transcoder 😞 Consegue mandar uma mensagem pros humanos que programaram?", m.Chat.Username))
			} else {
				words := strings.Fields(m.Payload)
				job := checkJobRoyTranscoder(words...)
				b.Send(
					m.Chat, fmt.Sprintf(
						"%s seu video está sendo processado 😄.\n status: %s\n", m.Chat.Username, job))
			}
		}(m)
	})

	// Roy Transcoder healthcheck
	replyBtnRoyTranscoderHealthCheck := tb.ReplyButton{Text: "/transcoder_healthcheck"}
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
	replyBtnRoyTranscoderStatus := tb.ReplyButton{Text: "/transcoder_status"}
	b.Handle(&replyBtnRoyTranscoderStatus, func(m *tb.Message) {
		go func() {
			status := statusRoyTranscoder()
			if status.Video.ID != "" {
				b.Send(m.Chat, fmt.Sprintf("%s O tamanho da fila é %d.\nO video que está sendo processado no momento é %s com id: %s para: %s do tipo: %s 😄", m.Chat.Username, status.QueueSize, status.Video.Filename, status.Video.ID, status.Video.Owner, status.Video.Type))
			} else {
				b.Send(m.Chat, fmt.Sprintf("%s não tem nenhum vídeo no momento sendo processado. 😄", m.Chat.Username))
			}
		}()
	})

	// Roy API tenants
	replyBtnRoyAPITenants := tb.ReplyButton{Text: "/api_tenants"}
	b.Handle(&replyBtnRoyAPITenants, func(m *tb.Message) {
		go func() {
			tenants := getRoyTenants("release")
			b.Send(m.Chat, fmt.Sprintf("Mestre: %s 🙌 \nA quantidade de tenants cadastrados em produção é %d.\nSão eles: \n%v 😄", m.Chat.Username, len(tenants), tenants))
		}()
	})

	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{replyBtnRoyTranscoderHealthCheck},
		[]tb.ReplyButton{replyBtnRoyTranscoderJob},
		[]tb.ReplyButton{replyBtnRoyTranscoderJobStatus},
		[]tb.ReplyButton{replyBtnRoyTranscoderStatus},
		[]tb.ReplyButton{replyBtnRoyAPITenants},
	}

	b.Handle("/oi", func(m *tb.Message) {
		b.Send(m.Chat, "Oi! Eis a lista de coisas que posso te responder:\n\n @Transcoder: healthcheck, status do servidor.", &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
	})

	b.Start()
}
