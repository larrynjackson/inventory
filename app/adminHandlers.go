package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github/lnj/inventory/model"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	gomail "gopkg.in/mail.v2"
)

var (
	ClientActionCreateUser     = "CREATE_USER"
	ClientActionLogin          = "LOGIN"
	ClientActionInventory      = "INVENTORY"
	ClientActionSendCode       = "SEND_CODE"
	ClientActionChangePassword = "CHANGE_PASSWORD"

	ErrorPasswordLength   = "password must be 6-15 characters"
	ErrorUserExists       = "user is already registered"
	ErrorUserNotExist     = "user is not registered"
	ErrorUserFieldMissing = "user field is missing"
	ErrorApplicationError = "application error"
	ErrorPassCodeMatch    = "code sent does not match"
	ErrorUserName         = "user name format error"
	ErrorEmailSend        = "send email code failed"
	ErrorLoginFail        = "invalid user name/password"
	ErrorLoginNeeded      = "You are not logged in"
	ErrorUserLoggedin     = "user is already logged in"
	ErrorPWPCmismatch     = "Passwords do not match"
)

const (
	LocalFreeFlagURL = "http://localhost:8000/api/v1/getfile/flag.jpeg"
)

func sendEmailCode(passCode, to, subject, message string) bool {

	EmailSender := os.Getenv("SMTP_SENDER")
	SmtpHost := os.Getenv("SMTP_HOST")
	EmailPwd := os.Getenv("SMTP_PWD")

	haveEmailEnv := true
	if EmailSender == "" || SmtpHost == "" || EmailPwd == "" {
		haveEmailEnv = false
	}
	if haveEmailEnv {
		email := gomail.NewMessage()

		email.SetHeader("From", EmailSender)
		email.SetHeader("To", to)
		email.SetHeader("Subject", subject)

		msg := []byte("To: " + to + "\r\n" + "Subject:" + subject + "\r\n" + "\r\n" + message + passCode + "\r\n")

		email.SetBody("text/plain", string(msg))

		dialer := gomail.NewDialer(SmtpHost, 587, EmailSender, EmailPwd)

		if err := dialer.DialAndSend(email); err != nil {
			fmt.Println("send email error:", err)
			fmt.Println("msg:", string(msg))
			return false
		}
		fmt.Println("Email sent successfully! new passCode:", passCode)
	} else {
		fmt.Println("In order for email to function, set these env variables:")
		fmt.Println("SMTP_SENDER, SMTP_PWD, SMTP_HOST")
		fmt.Println("Email NOT sent, local only passCode is:", passCode)
	}
	return true
}

// isEmailValid checks if the email provided is valid by regex.
func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegex.MatchString(e)
}

func setCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{}
	cookie.Name = "LInventoryCooker"
	cookie.Value = token
	cookie.Expires = time.Now().Add(8 * time.Hour)
	cookie.Secure = true
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteNoneMode
	cookie.Path = "/"
	http.SetCookie(w, &cookie)
}

