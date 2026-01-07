package mail

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	gomail "gopkg.in/mail.v2"

	"github.com/narukoshin/EnRaiJin/v2/pkg/config"
	p "github.com/narukoshin/EnRaiJin/v2/pkg/proxy"
	s "github.com/narukoshin/EnRaiJin/v2/pkg/structs"
)

var (
	Server s.YAMLEmailServer = config.YAMLConfig.E.Server
	Mail   s.YAMLEmailMail   = config.YAMLConfig.E.Mail
	E      s.YAMLEmail       = config.YAMLConfig.E

	Message *gomail.Message = nil

	// ./ ERROR MESSAGES .\
	// Server
	ErrEmptyServer   = errors.New("'server' field is empty")
	ErrEmptyHost     = errors.New("'server.host' field is empty")
	ErrEmptyPort     = errors.New("'server.port' field is empty")
	ErrEmptyEmail    = errors.New("'server.email' field is empty")
	ErrEmptyPassword = errors.New("'server.password' field is empty")
	// Mail
	ErrEmptyMail    = errors.New("'mail' field is empty")
	ErrEmptyName    = errors.New("'name' field is empty")
	ErrEmptySubject = errors.New("'subject' field is empty")
	ErrEmptyMessage = errors.New("'message' field is empty")
	ErrEmptyRecps   = errors.New("'recipients' field is empty")
)

// Enabled checks if the email feature is enabled in the config file.
// It returns true if the email feature is enabled, false otherwise.
func Enabled() bool {
	return E != (s.YAMLEmail{})
}

// Initializes the email feature by setting the default timeout for the email server.
// If the 'server.timeout' field is not set in the config file, it will be set to 10 seconds.
// Additionally, if the proxy is enabled, it sets the gomail.NetDialTimeout function to tunnel through the proxy.
// If any error occurs during the initialization, it is stored in the config.CError variable.
func init() {
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
			dialer, err := p.Dial("")
			if err != nil {
				config.CError = err
			}
			gomail.NetDialTimeout = func(network, addr string, timeout time.Duration) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		}
	}
}

// _recipientParsing returns a string that represents the recipients of the email.
// It checks if the recipients field is an array of strings or just a string.
// If it is an array, it will combine all the elements into a single string with a comma as a separator.
// If it is just a string, it will return the string as it is.
func _recipientParsing() string {
	switch Mail.Recipients.(type) {
	case []interface{}:
		{
			recipients := Mail.Recipients.([]interface{})
			combine := []string{}
			for _, recipient := range recipients {
				combine = append(combine, recipient.(string))
			}
			return strings.Join(combine, ",")
		}
	case string:
		{
			return Mail.Recipients.(string)
		}
	}
	return ""
}

// Init_Headers is a function that will initialize the headers for the email message.
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

// Set_Password will replace "<password>" in the email message with the given password and set it as the body of the email
// If the Email feature is not enabled, Set_Password will do nothing
func Set_Password(password string) {
	// If Email feature is not enabled, doing nothing
	if Enabled() {
		Mail.Message = strings.Replace(Mail.Message, "<password>", password, -1)
		Message.SetBody("text/plain", Mail.Message)
	}
}

// Send will send the email with the given password
// If the Email feature is not enabled, Send will do nothing
// It will return an error if there is any issue with sending the email
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

// Ping tests if the email server is reachable.
// It opens a connection to the server with the given configuration
// and then closes it.
//
// If the connection can be established, Ping returns nil.
// If the connection cannot be established, Ping returns an error.
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
		fmt.Print("\033[32m[~] Testing email configuration... \033[0m")
		// All the lines below are testing if the email configuration is correct or not.
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
		// Ping is connecting to the server to see that it can establish a connection with the mail server.
		err = Ping()
		if err != nil {
			return err
		}
		fmt.Print(" \033[32mdone.\033[0m\r\n")
	}
	return nil
}
