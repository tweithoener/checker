package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	po "github.com/gregdel/pushover"
	chkr "github.com/tweithoener/checker"
)

// PushoverArgs defines the arguments for a Pushover notifier.
type PushoverArgs struct {
	Prefix    string
	App       string
	Recipient string
}

type pushoverMaker struct{}

var pushoverMkr = pushoverMaker{}

func (pushoverMaker) Maker() string {
	return "Pushover"
}

func (pushoverMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := PushoverArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal PushoverArgs arguments: %v", err)
	}
	return args, nil
}

func (pushoverMaker) FromConfig(c chkr.NotifierConfig) (chkr.Notifier, error) {
	args, ok := c.Args.(PushoverArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not PushoverArgs arguments")
	}
	return Pushover(args.Prefix, args.App, args.Recipient), nil
}

// Pushover returns a notifier that sends messages to the Pushover service.
func Pushover(prefix, app, recipient string) chkr.Notifier {
	puApp := po.New(app)
	puRecipient := po.NewRecipient(recipient)
	return func(ctx context.Context, name string, h chkr.History) {
		message := po.NewMessage(fmt.Sprintf("%s%s", prefix, h))
		message.Title = fmt.Sprintf("%s %s", h.State(), name)
		switch h.State() {
		case chkr.Fail:
			message.Priority = po.PriorityHigh
			message.Sound = po.SoundCosmic
		case chkr.Warn:
			message.Priority = po.PriorityHigh
			message.Sound = po.SoundBike
		case chkr.OK:
			message.Priority = po.PriorityNormal
			message.Sound = po.SoundIncoming
		}

		if _, err := puApp.SendMessage(message, puRecipient); err != nil {
			log.Printf("Can't send pushover notification: %c", err)
		}
	}
}