func getCookie(r *http.Request) string {
	var returnStr string
	var token string
	for _, cookie := range r.Cookies() {
		if cookie.Name == "LInventoryCooker" {
			returnStr = returnStr + cookie.Name + ":" + cookie.Value + "\n"
			token = cookie.Value
		}
	}
	return token
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER
	var token = getCookie(r)

	mapFrame, ok := s.controlMap.GetMap(token)
	if !ok {
		return ErrorApplicationError, false
	}
	userModel.Password = r.PostFormValue("password")
	if !s.userRegistered(mapFrame.context.userId) {
		return ErrorUserNotExist, false
	}
	sum := sha256.Sum256([]byte(mapFrame.context.userId + ":" + userModel.Password))
	inputPwdHash := hex.EncodeToString(sum[:])
	var inputPassword string = string(inputPwdHash)

	userPasswordMap := s.Storage.SelectUser(mapFrame.context.userId)
	var storedPwd = userPasswordMap[mapFrame.context.userId]

	if inputPassword != storedPwd {
		return "Incorrect password.", false
	}
	dKeyList := s.Storage.SelectUserdKeyList(mapFrame.context.tKey)
	err := os.Chdir("./images")
	if err != nil {
		panic(err)
	}
	for _, key := range dKeyList {
		fmt.Println("remove:", key)
		err = os.RemoveAll("./" + key)
		if err != nil {
			panic(err)
		}
	}
	err = os.Chdir("..")
	if err != nil {
		panic(err)
	}
	s.Storage.DeleteItems(mapFrame.context.tKey)
	s.Storage.DeleteUser(mapFrame.context.userId)
	var newMapFrame MapFrame
	newMapFrame.token = token
	newMapFrame.createTime = time.Now()
	newMapFrame.timeToLive = time.Duration(30 * time.Minute)

	newMapFrame.context.userId = mapFrame.context.userId
	newMapFrame.context.machineId = mapFrame.context.machineId
	newMapFrame.context.password = "Deleted"
	newMapFrame.context.passCode = "Deleted"
	newMapFrame.context.currAction = "Deleted"
	newMapFrame.context.nextAction = "Deleted"

	s.controlMap.PushMap(token, newMapFrame)
	setCookie(w, token)
	return "Delete Complete", true
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER

	userModel.UserId = r.PostFormValue("userId")
	userModel.Password = r.PostFormValue("password")
	userModel.PassCode = r.PostFormValue("passCode")
	userModel.MachineId = r.PostFormValue("machineId")

	if !isEmailValid(userModel.UserId) {
		return ErrorUserName, false
	}
	if len(userModel.Password) < 6 || len(userModel.Password) > 15 {
		return ErrorPasswordLength, false
	}
	if userModel.Password != userModel.PassCode {
		return ErrorPWPCmismatch, false
	}
	if len(userModel.UserId) == 0 {
		return ErrorUserFieldMissing, false
	}
	if len(userModel.MachineId) == 0 {
		return ErrorApplicationError, false
	}
	userPasswordMap := s.Storage.SelectUser(userModel.UserId)
	password := userPasswordMap[userModel.UserId]
	if len(password) > 0 {
		return ErrorUserExists, false
	}

	sum := sha256.Sum256([]byte(userModel.UserId + ":" + userModel.Password))
	pwd := hex.EncodeToString(sum[:])

	strArr := []rune("0123456789")
	var passCode string
	for j := 0; j < 6; j++ {
		passCode += string(strArr[rand.Intn(10)])
	}
	fmt.Println("************* emailed code:", passCode)
	var subject string = "Registration Request"
	var message string = "To complete the registration, enter this code:"

	if !sendEmailCode(passCode, userModel.UserId, subject, message) {
		return ErrorEmailSend, false
	}

	// create tracker
	var token = GenerateToken(userModel.MachineId)
	var newMapFrame MapFrame

	newMapFrame.token = token
	newMapFrame.createTime = time.Now()
	newMapFrame.timeToLive = time.Duration(30 * time.Minute)

	newMapFrame.context.userId = userModel.UserId
	newMapFrame.context.tKey = ""
	newMapFrame.context.machineId = userModel.MachineId
	newMapFrame.context.password = pwd
	newMapFrame.context.passCode = passCode
	newMapFrame.context.currAction = ClientActionCreateUser
	newMapFrame.context.nextAction = ClientActionSendCode

	s.controlMap.PushMap(token, newMapFrame)

	setCookie(w, token)
	return "A security code was emailed.", true
}

