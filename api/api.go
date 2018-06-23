package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	strut "github.com/pintobikez/brazilian-correios-service/api/structures"
	cnf "github.com/pintobikez/brazilian-correios-service/config/structures"
	hand "github.com/pintobikez/brazilian-correios-service/correiosapi"
	repo "github.com/pintobikez/brazilian-correios-service/repository"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const DateRegex = "(^(((0[1-9]|1[0-9]|2[0-8])[/](0[1-9]|1[012]))|((29|30|31)[/](0[13578]|1[02]))|((29|30)[/](0[4,6,9]|11)))[/](19|[2-9][0-9])$)|(^29[/]02[/](19|[2-9][0-9])(00|04|08|12|16|20|24|28|32|36|40|44|48|52|56|60|64|68|72|76|80|84|88|92|96)$)"

var (
	OperatorValues   = map[string]bool{"LIKE": true, "=": true, ">=": true, "<=": true, "<>": true, "!=": true, "IN": true, "NOT IN": true}
	RequestNotFound  = "Request with ID: %d not found"
	ErrorNotSet      = "%s not set"
	ErrorIsEmpty     = "is empty"
	ErrorValidValues = "valid values are: %s"
)

type API struct {
	Repo repo.Definition
	Conf *cnf.CorreiosConfig
	Hand *hand.Handler
}

func New(r repo.Definition, c *cnf.CorreiosConfig) *API {
	return &API{Repo: r, Conf: c, Hand: &hand.Handler{Repo: r, Conf: c}}
}

// Handler to Get Tracking information
func (a *API) GetTracking() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Tracking)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidateTrackingJSON(o); len(err) > 0 {
			return c.JSON(http.StatusBadRequest, buildErrorResponse(err))
		}

		// Track the objects
		// If the number of objects is Lower then 5 we run it as a normal function
		// If not we run it as a go routine
		if len(o.Objects) <= 5 {
			ret, err := a.Hand.TrackObjects(o)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
			}

			return c.JSON(http.StatusOK, ret)
		}

		// run it as a go routine and reply to the callback url
		go func() {
			if ret, _ := a.Hand.TrackObjects(o); ret != nil {
				doCallbackRequest(ret, o.Callback)
			}
		}()

		return c.NoContent(http.StatusOK)
	}
}

// Handler to GET Reverse information of a request
func (a *API) GetReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		// if requestId doesn't exist
		if isset := c.Param("requestId"); isset == "" {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, fmt.Sprintf(ErrorNotSet, "requestId")}})
		}

		requestID, err := strconv.Atoi(c.Param("requestId"))
		// if requestId isn't an int
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// try to find the request
		res, err := a.Repo.GetRequestByID(requestID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		if res.RequestID == 0 {
			return c.JSON(http.StatusNotFound, &ErrResponse{ErrContent{http.StatusNotFound, fmt.Sprintf(RequestNotFound, requestID)}})
		}

		return c.JSON(http.StatusOK, res)
	}
}

// Handler to GET Reverse information for N Requests
func (a *API) GetReversesBy() echo.HandlerFunc {
	return func(c echo.Context) error {

		s := new(strut.Search)
		// if is an invalid json format
		if err := c.Bind(&s); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidateSearchJSON(s); len(err) > 0 {
			return c.JSON(http.StatusBadRequest, buildErrorResponse(err))
		}

		// try to find the requests
		res, err := a.Repo.GetRequestBy(s)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		return c.JSON(http.StatusOK, res)
	}
}

// Handler to POST a Correios Reverse request
func (a *API) PostReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Request)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidatePutJSON(o); len(err) > 0 {
			return c.JSON(http.StatusBadRequest, buildErrorResponse(err))
		}

		// insert the request into the db
		if err := a.Repo.InsertRequest(o); err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		// Create GO routine to perform Correios request
		go a.Hand.DoReverseLogistic(o)

		return c.JSON(http.StatusOK, struct {
			RequestID int64 `json:"request_id"`
		}{o.RequestID})
	}
}

