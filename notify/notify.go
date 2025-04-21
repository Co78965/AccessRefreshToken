package notify

import (
	testusersmanage "AccessRefreshToken/test_users_manage"
	"crypto/tls"
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

type IUser interface {
	GetUserEmail(guid string) (string, error)
}

type UsersManager struct {
	usersManager IUser
}

type Notifyer struct {
}

func newUsersManager(usersManager IUser) *UsersManager {
	return &UsersManager{usersManager: usersManager}
}

func (n *Notifyer) Notify(guid string, ip string) error {
	testUsersManager := &testusersmanage.TestUsersManager{}
	usersManager := newUsersManager(testUsersManager)

	email, err := usersManager.usersManager.GetUserEmail(guid)
	if err != nil {
		return err
	}

	title := "Угроза кражи аккаунта!"
	text := fmt.Sprintf("Внимание, с ip адреса: %s была попытка продления сессии!", ip)

	if err := sendEmail(email, text, title); err != nil {
		log.Printf("[ERROR] func: RefreshTokens --> SendEmail | error: %v\n", err)
	}
	return nil
}

func sendEmail(email, text string, title string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", "lyugaev2003@gmail.com")
	m.SetHeader("To", email)

	m.SetHeader("Subject", title)

	m.SetBody("text/plain", text)

	d := gomail.NewDialer("smtp.gmail.com", 587, "lyugaev2003@gmail.com", "gmzr jccs lxkd wbba")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