func (s *Server) handleMatchCode(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER

	userModel.PassCode = r.PostFormValue("passCode")
	userModel.MachineId = r.PostFormValue("machineId")

	var token = getCookie(r)

	mapFrame, ok := s.controlMap.GetMap(token)
	if !ok {
		return ErrorApplicationError, false
	}

	fmt.Println("userPC:", userModel.PassCode)
	fmt.Println("storePC:", mapFrame.context.passCode)

	if userModel.PassCode == mapFrame.context.passCode {
		var tKey = GenerateRandomKey(4)
		s.Storage.InsertUser(mapFrame.context.userId, tKey, mapFrame.context.password)

		setCookie(w, GenerateToken(userModel.MachineId))
		return "User Added Successfully", true
	} else {
		return ErrorPassCodeMatch, false
	}
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER

	var token = getCookie(r)

	userModel.MachineId = r.PostFormValue("machineId")
	userModel.Password = r.PostFormValue("password")
	userModel.PassCode = r.PostFormValue("newPwdOne")
	var pwdTwo string = r.PostFormValue("newPwdTwo")

	if userModel.PassCode != pwdTwo {
		return "New Passwords mismatch.", false
	}
	if len(userModel.PassCode) < 6 || len(userModel.PassCode) > 15 {
		return ErrorPasswordLength, false
	}

	mapFrame, ok := s.controlMap.GetMap(token)
	if !ok {
		return "Bad Password or not logged in.", false
	}

	userModel.UserId = mapFrame.context.userId
	var tKey = mapFrame.context.tKey

	sum := sha256.Sum256([]byte(mapFrame.context.userId + ":" + userModel.Password))
	pwd := hex.EncodeToString(sum[:])
	userPasswordMap := s.Storage.SelectUser(mapFrame.context.userId)
	password := userPasswordMap[mapFrame.context.userId]
	if password != string(pwd) {
		return "Invalid Password", false
	}

	sum = sha256.Sum256([]byte(mapFrame.context.userId + ":" + userModel.PassCode))
	pwd = hex.EncodeToString(sum[:])
	s.Storage.UpdateUser(mapFrame.context.userId, string(pwd))

	token = GenerateToken(userModel.MachineId)
	var newMapFrame MapFrame

	newMapFrame.token = token
	newMapFrame.createTime = time.Now()
	newMapFrame.timeToLive = time.Duration(30 * time.Minute)

	newMapFrame.context.userId = userModel.UserId
	newMapFrame.context.tKey = tKey
	newMapFrame.context.machineId = userModel.MachineId
	newMapFrame.context.password = pwd
	newMapFrame.context.currAction = ClientActionChangePassword
	newMapFrame.context.nextAction = ClientActionInventory

	s.controlMap.PushMap(token, newMapFrame)
	setCookie(w, token)

	return "Password Updated.", true
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER

	userModel.UserId = r.PostFormValue("userId")
	userModel.MachineId = r.PostFormValue("machineId")

	if !s.userRegistered(userModel.UserId) {
		return "User ID is not registered.", false
	}

	var token = getCookie(r)

	strArr := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	var passCode string
	for j := 0; j < 10; j++ {
		passCode += string(strArr[rand.Intn(36)])
	}
	sum := sha256.Sum256([]byte(userModel.UserId + ":" + passCode))
	emailPwdHash := hex.EncodeToString(sum[:])

	var subject string = "Password Reset"

	var message string = `A password reset for this login has been initiated. If
	you did not request this change, you can ignore this email (no changes have been
	made to your Inventory accout) or call and complain to the boss.
	If you did request this change, then use the temporary password provided to login
	then change your password using the Change Password option. The temporary password
	will remain active for 10 minutes. Temporary Password:`

	var newMapFrame MapFrame

	newMapFrame.context.userId = userModel.UserId
	newMapFrame.context.machineId = userModel.MachineId
	newMapFrame.context.password = string(emailPwdHash)
	newMapFrame.context.currAction = "Temp-Password-Reset"

	mapFrame, ok := s.controlMap.GetMap(token)
	if ok {
		if mapFrame.context.currAction == "Temp-Password-Reset" {
			if mapFrame.context.nextAction == "waitOne" {
				newMapFrame.context.nextAction = "waitTwo"
			} else if mapFrame.context.nextAction == "waitTwo" {
				newMapFrame.context.nextAction = "waitThree"
			} else {
				newMapFrame.context.nextAction = "waitOne"
			}

		} else {
			newMapFrame.context.nextAction = "waitOne"
		}
	} else {
		newMapFrame.context.nextAction = "waitOne"
	}
	if newMapFrame.context.nextAction == "waitOne" || newMapFrame.context.nextAction == "waitTwo" {

		s.Storage.UpdateUser(newMapFrame.context.userId, string(emailPwdHash))

		if !sendEmailCode(passCode, userModel.UserId, subject, message) {
			return ErrorEmailSend, false
		}
	} else {
		return "Too Many Attempts!", false
	}
	token = GenerateToken(userModel.MachineId)

	newMapFrame.token = token
	newMapFrame.createTime = time.Now()
	newMapFrame.timeToLive = time.Duration(30 * time.Minute)
	newMapFrame.context.tKey = mapFrame.context.tKey

	s.controlMap.PushMap(token, newMapFrame)
	setCookie(w, token)
	return "Emailed new password.", true
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) (string, bool) {
	var userModel model.USER

	userModel.UserId = r.PostFormValue("userId")
	userModel.Password = r.PostFormValue("password")
	userModel.MachineId = r.PostFormValue("machineId")

	var ok bool = true
	if !s.userRegistered(userModel.UserId) {
		return ErrorUserNotExist, !ok
	}
	if len(userModel.Password) == 0 {
		return ErrorPasswordLength, !ok
	}
	sum := sha256.Sum256([]byte(userModel.UserId + ":" + userModel.Password))
	inputPwdHash := hex.EncodeToString(sum[:])
	var inputPassword string = string(inputPwdHash)

	userPasswordMap := s.Storage.SelectUser(userModel.UserId)
	var storedPwd = userPasswordMap[userModel.UserId]
	var tKey = userPasswordMap["tKey"]

	if inputPassword != storedPwd {
		return ErrorLoginFail, !ok
	}

	var token = GenerateToken(userModel.MachineId)
	var newMapFrame MapFrame

	newMapFrame.token = token
	newMapFrame.createTime = time.Now()
	newMapFrame.timeToLive = time.Duration(30 * time.Minute)

	newMapFrame.context.userId = userModel.UserId
	newMapFrame.context.tKey = tKey
	newMapFrame.context.machineId = userModel.MachineId
	newMapFrame.context.password = inputPassword
	newMapFrame.context.currAction = ClientActionLogin
	newMapFrame.context.nextAction = ClientActionInventory

	s.controlMap.PushMap(token, newMapFrame)

	setCookie(w, token)
	result := strings.Split(userModel.UserId, "@")
	return "Welcome " + result[0], ok
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) string {
	var token = getCookie(r)

	mapFrame, ok := s.controlMap.GetMap(token)
	if !ok {
		return "So Long"
	}
	if mapFrame.context.nextAction == ClientActionInventory {
		setCookie(w, GenerateToken(mapFrame.context.machineId))
		return "Log Out Complete - Bye"

	} else {
		return "Log Out Complete"
	}
}

func (s *Server) userRegistered(userId string) bool {
	userPasswordMap := s.Storage.SelectUser(userId)
	user := userPasswordMap[userId]
	return len(user) > 0
}
