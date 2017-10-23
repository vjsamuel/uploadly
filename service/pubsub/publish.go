package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"log"
	"github.com/vjsamuel/uploadly/service/common"
	"io"
	"fmt"
	"io/ioutil"
)

type PubSub struct {
	topic     *pubsub.Topic
	client    *pubsub.Client
	ctx       context.Context
}

func NewPubSub(project, topic string, ctx context.Context) *PubSub {
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		log.Printf("Client connection failed with error: %v\n", err)
		return nil
	}

	t := client.Topic(topic)
	if t == nil {
		t, err = client.CreateTopic(ctx, topic)
		if err != nil {
			log.Printf("Topic creation failed with error: %v\n", err)
			return nil
		}
	}

	return &PubSub{client: client, topic: t, ctx: ctx}
}

func (p *PubSub) Publish(holder common.Holder) (error){
	reader, ok := holder.Object.(io.Reader)
	if !ok {
		return fmt.Errorf("Unable to get Reader for input object")
	}

	bytes, err := ioutil.ReadAll(reader)
	if !ok {
		return fmt.Errorf("Unable to get bytes from reader due to error: %v", err)
	}

	message := pubsub.Message{
		Attributes: map[string]string {
			"name": holder.File,
			"profile": holder.User.Profile,
			"contentType": holder.ContentType,
		},
		Data: bytes,
	}

	result := p.topic.Publish(p.ctx, &message)
	_, err = result.Get(p.ctx)

	if err != nil {
		log.Printf("Message publish failed with error: %v\n", err)
		return err
	}

	return nil
}



