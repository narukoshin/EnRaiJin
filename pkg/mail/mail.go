package mail

import (
	"errors"
	"net"
	"strings"
	"time"
	"fmt"

	gomail "gopkg.in/mail.v2"

	"EnRaiJin/pkg/config"
	p "EnRaiJin/pkg/proxy/v2"
	_ "EnRaiJin/pkg/config"
	s "EnRaiJin/pkg/structs"
)

var (
	Server	   s.YAMLEmailServer	= config.YAMLConfig.E.Server
	Mail 	   s.YAMLEmailMail		= config.YAMLConfig.E.Mail
	E 		   s.YAMLEmail			= config.YAMLConfig.E

	Message    *gomail.Message		= nil

	ErrEmptyName	= errors.New("name field is empty")
	ErrEmptySubject = errors.New("subject field is empty")
	ErrEmptyMessage = errors.New("message field is empty")
	ErrEmptyrecps   = errors.New("no recipients added")
)

func Enabled() bool {
	return E != (s.YAMLEmail{})
}

func init(){
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
	Mail.Message = strings.Replace(Mail.Message, "<password>", password, -1)
	Message.SetBody("text/plain", Mail.Message)
}

func Send() error {
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
	return nil
}


// To ensure that the tool is able to send an email
// We need to test the connection first.
func Ping() error {
	if Enabled() {
		dialer := gomail.NewDialer(Server.Host, 465, Server.Email, Server.Password)
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
	}
	return nil
}