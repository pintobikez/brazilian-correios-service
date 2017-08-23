package repository

import s "github.com/pintobikez/correios-service/api/structures"

type RepositoryDefinition interface {
	ConnectDB(stringConn string) error
	DisconnectDB()
	InsertRequest(object *s.Request) error
	FindRequestById(requestId int64) (bool, error)
	GetRequestBy(req *s.Search) ([]*s.Request, error)
	GetRequestById(requestId int) (*s.Request, error)
	GetRequestByPostageCode(code string) (*s.Request, error)
	UpdateRequest(object *s.Request) error
	UpdateRequestStatus(object *s.Request, status string, message string) (int64, error)
	UpdateRequestPostage(object *s.Request, code string) error
	UpdateRequestTracking(o *s.Request, code string) error
}
