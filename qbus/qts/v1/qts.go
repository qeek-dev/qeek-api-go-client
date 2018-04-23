package qts

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
)

type QtsErr struct {
	Code     QtsErrCode
	QbusCode int
	QbusErr  string
	Err      error
}

func (q *QtsErr) Error() string {
	if q.QbusErr != "" {
		return fmt.Sprintf("%d: %s, Qbus[%d]: %s", q.Code, q.Err.Error(), q.QbusCode, q.QbusErr)
	} else {
		return fmt.Sprintf("%d: %s", q.Code, q.Err.Error())
	}
}

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
	s             *sh.Session
	sid           string
	qbusNameSpace string
}

func logError(err error) error {
	log.Println(err)
	return err
}

func (s *Service) exec(out pointer, cmd ...interface{}) error {
	if !isPointer(out) {
		return errors.New(fmt.Sprintf("Value '%s' is not a pointer", out))
	}
	o, err := s.s.Command("qbus", cmd...).Output()
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
	var out NasMeResponse
	err = l.s.exec(&out, "get", fmt.Sprintf("%s/qts/user/me", l.s.qbusNameSpace), fmt.Sprintf(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = logError(&QtsErr{Code: QtsErrorInternalError, Err: err})
	} else if out.Code != 200 {
		err = logError(&QtsErr{Code: QtsErrorBadRequest, Err: err, QbusCode: out.ErrorCode, QbusErr: out.ErrorMsg})
	}

	if err != nil {
		return
	}

	return l.s.User().UserName(out.Result.User).Do()
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
	var out NasUserResponse
	err = l.s.exec(&out, "get", fmt.Sprintf("%s/qts/user/%s", l.s.qbusNameSpace, l.username), fmt.Sprintf(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = logError(&QtsErr{Code: QtsErrorInternalError, Err: err})
	} else {
		if out.Code != 200 {
			err = logError(&QtsErr{Code: QtsErrorBadRequest, Err: err, QbusCode: out.ErrorCode, QbusErr: out.ErrorMsg})
		} else {
			r = out.Result
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
	var out NasUsersResponse
	err = l.s.exec(&out, "get", fmt.Sprintf("%s/qts/users", l.s.qbusNameSpace), fmt.Sprintf(`{"sid":"%s"}`, l.s.sid))
	if err != nil {
		err = logError(&QtsErr{Code: QtsErrorInternalError, Err: err})
	} else {
		if out.Code != 200 {
			err = logError(&QtsErr{Code: QtsErrorBadRequest, Err: err, QbusCode: out.ErrorCode, QbusErr: out.ErrorMsg})
		} else {
			r = out.Result
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
	err = l.s.exec(&out, "get", fmt.Sprintf("%s/qts/verify_sid", l.s.qbusNameSpace), fmt.Sprintf(`{"sid":"%s"}`, l.s.sid))
	if err != nil || out.Code != 200 {
		if err != nil {
			err = logError(&QtsErr{Code: QtsErrorInternalError, Err: err})
		} else {
			err = logError(&QtsErr{Code: QtsErrorBadRequest, Err: err, QbusCode: out.ErrorCode, QbusErr: out.ErrorMsg})
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
	err = l.s.exec(&out, "get", fmt.Sprintf("%s/qts/account_login", l.s.qbusNameSpace), fmt.Sprintf(`{"user":"%s","pwd":"%s"}`, l.username, l.password))
	if err == nil && out.Code == 200 {
		l.s.sid = out.Result.AuthSid
	} else {
		if err != nil {
			err = logError(&QtsErr{Code: QtsErrorInternalError, Err: err})
		} else {
			err = logError(&QtsErr{Code: QtsErrorBadRequest, Err: err, QbusCode: out.ErrorCode, QbusErr: out.ErrorMsg})
		}
	}
	return
}

func NewClient(qbusNameSpace string, debugMode bool) *Service {
	s := &Service{}
	s.qbusNameSpace = qbusNameSpace
	s.s = sh.NewSession()
	s.s.ShowCMD = debugMode
	return s
}
