package tgbotroute

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Router struct {
	path   map[string]interface{}
	fromtg chan *tgbotapi.Message
	totg   chan tgbotapi.MessageConfig
	token  string
}

func tgbot(fromtg chan *tgbotapi.Message, totg chan tgbotapi.MessageConfig, token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			if update.Message != nil { // ignore any non-Message Updates

				fromtg <- update.Message

			}
		}
	}()

	go func() {
		for data := range totg {
			bot.Send(data)
		}
	}()

}

//func Cathandle(msg *tgbotapi.Message) string {
//	s := fmt.Sprintln("cat say", msg.Text)
//	return s
//}

//func Doghandle(msg *tgbotapi.Message) string {
//	s := fmt.Sprintln("dog say", msg.Text)
//	return s
//}

func (router *Router) Handle(s string, f func(msg *tgbotapi.Message) string) {
	router.path[s] = f
}

func MakeHandler(token string) *Router {
	r := &Router{
		path:   make(map[string]interface{}),
		fromtg: make(chan *tgbotapi.Message),
		totg:   make(chan tgbotapi.MessageConfig),
		token:  token,
	}
	return r
}

func (router *Router) Work(msg *tgbotapi.Message) {
	command := strings.Split(msg.Text, " ")[0]
	if val, ok := router.path[command].(func(*tgbotapi.Message) string); ok {
		nmsg := tgbotapi.NewMessage(msg.Chat.ID, val(msg))
		router.totg <- nmsg

	} else if val, ok := router.path["/ontext"].(func(*tgbotapi.Message) string); ok {
		nmsg := tgbotapi.NewMessage(msg.Chat.ID, val(msg))
		router.totg <- nmsg
	} else {

		nmsg := tgbotapi.NewMessage(msg.Chat.ID, "No command")
		router.totg <- nmsg
	}

}

func (router *Router) Listen() {
	go tgbot(router.fromtg, router.totg, router.token)
	for tv := range router.fromtg {
		go router.Work(tv)
	}
}

//func main() {
//	a := MakeHandler()
//	a.token = "TOKEN"
//	a.Handle("cat", Cathandle)
//	a.Handle("dog", Doghandle)
//	a.Listen()
//
//}
