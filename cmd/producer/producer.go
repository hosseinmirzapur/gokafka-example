package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/hosseinmirzapur/gokafka-example/pkg/models"
)

const (
	ProducerPort    = ":8080"
	KafkaServerAddr = "localhost:9092"
	KafkaTopic      = "notifications"
)

var ErrUserNotFoundInProducer = errors.New("user not found")

// ================== Helper Functions ==================

func findUserByID(id int, users []models.User) (*models.User, error) {
	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, ErrUserNotFoundInProducer
}

func getIDFromRequest(formValue string, ctx *gin.Context) (int, error) {
	id, err := strconv.Atoi(formValue)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ID from form value %s: %w", formValue, err)
	}

	return id, nil
}

// ================== Kafka Related Functions ==================
func sendKafkaMessage(producer sarama.SyncProducer, users []models.User, ctx *gin.Context, fromID, toID int) error {
	message := ctx.PostForm("message")

	fromUser, err := findUserByID(fromID, users)
	if err != nil {
		return err
	}

	toUser, err := findUserByID(toID, users)
	if err != nil {
		return err
	}

	notification := models.Notification{
		From:    fromUser.Name,
		To:      toUser.Name,
		Message: message,
	}

	notifJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: KafkaTopic,
		Key:   sarama.StringEncoder(strconv.Itoa(fromID)),
		Value: sarama.StringEncoder(notifJSON),
	}

	_, _, err = producer.SendMessage(msg)
	return err
}

func sendMessageHandler(producer sarama.SyncProducer, users []models.User) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fromID, err := getIDFromRequest(ctx.PostForm("fromID"), ctx)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		toID, err := getIDFromRequest(ctx.PostForm("toID"), ctx)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = sendKafkaMessage(producer, users, ctx, fromID, toID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "notification sent successfully"})
	}
}

func setupProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{KafkaServerAddr},
		config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup producer: %w", err)
	}
	return producer, nil
}

func main() {
	users := []models.User{
		{ID: 1, Name: "Emma"},
		{ID: 2, Name: "Bruno"},
		{ID: 3, Name: "Rick"},
		{ID: 4, Name: "Lena"},
	}

	producer, err := setupProducer()
	if err != nil {
		log.Fatalf("failed to initialize producer: %v", err)
	}
	defer producer.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/send", sendMessageHandler(producer, users))
	fmt.Printf("Kafka PRODUCER 📨 started at http://localhost%s\n", ProducerPort)

	if err := router.Run(ProducerPort); err != nil {
		log.Printf("failed to run the server: %v", err)
	}
}
