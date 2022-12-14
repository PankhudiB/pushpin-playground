package main

import (
	"fmt"
	"github.com/fanout/go-gripcontrol"
	pubcontrol "github.com/fanout/go-pubcontrol"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
)

func main() {
	router := gin.Default()

	router.Handle("POST", "/echo", func(context *gin.Context) {
		fmt.Println("Got /echo request on origin server")
		Echo(context.Writer, context.Request)
	})
	router.Handle("POST", "/subscribe", func(context *gin.Context) {
		fmt.Println("Got /subscribe  request on origin server")
		Subscribe(context.Writer, context.Request)
	})
	router.Handle("POST", "/publish", func(context *gin.Context) {
		fmt.Println("About to publish")
		Publish(context.Writer, context.Request)
	})

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Error starting router : ", err.Error())
	}
}

func Echo(writer http.ResponseWriter, request *http.Request) {

	body, _ := ioutil.ReadAll(request.Body)
	fmt.Println("/echo got ==> Body : ", string(body))

	writer.Header().Set("Content-Type", "application/websocket-events")
	writer.Write(body)
}

func Subscribe(writer http.ResponseWriter, request *http.Request) {

	inputBody, _ := ioutil.ReadAll(request.Body)
	fmt.Println("/echo got ==> Body : ", string(inputBody))

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
			&gripcontrol.WebSocketEvent{Type: "TEXT",
				Content: "Initial State"},
		}

		_, err1 := io.WriteString(writer, gripcontrol.EncodeWebSocketEvents(outEvents))
		if err1 != nil {
			fmt.Println("Err writing outEvents to writer: ", err1.Error())
		}
	}
}

func Publish(writer http.ResponseWriter, request *http.Request) {

	data, _ := ioutil.ReadAll(request.Body)
	fmt.Println("Data to be published: ", string(data))

	writer.Header().Set("Content-Type", "application/websocket-events")

	pub := gripcontrol.NewGripPubControl([]map[string]interface{}{
		map[string]interface{}{"control_uri": "http://pushpin:5561"}})

	format := &gripcontrol.WebSocketMessageFormat{
		Content: data}

	item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")

	err := pub.Publish("test", item)

	if err != nil {
		panic("Publish failed with: " + err.Error())
	}

}
