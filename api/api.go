package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	strut "github.com/pintobikez/correios-service/api/structures"
	cnf "github.com/pintobikez/correios-service/config/structures"
	hand "github.com/pintobikez/correios-service/correiosapi"
	repo "github.com/pintobikez/correios-service/repository"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var OPERATOR_VALUES = map[string]bool{"LIKE": true, "=": true, ">=": true, "<=": true, "<>": true, "!=": true, "IN": true, "NOT IN": true}

const DATE_REGEX = "(^(((0[1-9]|1[0-9]|2[0-8])[/](0[1-9]|1[012]))|((29|30|31)[/](0[13578]|1[02]))|((29|30)[/](0[4,6,9]|11)))[/](19|[2-9][0-9])$)|(^29[/]02[/](19|[2-9][0-9])(00|04|08|12|16|20|24|28|32|36|40|44|48|52|56|60|64|68|72|76|80|84|88|92|96)$)"

type Api struct {
	Repo repo.RepositoryDefinition
	Conf *cnf.CorreiosConfig
	Hand *hand.Handler
}

func New(r repo.RepositoryDefinition, c *cnf.CorreiosConfig) *Api {
	return &Api{Repo: r, Conf: c, Hand: &hand.Handler{Repo: r, Conf: c}}
}

// Handler to Get Tracking information
func (a *Api) GetTracking() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Tracking)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidateTrackingJson(o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// Track the objects
		// If the number of objects is Lower then 5 we run it as a normal function
		// If not we run it a go routine
		if len(o.Objects) <= 5 {
			err, ret := a.Hand.TrackObjects(o)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
			}

			return c.JSON(http.StatusOK, ret)

		}

		go func() {
			if _, ret := a.Hand.TrackObjects(o); ret != nil {
				doCallbackRequest(ret, o.Callback)
			}
		}()

		return c.NoContent(http.StatusOK)
	}
}

// Handler to GET Reverse information of a request
func (a *Api) GetReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		// if requestId doesn't exist
		if isset := c.Param("requestId"); isset == "" {
			return c.JSON(http.StatusBadRequest, "requestId not set")
		}

		requestId, err := strconv.Atoi(c.Param("requestId"))
		// if requestId isn't an int
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// try to find the request
		res, err := a.Repo.GetRequestById(requestId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		if res.RequestId == 0 {
			return c.JSON(http.StatusNotFound, &ErrResponse{ErrContent{http.StatusNotFound, fmt.Sprintf("Request with ID: %d not found", requestId)}})
		}

		return c.JSON(http.StatusOK, res)
	}
}

// Handler to GET Reverse information for N Requests
func (a *Api) GetReversesBy() echo.HandlerFunc {
	return func(c echo.Context) error {

		s := new(strut.Search)
		// if is an invalid json format
		if err := c.Bind(&s); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidateSearchJson(s); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
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
func (a *Api) PostReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Request)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidatePutJson(o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// insert the request into the db
		if err := a.Repo.InsertRequest(o); err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		// Create GO routine to perform Correios request
		go a.Hand.DoReverseLogistic(o)

		return c.JSON(http.StatusOK, struct {
			RequestId int64 `json:"request_id"`
		}{o.RequestId})
	}
}

// Handler to PUT a Correios Reverse request
func (a *Api) PutReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		o := new(strut.Request)
		// if is an invalid json format
		if err := c.Bind(&o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// check if the json is valid
		if err := a.ValidatePutJson(o); err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// try to find the request
		found, err := a.Repo.FindRequestById(o.RequestId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}
		if !found {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, "request not found"}})
		}

		// rollback the status to PENDING
		if o.Status == strut.STATUS_ERROR || o.Status == strut.STATUS_EXPIRED {
			o.Status = strut.STATUS_PENDING
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
func (a *Api) DeleteReverse() echo.HandlerFunc {
	return func(c echo.Context) error {

		// if requestId doesn't exist
		if isset := c.Param("requestId"); isset == "" {
			return c.JSON(http.StatusBadRequest, "requestId not set")
		}

		requestId, err := strconv.Atoi(c.Param("requestId"))
		// if requestId isn't an int
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrResponse{ErrContent{http.StatusBadRequest, err.Error()}})
		}

		// try to find the request
		res, err := a.Repo.GetRequestById(requestId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, err.Error()}})
		}

		if res.RequestId == 0 {
			return c.JSON(http.StatusInternalServerError, &ErrResponse{ErrContent{http.StatusInternalServerError, fmt.Sprintf("Request with ID: %d not found", requestId)}})
		}

		// Create GO routine to perform Correios request
		go a.Hand.DoReverseLogistic(res)

		return c.JSON(http.StatusOK, res)
	}
}

