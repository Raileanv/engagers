package main

import (
	"encoding/json"
	"fmt"
	"github.com/engagers/models"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"time"
)

type Client struct {
	hub       *Hub
	sessionId int
	conn      *websocket.Conn
	send      chan []byte
}

type Message struct {
	EventType  string      `json:"event_type"`
	Data       interface{} `json:"data"`
	SelfSended bool
	Client     *Client
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetPingHandler(nil)
	c.conn.SetPongHandler(nil)
	c.conn.SetCloseHandler(nil)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		mes := Message{Client: c, SelfSended: false}

		json.Unmarshal(message, &mes)
		fmt.Println("first", mes)
		processMessage(&mes)
		fmt.Println("second", mes)

		c.hub.broadcast <- mes
	}
}

type statistics struct {
	Amount  int
	Answer  string
	Correct bool
}

func calcResultsOfQuiz(m *Message, quiz_id interface{}) {
	var stat []statistics
	session_id := m.Client.sessionId

	DB.Raw(`SELECT count(answer) amount, answer, correct from quiz_answers q
	join answers a
	on a.id = q.answer_id
	where q.quiz_id = ?
	and q.session_id = ?
	group by answer, correct`, quiz_id, session_id).Scan(&stat)

	statJson, _ := json.Marshal(stat)

	m.EventType = "statistics"
	m.Data = string(statJson)

	time.Sleep(time.Second * 18)

	m.Client.hub.broadcast <- *m
}

func processMessage(m *Message) *Message {
	switch m.EventType {
	case "chat_message":
	case "slide":
	case "whiteboard":
	case "start_quiz":
		data, ok := m.Data.(map[string]interface{})
		id := data["quiz_id"]
		if ok {
			m.EventType = "quiz"
			quiz := models.Quiz{}

			DB.Preload("Answers").Find(&quiz, id)
			jstr, _ := json.Marshal(quiz)
			m.Data = string(jstr)
		}
		go calcResultsOfQuiz(m, id)
	case "ping_message":
		m.EventType = "pong_message"
	case "quiz_answer":
		data, ok := m.Data.(map[string]interface{})

		id, _ := strconv.ParseInt(data["quiz_id"].(string), 10, 32)

		quizAnswer := models.QuizAnswer{SessionID: uint(m.Client.sessionId), QuizID: uint(id)}
		if ok {
			quizAnswer.AnswerID = uint(data["answer_id"].(float64))
			quizAnswer.UserID = models.CurrentUser.ID
		}
		DB.Save(&quizAnswer)
	}

	return m
}

func (c *Client) writePump() {

	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, _ := <-c.send:

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
