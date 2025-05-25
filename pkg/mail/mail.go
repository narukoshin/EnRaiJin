package mail

import (
	"errors"
	"net"
	"strings"
	"time"

	gomail "gopkg.in/mail.v2"

	"github.com/naruoshin/EnRaiJin/pkg/config"
	p "github.com/naruoshin/EnRaiJin/pkg/proxy/v2"
	s "github.com/naruoshin/EnRaiJin/pkg/structs"
)

var (
	Server	   s.YAMLEmailServer	= config.YAMLConfig.E.Server
	Mail 	   s.YAMLEmailMail		= config.YAMLConfig.E.Mail
	E 		   s.YAMLEmail			= config.YAMLConfig.E

	Message    *gomail.Message		= nil

	// ./ ERROR MESSAGES .\
	// Server
	ErrEmptyServer   = errors.New("'server' field is empty")
	ErrEmptyHost	 = errors.New("'server.host' field is empty")
	ErrEmptyPort	 = errors.New("'server.port' field is empty")
	ErrEmptyEmail	 = errors.New("'server.email' field is empty")
	ErrEmptyPassword = errors.New("'server.password' field is empty")
	// Mail
	ErrEmptyMail    = errors.New("'mail' field is empty")
	ErrEmptyName	= errors.New("'name' field is empty")
	ErrEmptySubject = errors.New("'subject' field is empty")
	ErrEmptyMessage = errors.New("'message' field is empty")
	ErrEmptyRecps   = errors.New("'recipients' field is empty")
)

func Enabled() bool {
	return E != (s.YAMLEmail{})
}

func init(){
	// if Server.Timeout is not set, setting a default value
	if Server.Timeout == "" {
		Server.Timeout = "10s"
	}

	// If Email feature is not enabled, doing nothing
	if Enabled() {
		Message = Init_Headers()
	
		// If proxy is present, 
		// 		adding it to the gomail global variable to tunnel through the proxy.
		if p.IsProxy() {
			dialer, err := p.Dial()
			if err != nil {
				config.CError = err
			}
			gomail.NetDialTimeout = func (network, addr string, timeout time.Duration) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		}
	}
}

func _recipientParsing() string {
	switch Mail.Recipients.(type) {
	case []interface{}: {
		recipients := Mail.Recipients.([]interface{})
		combine := []string{}
		for _, recipient := range recipients {
			combine = append(combine, recipient.(string))
		}
		return strings.Join(combine, ",")
	}
	case string: {
		return Mail.Recipients.(string)
	}
	}
	return ""
}

func Init_Headers() *gomail.Message {
	message := gomail.NewMessage()
	// Extracting recipients
	recipients := strings.Split(_recipientParsing(), ",")

	message.SetHeader("From", Server.Email)
	message.SetHeader("To", recipients...)
	message.SetHeader("Subject", Mail.Subject)
	message.SetHeader("Content-Type", "text/plain; charset=\"UTF-8\"")
	message.SetHeader("Content-Transfer-Encoding", "8bit")
	message.SetHeader("X-Mailer", "EnRaiJin")
	message.SetHeader("X-Priority", "1")
	message.SetHeader("X-MSMail-Priority", "Normal")
	message.SetHeader("MIME-Version", "1.0")
	message.SetHeader("X-OriginalArrivalTime", time.Now().Format(time.RFC1123Z))
	return message
}


func Set_Password(password string) {
	// If Email feature is not enabled, doing nothing
	if Enabled() {
		Mail.Message = strings.Replace(Mail.Message, "<password>", password, -1)
		Message.SetBody("text/plain", Mail.Message)
	}
}

func Send() error {
	// If Email feature is not enabled, doing nothing
	if Enabled() {
		dialer := gomail.NewDialer(Server.Host, Server.Port, Server.Email, Server.Password)
		timeout, err := time.ParseDuration(Server.Timeout)
		if err != nil {
			return err
		}
		dialer.Timeout = timeout
		err = dialer.DialAndSend(Message)
		if err != nil {
			return err
		}
	}
	return nil
}


// To ensure that the tool is able to send an email
// We need to test the connection first.
func Ping() error {
	dialer := gomail.NewDialer(Server.Host, Server.Port, Server.Email, Server.Password)
	timeout, err := time.ParseDuration(Server.Timeout)
	if err != nil {
		return err
	}
	dialer.Timeout = timeout
	sender, err := dialer.Dial()
	if err != nil {
		return err
	}
	defer sender.Close()
	return nil
}

func Test() error {
	var err error
	if Enabled() {
		// Validating
		// Checking if Server is not empty
		if Server == (s.YAMLEmailServer{}) {
			return ErrEmptyServer
		}
		if Server.Host == "" {
			return ErrEmptyHost
		}
		if Server.Port == 0 {
			return ErrEmptyPort
		}
		if Server.Email == "" {
			return ErrEmptyEmail
		}
		if Server.Password == "" {
			return ErrEmptyPassword
		}
		// Checking if Mail is not empty
		if Mail == (s.YAMLEmailMail{}) {
			return ErrEmptyMail
		}
		if Mail.Subject == "" {
			return ErrEmptySubject
		}
		if Mail.Name == "" {
			return ErrEmptyName
		}
		if Mail.Message == "" {
			return ErrEmptyMessage
		}
		if Mail.Recipients == nil {
			return ErrEmptyRecps
		}
		// Ping test
		err = Ping()
		if err != nil {
			return err
		}
	}
	return nil
}