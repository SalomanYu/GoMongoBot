package main

import (
	"os"
	"fmt"
	"github.com/SalomanYu/GoMongoBot/tasker"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var token = os.Getenv("GOMONGOTOKEN")
// const token = "5750682049:AAGAUcHSqXuyo5DM0LWmCW8G08MrkDO3WIw"

func main(){
	bot, err := tgbotapi.NewBotAPI(token)
	tasker.CheckErr(err)
	
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil{
			continue
		}

		if update.Message.IsCommand(){
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command(){
			case "start":
				fallthrough
			case "help":
				msg.Text = "Все предельно просто! Я знаю несколько команд:\n\n1.Напиши `all` -- чтобы увидеть список задач\n2.Напиши `add Текст задачи` -- чтобы создать задачу\n3.Напиши `done номер задачи` -- чтобы закрыть задачу\n4.Напиши: `unfinished` -- чтобы посмотреть список невыполненных задач\n5.Напиши: `finished` -- чтобы посмотреть список выполненных задач\n6.Напиши `drop номер задачи` -- чтобы удалить ее из списка задач"
			default:
				msg.Text = "Я не знаю таких команд"
				msg.ReplyToMessageID = update.Message.MessageID
			}
			if _, err := bot.Send(msg); err != nil{
				panic(err)
			}
			continue
		}
		msg := checkUserChoise(update)
		if _, err := bot.Send(msg); err != nil{
			panic(err)
		}
	} 
}

func checkUserChoise(update tgbotapi.Update)  tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	command := strings.Split(update.Message.Text, " ")[0]
	userText := strings.Replace(update.Message.Text, command, "", 1)
	userId := update.Message.Chat.ID

	switch strings.ToLower(command){
	case "add":
		task := tasker.Task{
			ID: primitive.NewObjectID(),
			Text: strings.Title(userText),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Completed: false,
		}
		err := tasker.CreateTask(&task, userId)	
		tasker.CheckErr(err)
		msg.Text = "Успешно добавили задачу!"
		fallthrough

	case "all":
		tasks, err := tasker.GetAll(userId)
		if err == mongo.ErrNoDocuments{
			msg.Text = "У вас пока нет заметок. Чтобы создать новую заметку напишите мне `add и текст заметки`"
		} else {
			msg.Text = formatTasks(tasks)
		}
	
	case "unfinished":
		tasks, err := tasker.GetUnfinished(userId)
		tasker.CheckErr(err)
		msg.Text = formatTasks(tasks)
	
	case "finished":
		tasks, err := tasker.GetFinished(userId)
		tasker.CheckErr(err)
		msg.Text = formatTasks(tasks)
	
	case "drop":
		taskNum, err:= strconv.Atoi(strings.TrimSpace(userText))
		tasker.CheckErr(err)
		err = tasker.DropTask(taskNum, userId)
		if err != nil{
			msg.Text = "Мы не нашли такой задачи"
		}else {
			msg.Text = "Задача успешно удалена! Можете проверить командой all"
		}

	case "done":
		taskNum, err := strconv.Atoi(strings.TrimSpace(userText))
		tasker.CheckErr(err)
		err = tasker.CompleteTask(taskNum, userId)
		if err != nil{
			msg.Text = "Не удалось обновить статус задачи"
		} else {
			msg.Text = "Задача успешно закрыта! Можете проверить командой all"
		}

	default:
		msg.Text = "Я не знаю такой команды"
	}
	return msg
}

func formatTasks(tasks []*tasker.Task) string{
	formatedTasks := []string{}
	for i, v := range tasks{
		task := fmt.Sprintf("%d. %s", i+1, v.Text)
		if v.Completed{	
			task += "✅" 
		}
		formatedTasks = append(formatedTasks, task)
	}
	return strings.Join(formatedTasks, "\n")
}
