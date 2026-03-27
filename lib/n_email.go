package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"

	chkr "github.com/tweithoener/checker"
	"github.com/wneessen/go-mail"
)

// EmailArgs defines the arguments for an Email notifier.
type EmailArgs struct {
	SmtpServer string
	User       string
	Password   string
	To         []string
	From       string
	Template   string
}

type emailMaker struct{}

var emailMkr = emailMaker{}

func (emailMaker) Maker() string {
	return "Email"
}

func (emailMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := EmailArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal EmailArgs arguments: %v", err)
	}
	return args, nil
}

func (emailMaker) FromConfig(c chkr.NotifierConfig) (chkr.Notifier, error) {
	args, ok := c.Args.(EmailArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Email arguments")
	}
	if len(args.To) == 0 {
		return nil, fmt.Errorf("Email notifier requires at least one 'To' address")
	}
	var opts []EmailOption
	if args.Template != "" {
		opts = append(opts, WithTemplate(args.Template))
	}
	if args.From != "" {
		opts = append(opts, WithFrom(args.From))
	}
	return Email(args.SmtpServer, args.User, args.Password, args.To, opts...), nil
}

type emailOptions struct {
	from     string
	template *template.Template
}

// EmailOption defines a functional option for the Email notifier.
type EmailOption func(*emailOptions)

const defaultEmailTemplate = `
Check {{.Name}} is in state {{.State}}.
Message: {{.Message}}

{{.Streak}}x since {{.Since.Format "2006-01-02 15:04:05" }}`

// WithTemplate returns an EmailOption that sets a custom email body template.
func WithTemplate(mailTemplate string) EmailOption {
	return func(o *emailOptions) {
		tmpl, err := template.New("email").Parse(mailTemplate)
		if err != nil {
			log.Printf("Can't parse email template: %v", err)
			return
		}
		o.template = tmpl
	}
}

// WithFrom returns an EmailOption that sets the sender address.
func WithFrom(from string) EmailOption {
	return func(o *emailOptions) {
		o.from = from
	}
}

var sendEmailMsg = func(smtpServer string, user string, password string, msg *mail.Msg) error {
	host, portStr, err := net.SplitHostPort(smtpServer)
	if err != nil {
		host = smtpServer
		portStr = "25" // default port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 25
	}

	opts := []mail.Option{
		mail.WithPort(port),
	}

	if user != "" || password != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(user),
			mail.WithPassword(password),
		)
	}

	if port == 465 {
		opts = append(opts, mail.WithSSLPort(false))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return fmt.Errorf("can't create mail client: %v", err)
	}
	return client.DialAndSend(msg)
}

// Email returns a notifier that sends emails via SMTP.
func Email(smtpServer, user, password string, to []string, opts ...EmailOption) chkr.Notifier {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	defaultFrom := user
	if !strings.Contains(user, "@") {
		defaultFrom = fmt.Sprintf("%s@%s", user, hostname)
	}
	if defaultFrom == "" {
		defaultFrom = "checker@" + hostname
	}

	options := &emailOptions{
		from: defaultFrom,
	}
	for _, opt := range opts {
		opt(options)
	}
	if options.template == nil {
		WithTemplate(defaultEmailTemplate)(options)
	}

	return func(ctx context.Context, name string, cs chkr.CheckState) {
		if len(to) == 0 {
			log.Printf("Can't send email: no recipients")
			return
		}

		var body bytes.Buffer
		if err := options.template.Execute(&body, cs); err != nil {
			log.Printf("Can't execute email template: %v", err)
			return
		}

		m := mail.NewMsg()
		if err := m.From(options.from); err != nil {
			log.Printf("Can't set from address: %v", err)
			return
		}
		if err := m.To(to...); err != nil {
			log.Printf("Can't set to addresses: %v", err)
			return
		}
		m.Subject(fmt.Sprintf("[Checker] %s %s", cs.State, name))
		m.SetBodyString(mail.TypeTextPlain, body.String())

		if err := sendEmailMsg(smtpServer, user, password, m); err != nil {
			log.Printf("Can't send email notification: %v", err)
		}
	}
}