// Handler to PUT a Correios Reverse request
func (a *API) PutReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Request)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidatePutJSON(o); len(err) > 0 {
			return c.JSON(http.StatusBadRequest, buildErrorResponse(err))
		}

		// try to find the request
		found, err := a.Repo.FindRequestByID(o.RequestID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}
		if !found {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, fmt.Sprintf(RequestNotFound, o.RequestID)}})
		}

		// rollback the status to PENDING
		if o.Status == strut.StatusError || o.Status == strut.StatusExpired {
			o.Status = strut.StatusPending
			o.ErrorMessage = ""
		}
		// update the request
		err = a.Repo.UpdateRequest(o)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		return c.JSON(http.StatusOK, o)
	}
}

// Handler to DELETE a Correios Reverse request
func (a *API) DeleteReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		// if requestId doesn't exist
		if isset := c.Param("requestId"); isset == "" {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, fmt.Sprintf(ErrorNotSet, "requestId")}})
		}

		requestID, err := strconv.Atoi(c.Param("requestId"))
		// if requestId isn't an int
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// try to find the request
		res, err := a.Repo.GetRequestByID(requestID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		if res.RequestID == 0 {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, fmt.Sprintf(RequestNotFound, requestID)}})
		}

		// Create GO routine to perform Correios request
		go a.Hand.CancelReverseLogistic(res)

		return c.JSON(http.StatusOK, res)
	}
}

// Validates the consistency of the Tracking struct
func (a *API) ValidateTrackingJSON(s *strut.Tracking) map[string]string {

	ret := make(map[string]string)

	if s.Callback == "" {
		ret["callback"] = ErrorIsEmpty
	}
	if s.TrackingType == "" {
		ret["tracking_type"] = ErrorIsEmpty
	} else {
		s.TrackingType = strings.ToUpper(s.TrackingType)
		if _, ok := hand.TrackingTypeMap[s.TrackingType]; !ok {
			values := ""
			for _, value := range hand.TrackingTypeMap {
				values += value + " "
			}
			ret["tracking_type"] = fmt.Sprintf(ErrorValidValues, values)
		}
	}
	if s.Language == "" {
		ret["language"] = ErrorIsEmpty
	} else {
		s.Language = strings.ToUpper(s.Language)
		if _, ok := hand.LanguageMap[s.Language]; !ok {
			values := ""
			for _, value := range hand.LanguageMap {
				values += value + " "
			}
			ret["language"] = fmt.Sprintf(ErrorValidValues, values)
		}
	}
	if len(s.Objects) == 0 {
		ret["objects"] = ErrorIsEmpty
	}

	return ret
}

