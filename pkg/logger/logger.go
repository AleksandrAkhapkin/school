package logger

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/tarasova-school/internal/types/config"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	chatID    = ""
	channelID = ""
	infoMsg   = "INFO"
	errMsg    = "ERROR"
	token     = ""
)

func NewLogger(t *config.Telegram) error {

	chatID = t.ChatID
	channelID = t.ChannelID
	token = t.TelegramToken

	return nil
}

var logger = zerolog.New(os.Stdout)
var Debug = true

func CheckDebug() {
	if Debug {
		infoMsg += "-DEBUG"
		errMsg += "-DEBUG"
	} else {
		chatID = channelID
	}
}

func LogError(err error) {

	logger.Err(err).Send()
	SendError(err)
}

func LogInfo(msg string) {
	logger.Info().Msg(msg)
	SendMessage(msg)
}

func LogFatal(err error) {

	if unwrap := errors.Unwrap(err); unwrap != nil {
		err = unwrap
	}

	SendError(err)
	logger.Fatal().Err(err).Send()
}
func SendError(err error) {

	if unwrap := errors.Unwrap(err); unwrap != nil {
		err = unwrap
	}

	url := makeURLSendMessage(errMsg, err.Error())
	if err := send(url); err != nil {
		logger.Err(err).Send()
	}
}

func SendMessage(msg string) {
	url := makeURLSendMessage(infoMsg, msg)
	if err := send(url); err != nil {
		logger.Err(err).Send()
	}
}

func makeURLSendMessage(typeMsg, text string) string {

	text = fmt.Sprintf("%s [%s]: %s", typeMsg, time.Now().Format("2006-01-02T15:04:05"), text)
	str := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		token, chatID, text)
	return strings.ReplaceAll(str, " ", "+")
}

func send(urlForSend string) error {
	req, err := http.NewRequest(http.MethodPost, urlForSend, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Err(err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}
	return nil
}
