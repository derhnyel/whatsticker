package handler

import (
	"encoding/json"
	"fmt"
	"log"

	rmq "github.com/adjust/rmq/v4"
	"github.com/deven96/whatsticker/task"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// WebPFormat is the extension of webp
const WebPFormat = ".webp"

// ImageFileSizeLimit limits file sizes to be converted to 2MB(Mebibytes) in Bytes
const ImageFileSizeLimit = 2097000

// VideoFileSizeLimit limits video to be much smaller cuz over 1000KiB(1MiB)
// seems not to animate
const VideoFileSizeLimit = 1024000

// VideoFileSecondsLimit sets video to be less than 7 seconds
// or it no longer animates
const VideoFileSecondsLimit = 7

// Handler interface for multiple message types
type Handler interface {
	// Setup the handler, event and context to reply
	SetUp(client *whatsmeow.Client, event *events.Message, replyTo bool)
	// Validate : ensures the media conforms to some standards
	// also sends message to client about issue
	Validate() error
	// Handle : obtains the message to be sent as response
	Handle(pushTo rmq.Queue) error
}

// Run : the appropriate handler using the event type
func Run(client *whatsmeow.Client, event *events.Message, replyTo bool, convertQueue rmq.Queue, loggingQueue rmq.Queue) {
	var handle Handler
	log.Printf("Running for %s type\n", event.Info.MediaType)
	messageSender := event.Info.Sender.User
	requestTime := event.Info.Timestamp
	isgroupMessage := event.Info.IsGroup
	metric := task.StickerizationMetric{
		InitialMediaLength: 0,
		FinalMediaLength:   0,
		MediaType:          event.Info.MediaType,
		IsGroupMessage:     isgroupMessage,
		MessageSender:      messageSender,
		TimeOfRequest:      requestTime.String(),
		Validated:          false,
	}
	metricBytes, _ := json.Marshal(&metric)
	switch event.Info.MediaType {
	case "image":
		fmt.Println("Using Image Handler")
		handle = &Image{}
	case "video", "gif":
		fmt.Println("Using Video Handler")
		handle = &Video{}
	default:
		responseMessage := &waProto.Message{Conversation: proto.String("Bot currently supports sticker creation from (video/images) only")}
		client.SendMessage(event.Info.Chat, "", responseMessage)
		loggingQueue.PublishBytes(metricBytes)
		return
	}
	handle.SetUp(client, event, replyTo)
	invalid := handle.Validate()
	if invalid != nil {
		log.Printf("%s\n", invalid)
		loggingQueue.PublishBytes(metricBytes)
		return
	}
	if handle.Handle(convertQueue) != nil {
		loggingQueue.PublishBytes(metricBytes)
	}
}

//func RunTest() {
//  metadata.GenerateMetadata("images/converted/STK-20211123-WA0006.webp", "images/metadata/test.exif")
//}