// Validates the consistency of the Request struct
func (a *API) ValidatePutJSON(s *strut.Request) map[string]string {

	ret := make(map[string]string)

	if s.RequestType == "" {
		ret["request_type"] = ErrorIsEmpty
	} else {
		s.RequestType = strings.ToUpper(s.RequestType)
		if _, ok := hand.RequestTypeMap[s.RequestType]; !ok {
			values := ""
			for _, value := range hand.RequestTypeMap {
				values += value + " "
			}
			ret["request_type"] = fmt.Sprintf(ErrorValidValues, values)
		}
	}

	if s.RequestService == "" {
		ret["request_service"] = ErrorIsEmpty
	} else {
		s.RequestService = strings.ToUpper(s.RequestService)
		if _, ok := hand.ServiceTypeMap[s.RequestService]; !ok {
			values := ""
			for _, value := range hand.ServiceTypeMap {
				values += value + " "
			}
			ret["request_service"] = fmt.Sprintf(ErrorValidValues, values)
		}
	}

	if s.ColectDate != "" {
		if re := regexp.MustCompile(DateRegex); !re.MatchString(s.ColectDate) {
			ret["colect_date"] = "Colect Date must be in format dd/mm/yyyy"
		}
	}

	if s.OriginNome == "" {
		ret["origin_nome"] = ErrorIsEmpty
	}
	if s.OriginLogradouro == "" {
		ret["origin_logradouro"] = ErrorIsEmpty
	}
	if s.OriginNumero <= 0 {
		ret["origin_numero"] = ErrorIsEmpty
	}
	if s.OriginCep == "" {
		ret["origin_cep"] = ErrorIsEmpty
	}
	if s.OriginBairro == "" {
		ret["origin_bairro"] = ErrorIsEmpty
	}
	if s.OriginCidade == "" {
		ret["origin_cidade"] = ErrorIsEmpty
	}
	if s.OriginUf == "" {
		ret["origin_uf"] = ErrorIsEmpty
	}
	if s.OriginEmail == "" {
		ret["origin_email"] = ErrorIsEmpty
	}
	if s.Callback == "" {
		ret["callback"] = ErrorIsEmpty
	}
	if s.OriginDdd != "" {
		if _, err := strconv.Atoi(s.OriginDdd); err != nil {
			ret["origin_ddd"] = fmt.Sprintf(ErrorValidValues, "numeric characters")
		}
	}
	if s.OriginTelefone != "" {
		if _, err := strconv.Atoi(s.OriginTelefone); err != nil {
			ret["origin_telefone"] = fmt.Sprintf(ErrorValidValues, "numeric characters")
		}
	}
	if s.SlipNumber == "" {
		ret["slip_number"] = ErrorIsEmpty
	}
	if s.DestinationNome == "" {
		ret["destination_nome"] = ErrorIsEmpty
	}
	if s.DestinationLogradouro == "" {
		ret["destination_logradouro"] = ErrorIsEmpty
	}
	if s.DestinationNumero <= 0 {
		ret["destination_numero"] = ErrorIsEmpty
	}
	if s.DestinationCep == "" {
		ret["destination_cep"] = ErrorIsEmpty
	}
	if s.DestinationBairro == "" {
		ret["destination_bairro"] = ErrorIsEmpty
	}
	if s.DestinationCidade == "" {
		ret["destination_cidade"] = ErrorIsEmpty
	}
	if s.DestinationUf == "" {
		ret["destination_uf"] = ErrorIsEmpty
	}
	if s.DestinationEmail == "" {
		ret["destination_email"] = ErrorIsEmpty
	}

	for _, i := range s.Items {
		if i.Item == "" {
			ret["items"] = ErrorIsEmpty
		}
		if i.ProductName == "" {
			ret["product_name"] = ErrorIsEmpty
		}
	}

	return ret
}

// Validates the consistency of the Search struct
func (a *API) ValidateSearchJSON(s *strut.Search) map[string]string {

	s.OrderType = strings.ToUpper(s.OrderType)
	ret := make(map[string]string)

	if s.From < 0 {
		ret["from"] = "lower than 0"
	}
	if s.Offset < 0 {
		ret["offset"] = "lower than 0"
	}
	if s.OrderType != "" && s.OrderType != "ASC" && s.OrderType != "DESC" {
		ret["order_type"] = "must be ASC or DESC"
	}

	for _, e := range s.Where {
		e.Operator = strings.ToUpper(e.Operator)
		e.JoinBy = strings.ToUpper(e.JoinBy)
		if e.Operator != "" && OperatorValues[e.Operator] {
			ret["operator"] = fmt.Sprintf(ErrorValidValues, "like, =, >=, <=, <>, IN, NOT IN")
		}
		if e.JoinBy != "" && e.JoinBy != "AND" && e.JoinBy != "OR" {
			ret["joinby"] = fmt.Sprintf(ErrorValidValues, "AND, OR")
		}
	}
	return ret
}

func buildErrorResponse(err map[string]string) *ErrResponseValidation {

	ret := &ErrResponseValidation{Type: "validation", Errors: make([]*ErrValidation, 0, len(err))}
	i := 0

	for k, v := range err {
		ret.Errors[i] = &ErrValidation{Field: k, Message: v}
		i++
	}

	return ret
}

// Performs an Http request
func doCallbackRequest(e *strut.TrackingResponse, url string) {
	buffer := new(bytes.Buffer)
	_ = json.NewEncoder(buffer).Encode(e)

	// Create the POST request to the callback
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		// log error
		fmt.Println(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Close = true

	// check if it is an https request
	re := regexp.MustCompile("^https://")
	useTlS := re.MatchString(url)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: useTlS},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		// log error
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()
}
