package main

import (
	"fmt"
	"github.com/fanout/go-gripcontrol"
	pubcontrol "github.com/fanout/go-pubcontrol"
	"github.com/gin-gonic/gin"
	zmq "github.com/pebbe/zmq4"
	"io"
	"io/ioutil"
	"net/http"
)

func main() {
	router := gin.Default()

	zctx, err := zmq.NewContext()
	s, err := zctx.NewSocket(zmq.PUB)
	err = s.Connect("tcp://docker.for.mac.localhost:5562")
	if err != nil {
		fmt.Println("Err binding socket : ", err.Error())
	}

	router.Handle("POST", "/publish-on-zmq", func(context *gin.Context) {
		fmt.Println("Got /publish-on-zmq request on origin server")
		PublishOverZMQ(s, context.Request)
	})
	router.Handle("POST", "/api/go-app/subscribe", func(context *gin.Context) {
		fmt.Println("Got /subscribe  request on origin server")
		Subscribe(context.Writer, context.Request)
	})
	router.Handle("POST", "/publish", func(context *gin.Context) {
		fmt.Println("About to publish")
		Publish(context.Writer, context.Request)
	})

	err1 := http.ListenAndServe(":8080", router)
	if err1 != nil {
		fmt.Println("Error starting router : ", err1.Error())
	}
}

func PublishOverZMQ(s *zmq.Socket, request *http.Request) {
	data, _ := ioutil.ReadAll(request.Body)
	fmt.Println("Data to be published: ", string(data))

	fmt.Println("/publish-on-zmq got ==> Body : ")

	sent, err := s.Send("test", zmq.SNDMORE)
	if err != nil {
		fmt.Println("Err sending : ", err.Error())
	}
	fmt.Printf("Sent 1 : ", sent)

	formattedMessage := `J{"formats": {"ws-message": {"content": "` + string(data) + `\n"}}}`
	sent2, err2 := s.Send(formattedMessage, 0)
	if err2 != nil {
		fmt.Println("Err sending : ", err2.Error())
	}
	fmt.Printf("Sent 2 : ", sent2)
}

func Subscribe(writer http.ResponseWriter, request *http.Request) {

	inputBody, _ := ioutil.ReadAll(request.Body)
	fmt.Println("/subscribe got ==> Body : ", string(inputBody))

	writer.Header().Set("Content-Type", "application/websocket-events")

	inEvents, err := gripcontrol.DecodeWebSocketEvents(string(inputBody))
	if err != nil {
		panic("Failed to decode WebSocket inEvents: " + err.Error())
	}

	fmt.Printf("IN events.length: %d", len(inEvents))
	fmt.Printf("IN events[0]: %+v", inEvents[0])

	shouldSubscribe := false

	if inEvents[0].Type == "OPEN" {
		fmt.Println("OPEN event came to origin server ! ")
		writer.Header().Set("Sec-WebSocket-Extensions", `grip; message-prefix=""`)
		shouldSubscribe = true
	}

	fmt.Println("should subscribe ? ", shouldSubscribe)

	if shouldSubscribe {
		wsControlMessage, err := gripcontrol.WebSocketControlMessage("subscribe",
			map[string]interface{}{"channel": "test"})
		if err != nil {
			panic("Unable to create control message: " + err.Error())
		}

		outEvents := []*gripcontrol.WebSocketEvent{
			&gripcontrol.WebSocketEvent{Type: "OPEN"},
			&gripcontrol.WebSocketEvent{Type: "TEXT",
				Content: "c:" + wsControlMessage},
		}

		n, err1 := io.WriteString(writer, gripcontrol.EncodeWebSocketEvents(outEvents))
		if err1 != nil {
			fmt.Println("Err writing outEvents to writer: ", err1.Error())
			return
		}
		fmt.Println("Wrote : ", n)
	}

}

func Publish(writer http.ResponseWriter, request *http.Request) {

	data, _ := ioutil.ReadAll(request.Body)
	fmt.Println("Data to be published: ", string(data))

	writer.Header().Set("Content-Type", "application/websocket-events")

	pub := gripcontrol.NewGripPubControl([]map[string]interface{}{
		map[string]interface{}{"control_uri": "http://docker.for.mac.localhost:5561"}})

	format := &gripcontrol.WebSocketMessageFormat{
		Content: data}

	item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")

	err := pub.Publish("test", item)

	if err != nil {
		panic("Publish failed with: " + err.Error())
	}

}
