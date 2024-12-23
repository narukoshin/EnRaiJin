package email

import (
	"crypto/tls"
	"custom-bruteforce/pkg/config"
	s "custom-bruteforce/pkg/structs"
	p "custom-bruteforce/pkg/proxy/v1"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

var (
	Server	   s.YAMLEmailServer	= config.YAMLConfig.E.Server
	Mail 	   s.YAMLEmailMail		= config.YAMLConfig.E.Mail
	E 		   s.YAMLEmail			= config.YAMLConfig.E

	ErrEmptyName	= errors.New("name field is empty")
	ErrEmptySubject = errors.New("subject field is empty")
	ErrEmptyMessage = errors.New("message field is empty")
	ErrEmptyrecps   = errors.New("no recipients added")
)
	
func Enabled() bool {
	return E != (s.YAMLEmail{})
}

func Set_Password(password string) {
	Mail.Message = strings.Replace(Mail.Message, "<password>", password, -1)
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}
  
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}
  
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
			case "Username:":
				return []byte(a.username), nil
			case "Password:":
				return []byte(a.password), nil
			default:
				return nil, errors.New("smtp: unknown from server")
		}
	}
	return nil, nil
}


// first connection test will be at the beginning of the tool
// second one will be after the the tool will find the password... just to be sure that the mail is still working properly and we can sent an email
func Test_Connection() (net.Conn, *smtp.Client, error) {
	// Nothing will happen in the email is not set
	if Enabled() {
		// before testing connection, let's test the config that is provided in the config file.
		if len(Mail.Name) == 0 {
			return nil, nil, ErrEmptyName
		}
		if len(Mail.Subject) == 0 {
			return nil, nil, ErrEmptySubject
		}
		if len(Mail.Message) == 0 {
			return nil, nil, ErrEmptyMessage
		}
		if Mail.Recipients == nil {
			return nil, nil, ErrEmptyrecps
		}
		// testing if we can connect to the server
		if Server.Timeout == 0 {
			Server.Timeout = 30
		}
		var conn net.Conn
		if p.IsProxy() {
			dialer, err := p.Dialer(time.Duration(Server.Timeout))
			if err != nil {
				return nil, nil, err
			}
			conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%s", Server.Host, Server.Port))
			if err != nil {
				return nil, nil, err
			}
		} else {
			var err error
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", Server.Host, Server.Port), time.Second * time.Duration(Server.Timeout))
			if err != nil {
				return nil, nil, err
			}

		}

		client, err := smtp.NewClient(conn, Server.Host)
		if err != nil {
			return nil, nil, err
		}

		// testing the connection between the client and server
		err = client.Hello(Server.Host)
		if err != nil {
			return nil, nil, err
		}
		// testing if we can connect to the mail
		tls := &tls.Config {
			InsecureSkipVerify: true,
		}
		err = client.StartTLS(tls)
		if err != nil {
			return nil, nil, err
		}
		auth := LoginAuth(Server.Email, Server.Password)
		err   = client.Auth(auth)
		if err != nil {
			return nil, nil, err
		}
		return conn, client, nil
	}
	return nil, nil, nil
}

func Send_Message(password string) error {
	// checking if the email is active
	if Enabled(){
		var recs string
		Set_Password(password)
		// another test
		conn, client, err := Test_Connection()
		if err != nil {
		  return err
		}
		err = client.Mail(Server.Email)
		if err != nil {
			return err
		}
		switch Mail.Recipients.(type) {
			case []interface{}:{
				recipients := Mail.Recipients.([]interface{})
				combine := []string{}
				for _, recp := range recipients {
					err = client.Rcpt(recp.(string))
					if err != nil {
						return err
					}
					combine = append(combine, recp.(string))
				}
				recs = strings.Join(combine, ",")
			}
			case string:{
				err = client.Rcpt(Mail.Recipients.(string))
				if err != nil {
					return err
				}
				recs = Mail.Recipients.(string)
			}
		}
		data, err := client.Data()
		if err != nil {
			return err
		}
		data.Write([]byte(fmt.Sprintf("To: %s\r\nFrom: \"%s\" <%s>\r\nSubject: %s\r\n\r\n%s", recs, Mail.Name, Server.Email, Mail.Subject, Mail.Message)))
		data.Close()
		err = client.Quit()
		if err != nil {
			return err
		}
		client.Close()
		conn.Close()
	}
	return nil
}