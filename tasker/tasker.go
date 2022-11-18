package tasker

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // В пакет заложены объекты документов MongoDB
	"go.mongodb.org/mongo-driver/mongo"          // Основной пакет mongoDB для создания БД, коллекций и подключения к БД
	"go.mongodb.org/mongo-driver/mongo/options"  // В опциях указываем местоположение БД
)


var collection *mongo.Collection
var ctx = context.TODO()

type Task struct{
	ID			primitive.ObjectID 	`bson:"_id"`
	CreatedAt	time.Time 			`bson:"created_at"`
	UpdatedAt	time.Time 			`bson:"updated_at"`
	Text		string 				`bson:"text"`
	Completed	bool 				`bson:"completed"`
}

func CheckErr(err error){
	if err != nil{
		panic(err)
	}
}


func initUser(userId int64){
	// Метод вызывается сам по себе, видимо это такая фича Go

	// Подключаемся к Монго
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	CheckErr(err)
	
	// Проверяем наше подключение
	err = client.Ping(ctx, nil)
	CheckErr(err)

	// Создаем Базу данных tasker и внутри создаем коллекцию tasks
	
	collection = client.Database("tasker").Collection(strconv.FormatInt(userId, 10))
}


func CreateTask(task *Task, userId int64) error{
	initUser(userId)
	// InsertOne возвращает айди документа и ошибку. Айди нам не нужен
	_, err := collection.InsertOne(ctx, task) 
	return err
}

func GetAll(userId int64) ([]*Task, error) {
	initUser(userId)
	filter := bson.D{{}} // Тип bson.D - это документ типа BSON. {{}} Показывает что мы хотим взять всё
	return filterTasks(filter)
}

func filterTasks(filter interface{}) ([]*Task, error){
	var tasks []*Task

	cur, err := collection.Find(ctx, filter)
	CheckErr(err)

	for cur.Next(ctx){
		var t Task
		err := cur.Decode(&t)
		CheckErr(err)

		tasks = append(tasks, &t)
	}
	if err := cur.Err(); err != nil{
		return tasks, err
	}

	cur.Close(ctx)
	if len(tasks) == 0{
		return tasks, mongo.ErrNoDocuments
	}
	return tasks, nil
}

func CompleteTask(taskNum int, userId int64) error {
	tasks, err := GetAll(userId)
	CheckErr(err)

	for i, v := range tasks{
		if i+1 == taskNum{
			filter := bson.D{primitive.E{
				Key: "_id",
				Value: v.ID,
			}}
			update := bson.D{primitive.E{
				Key: "$set",
				Value: bson.D{
					primitive.E{
						Key: "completed",
						Value: true,
					},
				},
			}}
			t := &Task{}
			return collection.FindOneAndUpdate(ctx, filter, update).Decode(t)
		}
	}	
	return errors.New("Failed to update task")
}

func GetUnfinished(userId int64) ([]*Task, error) {
	initUser(userId)
	filter := bson.D{primitive.E{
		Key: "completed",
		Value: false,
	}}
	return filterTasks(filter)
}

func GetFinished(userId int64) ([]*Task, error) {
	initUser(userId)
	filter := bson.D{primitive.E{
		Key: "completed",
		Value: true,
	}}
	return filterTasks(filter)
}

func DropTask(taskNum int, userId int64) error {
	tasks, err := GetAll(userId)
	CheckErr(err)

	for i, v := range tasks{
		if i+1 == taskNum{
			filter := bson.D{bson.E{
				Key: "_id",
				Value: v.ID,
			}}
			_, err := collection.DeleteOne(ctx, filter)
			CheckErr(err)
			return nil
		}
	}
	return errors.New("No tasks were deleted")

}