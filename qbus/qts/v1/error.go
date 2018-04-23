package qts

type QtsErrCode uint

const (
	QtsErrorBadRequest    QtsErrCode = iota + 40000
	QtsErrorInternalError QtsErrCode = iota + 50000
)
