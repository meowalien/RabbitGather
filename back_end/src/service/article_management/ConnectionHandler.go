package article_management

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"rabbit_gather/util"
	"time"
)

type SocketEvent uint16

const (
	CloseEvent SocketEvent = iota
	PingEvent
	TextMessageEvent
	BinaryMessageEvent
	OpenEvent
	PongEvent
	ErrorEvent
)

type ConnectionHandler struct {
	OnOpenEvent          func(connectionID int64)
	OnTextMessageEvent   func(...TextMessage)
	OnBinaryMessageEvent func(message ...*RawMessage)
	OnPongEvent          func(message ...*RawMessage)
	OnPingEvent          func(message ...*RawMessage)
	OnCloseEvent         func(message ...*RawMessage)
	//OnPongEvent func(message *RawMessage)
	*websocket.Conn
	//Handlers           map[SocketEvent]Handler
	sentMessageChannel chan RawMessage

	maxMessageSize int64
	pongWait       time.Duration
	pingPeriod     time.Duration
	writeWait      time.Duration
}

const (
	//Time allowed to write a message to the peer.
	writeWait_defult = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait_defult = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod_defult = (pongWait_defult * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize_defult = 512
)

func DefaultConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{
		maxMessageSize: maxMessageSize_defult,
		pongWait:       pongWait_defult,
		pingPeriod:     pingPeriod_defult,
		writeWait:      writeWait_defult,
	}
}

type Handler func(...interface{})

func (w *ConnectionHandler) SentRawMessage(message *RawMessage) {
	w.sentMessageChannel <- *message
}

//func (w *ConnectionHandler) Emit(ev SocketEvent, v ...interface{}) {
//	h, e := w.Handlers[ev]
//	if !e {
//		return
//	}
//	h(v...)
//}

func (c *ConnectionHandler) Initialize() {
	go c.readPump()
	go c.writePump()
}

func (c *ConnectionHandler) readPump() {
	c.SetReadLimit(c.maxMessageSize)
	err := c.SetReadDeadline(time.Now().Add(c.pongWait))
	if err != nil {
		panic(err.Error())
	}
	c.SetPongHandler(func(string) error {
		e := c.SetReadDeadline(time.Now().Add(c.pongWait))
		if e != nil {
			return e
		}
		return nil
	})
	for {
		messageType, reader, err := c.NextReader()
		if err != nil {
			if closeError, ok := err.(*websocket.CloseError); ok {
				errorCode := closeError.Code
				switch errorCode {
				case websocket.CloseNormalClosure: //1000
				case websocket.CloseGoingAway: //1001
				//case websocket.CloseProtocolError: //1002
				//case websocket.CloseUnsupportedData: //1003
				//case websocket.CloseNoStatusReceived: //1005
				//case websocket.CloseAbnormalClosure: //1006
				//case websocket.CloseInvalidFramePayloadData: //1007
				//case websocket.ClosePolicyViolation: //1008
				//case websocket.CloseMessageTooBig: //1009
				//case websocket.CloseMandatoryExtension: //1010
				//case websocket.CloseInternalServerErr: //1011
				//case websocket.CloseServiceRestart: //1012
				//case websocket.CloseTryAgainLater: //1013
				//case websocket.CloseTLSHandshake: //1015
				default:
					log.WARNING.Printf("Close with code:%d, %s", errorCode, util.WebsocketCloseCodeNumberToString(errorCode))
				}
			} else {
				log.DEBUG.Println("NextReader error: ", err.Error())
			}
			break
		}
		message := &RawMessage{
			//MessageType: messageType,
			Reader: reader,
		}
		switch messageType {
		case websocket.CloseMessage:
			close(c.sentMessageChannel)
			if c.OnCloseEvent != nil {
				c.OnCloseEvent(message)
			}
			//c.Emit(CloseEvent, message)
		case websocket.PingMessage:
			if c.OnPingEvent != nil {
				c.OnPingEvent(message)
			}
			//c.Emit(PingEvent, message)
		case websocket.PongMessage:
			if c.OnPongEvent != nil {
				c.OnPongEvent(message)
			}
			//c.Emit(PongEvent, message)
		default:
			switch messageType {
			case websocket.TextMessage:
				if c.OnTextMessageEvent != nil {
					c.OnTextMessageEvent(TextMessage(message))
				}
			case websocket.BinaryMessage:
				if c.OnBinaryMessageEvent != nil {
					c.OnBinaryMessageEvent(message)
				}
			default:
				log.ERROR.Println("Unknown event")
			}
		}
	}
}

func (c *ConnectionHandler) writePump() {
	c.sentMessageChannel = make(chan RawMessage, 256)
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.sentMessageChannel:
			if message.SentMessageErrorCallback == nil {
				message.SentMessageErrorCallback = func(err error) {
					log.DEBUG.Println("error: ", err.Error())
				}
			}
			err := c.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err != nil {
				log.DEBUG.Println("error when SetWriteDeadline")
				message.SentMessageErrorCallback(err)
			}
			if !ok {
				// The hub closed the channel.
				err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server close the write channel"))
				if err != nil && err != websocket.ErrCloseSent {
					log.DEBUG.Println("error when WriteMessage")
					message.SentMessageErrorCallback(err)
				}
				return
			}

			writer, err := c.NextWriter(websocket.TextMessage)
			if err != nil {
				log.DEBUG.Println("error when NextWriter")
				message.SentMessageErrorCallback(err)
				return
			}
			_, err = io.Copy(writer, message.Reader)
			if err != nil {
				log.DEBUG.Println("error when Copy")
				message.SentMessageErrorCallback(err)
			}

			if err := writer.Close(); err != nil {
				log.DEBUG.Println("error when Close writer: ", err.Error())
				message.SentMessageErrorCallback(err)
				return
			}
			if message.AfterSentCallback != nil {
				message.AfterSentCallback()
			}
		case <-ticker.C:
			err := c.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err != nil {
				log.DEBUG.Println("error when SetWriteDeadline")
			}
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.DEBUG.Println("error when SetWriteDeadline")
				return
			}
		}
	}
}

func (w *ConnectionHandler) JoinGroup(s string) {

}

func (c *ConnectionHandler) Close() error {
	close(c.sentMessageChannel)
	return nil
}

func (w *ConnectionHandler) SentTextMessage(s string) {
	w.SentRawMessage(&RawMessage{Reader: bytes.NewReader([]byte(s))})
}

func (w *ConnectionHandler) SentMessage(event interface{}, sentMessageErrorCallback func(err error), afterSentCallback func()) error {
	b, err := json.Marshal(event)
	if err != nil {
		log.ERROR.Println("error when marshal Message")
		return err
	}
	if sentMessageErrorCallback == nil {
		sentMessageErrorCallback = func(err error) {
			log.ERROR.Println("fail to sent ArticleErrorEvent")
		}
	}

	w.SentRawMessage(&RawMessage{
		Reader:                   bytes.NewReader(b),
		SentMessageErrorCallback: sentMessageErrorCallback,
		AfterSentCallback:        afterSentCallback,
	})
	return nil
}
