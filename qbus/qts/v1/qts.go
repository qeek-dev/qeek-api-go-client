package qts

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
)

// any is purely semantic
type any interface{}

// Pointer is purely semantic
type pointer interface{}

func isPointer(value any) bool {
	if reflect.ValueOf(value).Kind() != reflect.Ptr {
		return false
	}
	return true
}

type Response struct {
	Code      int
	ErrorCode int
	ErrorMsg  string
}

// Nas login
type NasLoginResponse struct {
	Response
	Result NasLoginResult
}

type NasLoginResult struct {
	AuthPassed int
	IsAdmin    int
	AuthSid    string
}

// verify sid
type VerifySidResponse struct {
	Response
}

// Nas me
type NasMeResponse struct {
	Response
	Result NasMeResult
}

type NasMeResult struct {
	User string
}

// Nas Users
type NasUsersResponse struct {
	Response
	Result []NasUserResult
}

// Nas User
type NasUserResponse struct {
	Response
	Result NasUserResult
}

type NasUserResult struct {
	Email  string
	Enable int
	Group  []string
	Lang   string
	Name   string
	Avatar string
}

// Nas Account avatar
type NasUserAvatarResponse struct {
	Response
	Result NasUserAvatarResult
}

type NasUserAvatarResult struct {
	Path string
}

type Service struct {
	sid           string
	qbusNameSpace string
}

func logError(err error) error {
	log.Println(err)
	return err
}

func exec(out pointer, cmd ...interface{}) error {
	if !isPointer(out) {
		return errors.New(fmt.Sprintf("Value '%s' is not a pointer", out))
	}

	o, err := sh.Command("qbus", cmd...).Output()
	if err != nil {
		return logError(errors.Wrap(err, "qbus command exec fail"))
	}

	err = json.Unmarshal(o, &out)
	if err != nil {
		return logError(errors.Wrap(err, "qbus json unmarshal fail"))
	}

	return nil
}

func (l *Service) GetSid() string {
	return l.sid
}

// Nas Me call
type NasMeCall struct {
	s        *Service
	username string
}

func (l *Service) Me() *NasMeCall {
	return &NasMeCall{l, ""}
}

func (l *NasMeCall) Do() (r NasUserResult, err error) {
	var a NasMeResponse
	err = exec(&a, "get", fmt.Sprintf("%s/qts/user/me", l.s.qbusNameSpace), fmt.Sprint(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = errors.Wrap(err, "Get Nas User Me fail")
	} else if a.Code != 200 {
		err = logError(errors.Wrap(errors.New(fmt.Sprintf("[%d] %s", a.ErrorCode, a.ErrorMsg)), "Get Nas User Me fail"))
	}

	if err != nil {
		return
	}

	return l.s.User().UserName(a.Result.User).Do()
}

// Nas user call
type NasUserCall struct {
	s        *Service
	username string
}

func (l *Service) User() *NasUserCall {
	return &NasUserCall{l, ""}
}

func (l *NasUserCall) UserName(username string) *NasUserCall {
	l.username = username
	return l
}

func (l *NasUserCall) Do() (r NasUserResult, err error) {
	var a NasUserResponse
	err = exec(&a, "get", fmt.Sprintf("%s/qts/user/%s", l.s.qbusNameSpace, l.username), fmt.Sprint(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = errors.Wrap(err, "Get Nas User fail")
	} else {
		if a.Code != 200 {
			err = logError(errors.Wrap(errors.New(fmt.Sprintf("[%d] %s", a.ErrorCode, a.ErrorMsg)), "Get Nas User fail"))
		} else {
			r = a.Result
		}
	}
	return
}

// Nas users call
type NasUsersCall struct {
	s *Service
}

func (l *Service) Users() *NasUsersCall {
	return &NasUsersCall{l}
}

func (l *NasUsersCall) Do() (r []NasUserResult, err error) {
	var a NasUsersResponse
	err = exec(&a, "get", fmt.Sprintf("%s/qts/users", l.s.qbusNameSpace), fmt.Sprint(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = errors.Wrap(err, "Get Nas Users fail")
	} else {
		if a.Code != 200 {
			err = logError(errors.Wrap(errors.New(fmt.Sprintf("[%d] %s", a.ErrorCode, a.ErrorMsg)), "Get Nas Users fail"))
		} else {
			r = a.Result
		}
	}
	return
}

// verify sid call
type VerifySidCall struct {
	s *Service
}

func (l *Service) Verify() *VerifySidCall {
	return &VerifySidCall{l}
}

func (l *VerifySidCall) Sid(sid string) *VerifySidCall {
	l.s.sid = sid
	return l
}

func (l *VerifySidCall) Do() (err error) {
	var out VerifySidResponse
	err = exec(&out, "get", fmt.Sprintf("%s/qts/verify_sid", l.s.qbusNameSpace), fmt.Sprint(`{"sid":"%s"}`, l.s.sid))
	if err != nil || out.Code != 200 {
		if err != nil {
			err = logError(errors.Wrap(err, "Verify Sid fail"))
		} else {
			err = logError(errors.Wrap(errors.New(fmt.Sprintf("[%d] %s", out.ErrorCode, out.ErrorMsg)), "Verify Sid fail"))
		}
		// clear sid
		l.s.sid = ""
	}
	return
}

// login call
type LoginCall struct {
	s        *Service
	username string
	password string
}

func (l *Service) Login() *LoginCall {
	return &LoginCall{l, "", ""}
}

func (l *LoginCall) UserName(username string) *LoginCall {
	l.username = username
	return l
}

func (l *LoginCall) Password(password string) *LoginCall {
	l.password = password
	return l
}

func (l *LoginCall) Do() (err error) {
	var out NasLoginResponse
	err = exec(&out, "get", fmt.Sprintf("%s/qts/account_login", l.s.qbusNameSpace), fmt.Sprint(`{"user":"%s","pwd":"%s"}`, l.username, l.password))
	if err == nil && out.Code == 200 {
		l.s.sid = out.Result.AuthSid
	} else {
		if err != nil {
			err = logError(errors.Wrap(err, "Nas login fail"))
		} else {
			err = logError(errors.Wrap(errors.New(fmt.Sprintf("[%d] %s", out.ErrorCode, out.ErrorMsg)), "Nas login fail"))
		}
	}
	return
}

func NewClient(qbusNameSpace string) *Service {
	return &Service{qbusNameSpace, ""}
}
