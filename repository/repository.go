package repository

import s "github.com/pintobikez/brazilian-correios-service/api/structures"

type Definition interface {
	ConnectDB(stringConn string) error
	DisconnectDB()
	InsertRequest(object *s.Request) error
	FindRequestByID(requestID int64) (bool, error)
	GetRequestBy(req *s.Search) ([]*s.Request, error)
	GetRequestByID(requestID int) (*s.Request, error)
	GetRequestByPostageCode(code string) (*s.Request, error)
	UpdateRequest(object *s.Request) error
	UpdateRequestStatus(object *s.Request, status string, message string) (int64, error)
	UpdateRequestPostage(object *s.Request, code string) error
	UpdateRequestTracking(o *s.Request, code string) error
}
