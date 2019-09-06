package main

import (
	"encoding/json"
	"github.com/Raileanv/engagers/models"
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
	SelfSended bool `json:"self_sended"`
	Client     *Client `json:"client"`
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

		processMessage(&mes)

		c.hub.broadcast <- mes
	}
}

type statistics struct {
	Amount  int `json:"amount"`
	Answer  string `json:"answer"`
	Correct bool `json:"correct"`
}

func calcResultsOfQuiz(m *Message, quiz_id interface{}) {
	var stat []statistics
	quiz := &models.Quiz{}
	session_id := m.Client.sessionId
	id, ok := quiz_id.(string)
	if !ok {
		m.EventType = "BadRequesr"
		m.Data = "something went wrong, check format of params"

		m.Client.hub.broadcast <- *m
		return
	}
	quiz_id, _ = strconv.ParseInt(id, 10, 32)

	time.Sleep(time.Second * 18)

	DB.Raw(`SELECT count(answer) amount, answer, correct from quiz_answers q
	join answers a
	on a.id = q.answer_id
	where q.quiz_id = ?
	and q.session_id = ?
	group by answer, correct`, quiz_id, session_id).Scan(&stat)

	DB.First(&quiz, quiz_id)
	type resp struct {
		Answers []statistics `json:"answers"`
		Question string `json:"question"`
	}

	r := resp{
		stat,
		quiz.Question,
	}

	m.EventType = "statistics"
	m.Data = r

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

			id, ok := id.(string)

			if !ok {
				m.EventType = "BadRequesr"
				m.Data = "something went wrong, check format of params"

				return m
			}

			quiz_id, _ := strconv.ParseInt(id, 10, 32)

			DB.Preload("Answers").First(&quiz, quiz_id)

			m.Data = quiz
		}
		go calcResultsOfQuiz(m, id)
	case "ping_message":
		m.EventType = "pong_message"
	case "quiz_answer":
		data, ok := m.Data.(map[string]interface{})

		//id, _ := strconv.ParseUint(data["quiz_id"].(string), 10, 32)
		//answer_id, _ := strconv.ParseUint(data["answer_id"].(string), 10, 32)

		id, ok := data["quiz_id"].(string)

		if !ok {
			m.EventType = "BadRequesr"
			m.Data = "something went wrong, check format of params"

			return m
		}

		quiz_id, _ := strconv.ParseInt(id, 10, 32)

		a_id, ok := data["answer_id"].(string)

		if !ok {
			m.EventType = "BadRequesr"
			m.Data = "something went wrong, check format of params"

			return m
		}

		answer_id, _ := strconv.ParseInt(a_id, 10, 32)



		quizAnswer := models.QuizAnswer{SessionID: uint(m.Client.sessionId), QuizID: uint(answer_id)}
		if ok {
			quizAnswer.AnswerID = uint(answer_id)
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