// Validates the consistency of the Tracking struct
func (a *Api) ValidateTrackingJson(s *strut.Tracking) error {

	if s.Callback == "" {
		return fmt.Errorf("Callback is empty")
	}
	if s.TrackingType == "" {
		return fmt.Errorf("Tracking Type is empty")
	} else {
		s.TrackingType = strings.ToUpper(s.TrackingType)
		if _, ok := hand.TRACKING_TYPE[s.TrackingType]; !ok {
			values := ""
			for _, value := range hand.TRACKING_TYPE {
				values += value + " "
			}
			return fmt.Errorf("Tracking Type valid values are: %s", values)
		}
	}
	if s.Language == "" {
		return fmt.Errorf("Language is empty")
	} else {
		s.Language = strings.ToUpper(s.Language)
		if _, ok := hand.LANGUAGE[s.Language]; !ok {
			values := ""
			for _, value := range hand.LANGUAGE {
				values += value + " "
			}
			return fmt.Errorf("Language valid values are: %s", values)
		}
	}
	if len(s.Objects) == 0 {
		return fmt.Errorf("Objects is empty")
	}
	return nil
}

// Validates the consistency of the Request struct
func (a *Api) ValidatePutJson(s *strut.Request) error {

	if s.RequestType == "" {
		return fmt.Errorf("Request Type is empty")
	} else {
		s.RequestType = strings.ToUpper(s.RequestType)
		if _, ok := hand.REQUEST_TYPE[s.RequestType]; !ok {
			values := ""
			for _, value := range hand.REQUEST_TYPE {
				values += value + " "
			}
			return fmt.Errorf("Request Type valid values are: %s", values)
		}
	}

	if s.RequestService == "" {
		return fmt.Errorf("Request Service is empty")
	} else {
		s.RequestService = strings.ToUpper(s.RequestService)
		if _, ok := hand.SERVICE_TYPE[s.RequestService]; !ok {
			values := ""
			for _, value := range hand.SERVICE_TYPE {
				values += value + " "
			}
			return fmt.Errorf("Service Type valid values are: %s", values)
		}
	}

	if s.ColectDate != "" {
		if re := regexp.MustCompile(DATE_REGEX); !re.MatchString(s.ColectDate) {
			return fmt.Errorf("Colect Date must be in format dd/mm/yyyy")
		}
	}

	if s.OriginNome == "" {
		return fmt.Errorf("Origin Nome is empty")
	}
	if s.OriginLogradouro == "" {
		return fmt.Errorf("Origin Logradouro is empty")
	}
	if s.OriginNumero <= 0 {
		return fmt.Errorf("Origin Numero is empty")
	}
	if s.OriginCep == "" {
		return fmt.Errorf("Origin Cep is empty")
	}
	if s.OriginBairro == "" {
		return fmt.Errorf("Origin Bairro is empty")
	}
	if s.OriginCidade == "" {
		return fmt.Errorf("Origin Cidade is empty")
	}
	if s.OriginUf == "" {
		return fmt.Errorf("Origin Uf is empty")
	}
	if s.OriginEmail == "" {
		return fmt.Errorf("Origin Email is empty")
	}
	if s.Callback == "" {
		return fmt.Errorf("Callback is empty")
	}
	if s.OriginDdd != "" {
		if _, err := strconv.Atoi(s.OriginDdd); err != nil {
			return fmt.Errorf("Origin Ddd contains non-numeric characters")
		}
	}
	if s.OriginTelefone != "" {
		if _, err := strconv.Atoi(s.OriginTelefone); err != nil {
			return fmt.Errorf("Origin Telefone contains non-numeric characters")
		}
	}
	if s.SlipNumber == "" {
		return fmt.Errorf("Slip number is empty")
	}
	if s.DestinationNome == "" {
		return fmt.Errorf("Destination Nome is empty")
	}
	if s.DestinationLogradouro == "" {
		return fmt.Errorf("Destination Logradouro is empty")
	}
	if s.DestinationNumero <= 0 {
		return fmt.Errorf("Destination Numero is empty")
	}
	if s.DestinationCep == "" {
		return fmt.Errorf("Destination Cep is empty")
	}
	if s.DestinationBairro == "" {
		return fmt.Errorf("Destination Bairro is empty")
	}
	if s.DestinationCidade == "" {
		return fmt.Errorf("Destination Cidade is empty")
	}
	if s.DestinationUf == "" {
		return fmt.Errorf("Destination Uf is empty")
	}
	if s.DestinationEmail == "" {
		return fmt.Errorf("Origin Email is empty")
	}

	for _, i := range s.Items {
		if i.Item == "" {
			return fmt.Errorf("Item is empty")
		}
		if i.ProductName == "" {
			return fmt.Errorf("Product name is empty")
		}
	}

	return nil
}

// Validates the consistency of the Search struct
func (a *Api) ValidateSearchJson(s *strut.Search) error {
	s.OrderType = strings.ToUpper(s.OrderType)

	if s.From < 0 || s.Offset < 0 {
		return fmt.Errorf("From and Offset is lower than 0")
	}
	if s.OrderType != "" && s.OrderType != "ASC" && s.OrderType != "DESC" {
		return fmt.Errorf("Order type must be ASC or DESC")
	}
	for _, e := range s.Where {
		e.Operator = strings.ToUpper(e.Operator)
		if e.Operator != "" && OPERATOR_VALUES[e.Operator] {
			return fmt.Errorf("Operator as an invalid value, valid: like, =, >=, <=, <>, IN, NOT IN")
		}
	}
	return nil
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
	useTls := re.MatchString(url)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: useTls},
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
