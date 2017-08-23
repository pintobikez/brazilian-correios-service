package correiosapi

import (
	"fmt"
	strut "github.com/pintobikez/correios-service/api/structures"
	cnf "github.com/pintobikez/correios-service/config/structures"
	rever "github.com/pintobikez/correios-service/correiosapi/soapreverse"
	track "github.com/pintobikez/correios-service/correiosapi/soaptracking"
	repo "github.com/pintobikez/correios-service/repository"
	"regexp"
	"strconv"
	"time"
)

var (
	SERVICE_TYPE  = map[string]string{"PAC": "04677", "SEDEX": "41076", "ESEDEX": "81043"}
	REQUEST_TYPE  = map[string]string{"POSTAGE": "AP", "COLECT": "LR"}
	COLECT_TYPE   = map[string]string{"PAC": "LR", "SEDEX": "LS", "ESEDEX": "LV"}
	FOLLOW        = map[string]string{"POSTAGE": "A", "COLECT": "C"}
	LANGUAGE      = map[string]string{"BR": "101", "EN": "102"}
	TRACKING_TYPE = map[string]string{"ALL": "T", "LAST": "U"}
)

const (
	SOAP_URL        = "SOAP_URL"
	FOLLOW_PENDING  = "55"
	FOLLOW_CANCELED = "9"
	FOLLOW_EXPIRED  = "57"
	FOLLOW_OK       = "0"
)

type Handler struct {
	Repo repo.RepositoryDefinition
	Conf *cnf.CorreiosConfig
}

// Checks in Correios WebServicethe Tracking status of the given objects
func (h *Handler) TrackObjects(o *strut.Tracking) (error, *strut.TrackingResponse) {
	if len(o.Objects) > 0 {
		client := track.NewRastroWS(h.Conf.UrlTracking, true)

		response, err := client.BuscaEventosLista(&track.BuscaEventosLista{User: h.Conf.UserTracking, Password: h.Conf.PwTracking, Type: "L", Result: TRACKING_TYPE[o.TrackingType], Language: LANGUAGE[o.Language], Objects: o.Objects})
		if err != nil {
			return err, nil
		}

		res := new(strut.TrackingResponse)
		res.Items = make([]*strut.TrackingHeader, 0, response.Result.Quantity)

		for i, el := range response.Result.Objects {
			res.Items = append(res.Items, new(strut.TrackingHeader))
			res.Items[i].Object = el.TrackingCode
			if el.Error == "" {
				res.Items[i].Name = el.Name
				res.Items[i].Category = el.Category
				res.Items[i].Events = make([]*strut.TrackingEvents, 0, 1)

				//Get all events and append them
				for _, ev := range el.Events {
					dt := ev.Date + " " + ev.Hour
					cte := ev.Local + ", " + ev.Code + ", " + ev.City + "(" + ev.FiscalUnit + ")"
					res.Items[i].Events = append(res.Items[i].Events, &strut.TrackingEvents{Type: ev.Type, StatusCode: ev.StatusCode, DateTime: dt, Description: ev.Description, Details: ev.Detail, CTECorreios: cte})
				}

			} else {
				res.Items[i].Error = el.Error
			}
		}

		return nil, res
	}

	return nil, nil
}

// Checks in Correios WebService which requests have changed
func (h *Handler) FollowReverseLogistic(requestType string) []*strut.RequestResponse {
	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.UrlReverse, true, &oauth)

	// Get the Requests that had updates today in Correios
	current_time := time.Now().Local()
	response, err := client.AcompanharPedidoPorData(&rever.AcompanharPedidoPorData{CodAdministrativo: h.Conf.CodAdministrativo, TipoSolicitacao: requestType, Data: current_time.Format("02/01/2006")})
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	if response.AcompanharPedidoPorData.Coderro != "0" {

		length := len(response.AcompanharPedidoPorData.Coleta)
		toRet := make([]*strut.RequestResponse, 0, length)

		for _, col := range response.AcompanharPedidoPorData.Coleta {
			request, err := h.Repo.GetRequestByPostageCode(strconv.Itoa(col.Numeropedido))

			if err == nil && request.RequestId > 0 {
				switch col.Objeto[0].Ultimostatus {
				case FOLLOW_CANCELED:
					af, err2 := h.Repo.UpdateRequestStatus(request, strut.STATUS_CANCELED, col.Objeto[0].Descricaostatus)
					if err2 != nil && af > 0 {
						toRet = append(toRet, &strut.RequestResponse{request.RequestId, request.PostageCode, request.TrackingCode, strut.STATUS_CANCELED, request.Callback})
					}
					break
				case FOLLOW_EXPIRED:
					af, err2 := h.Repo.UpdateRequestStatus(request, strut.STATUS_EXPIRED, col.Objeto[0].Descricaostatus)
					if err2 != nil && af > 0 {
						toRet = append(toRet, &strut.RequestResponse{request.RequestId, request.PostageCode, request.TrackingCode, strut.STATUS_EXPIRED, request.Callback})
					}
					break
				case FOLLOW_OK:
					err2 := h.Repo.UpdateRequestTracking(request, col.Objeto[0].Numeroetiqueta)
					if err2 != nil {
						toRet = append(toRet, &strut.RequestResponse{request.RequestId, request.PostageCode, request.TrackingCode, strut.STATUS_USED, request.Callback})
					}
					break
				default:
					break
				}
			}
		}

		return toRet
	}
	return nil
}

// Performs in Correios WebService a request for a Reverse Postage
func (h *Handler) DoReverseLogistic(o *strut.Request) {

	//Update the status of the items to Processing
	_, err := h.Repo.UpdateRequestStatus(o, strut.STATUS_PROCESSING, "")
	if err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.UrlReverse, true, &oauth)

	// Get the PostalRange from Correios
	response, err := client.SolicitarRange(&rever.SolicitarRange{CodAdministrativo: h.Conf.CodAdministrativo, Tipo: REQUEST_TYPE[o.RequestType], Quantidade: "1"})
	if err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	// If no ERROR
	if response.SolicitarRange.Coderro == "0" || response.SolicitarRange.Coderro == "247" {
		code := strconv.Itoa(response.SolicitarRange.Faixafinal)
		//Calculate the verifying digit
		dig, err := calcularDigitoVerificadorStatic(code)
		if err != nil {
			h.saveErrorMessage(o, err.Error())
			return
		}
		code += dig
		dest := buildDestinatario(o)
		cc, _ := strconv.Atoi(code)
		cole := buildColetaReversa(o, cc)
		coletas := make([]*rever.ColetaReversa, 1)
		coletas[0] = cole

		req := rever.SolicitarPostagemReversa(rever.SolicitarPostagemReversa{CodAdministrativo: h.Conf.CodAdministrativo, Codigoservico: SERVICE_TYPE[o.RequestService], Cartao: h.Conf.CartaoPostagem,
			Destinatario: dest, Coletassolicitadas: coletas})
		resp, err := client.SolicitarPostagemReversa(&req)

		if err != nil {
			h.saveErrorMessage(o, err.Error())
			return
		}

		if resp.SolicitarPostagemReversa.Coderro != "00" {
			// Error in the request
			h.saveErrorMessage(o, resp.SolicitarPostagemReversa.Msgerro)
			return
		}

		r := resp.SolicitarPostagemReversa.Resultadosolicitacao[0]
		// The request has been made before the Numerocoleta is inside the error message
		if r.Codigoerro == 121 {
			re := regexp.MustCompile("[0-9]{9}")
			r.Numerocoleta = re.FindAllString(r.Descricaoerro, 1)[0]
		}
		// Error in the result of the request
		if r.Codigoerro != 0 && r.Codigoerro != 121 {
			h.saveErrorMessage(o, "Error coleta: "+strconv.Itoa(r.Codigoerro)+" - "+r.Descricaoerro)
			return
		}
		//Update the DB with the Numerocoleta
		if err2 := h.Repo.UpdateRequestPostage(o, r.Numerocoleta); err != nil {
			fmt.Println(err2.Error())
		}
		return

	} else {
		h.saveErrorMessage(o, "Error range: "+response.SolicitarRange.Coderro+" - "+response.SolicitarRange.Msgerro)
		return
	}
}

// Performs in Correios WebService a request for a Reverse Postage
func (h *Handler) CancelReverseLogistic(o *strut.Request) {

	//Update the status of the items to Processing
	_, err := h.Repo.UpdateRequestStatus(o, strut.STATUS_PROCESSING, "")
	if err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.UrlReverse, true, &oauth)

	// Get the PostalRange from Correios
	response, err := client.CancelarPedido(&rever.CancelarPedido{CodAdministrativo: h.Conf.CodAdministrativo, NumeroPedido: o.PostageCode, Tipo: FOLLOW[o.RequestService]})
	if err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	if response.CancelarPedido.Coderro != "" {
		h.saveErrorMessage(o, response.CancelarPedido.Coderro+" - "+response.CancelarPedido.Msgerro)
		return
	}
}

func (h *Handler) saveErrorMessage(o *strut.Request, message string) {
	o.Retries += 1
	if _, err2 := h.Repo.UpdateRequestStatus(o, strut.STATUS_ERROR, message); err2 != nil {
		fmt.Println(err2.Error())
	}
	return
}

//Builds the ColectaReversa struct
func buildColetaReversa(o *strut.Request, numero int) *rever.ColetaReversa {
	r := buildRemetente(o)

	cc := rever.Coleta{Tipo: FOLLOW[o.RequestType], Remetente: r}
	ob := make([]*rever.Objeto, 0)

	for _, i := range o.Items {
		obj := new(rever.Objeto)
		obj.Id = o.SlipNumber
		obj.Desc = i.ProductName
		obj.Item = "1"
		ob = append(ob, obj)
	}

	c := rever.ColetaReversa{Numero: numero, Coleta: &cc, Objcol: ob}
	if FOLLOW[o.RequestType] == "C" {
		c.Ag = o.ColectDate
	}

	return &c
}

//Builds the Destinatorio struct
func buildDestinatario(o *strut.Request) *rever.Pessoa {
	p := rever.Pessoa{
		Nome:        o.DestinationNome,
		Logradouro:  o.DestinationLogradouro,
		Numero:      string(o.DestinationNumero),
		Complemento: o.DestinationComplemento,
		Bairro:      o.DestinationBairro,
		Referencia:  o.DestinationReferencia,
		Cidade:      o.DestinationCidade,
		Uf:          o.DestinationUf,
		Cep:         o.DestinationCep,
		Ddd:         "",
		Telefone:    "",
		Email:       o.DestinationEmail,
	}

	return &p
}

//Builds the Remetente struct
func buildRemetente(o *strut.Request) *rever.Remetente {
	p := rever.Pessoa{
		Nome:        o.OriginNome,
		Logradouro:  o.OriginLogradouro,
		Numero:      string(o.OriginNumero),
		Complemento: o.OriginComplemento,
		Bairro:      o.OriginBairro,
		Referencia:  o.OriginReferencia,
		Cidade:      o.OriginCidade,
		Uf:          o.OriginUf,
		Cep:         o.OriginCep,
		Ddd:         o.OriginDdd,
		Telefone:    o.OriginTelefone,
		Email:       o.OriginEmail,
	}

	r := rever.Remetente{Pessoa: &p}

	return &r
}

// Calculates the verifying digit of a postage code
func calcularDigitoVerificadorStatic(eticket string) (string, error) {
	multipliers := [8]int{8, 6, 4, 2, 3, 5, 9, 7}
	sum := 0
	dv := 0

	if len(eticket) != 8 {
		return strconv.Itoa(dv), fmt.Errorf("e-ticket need have 8 digits")
	}

	for k, v := range multipliers {
		a := string(eticket[k])
		b, _ := strconv.Atoi(a)
		sum += b * v
	}

	remainder := sum % 11

	switch remainder {
	case 0:
		dv = 5
	case 1:
		dv = 0
	default:
		dv = 11 - remainder
	}

	return strconv.Itoa(dv), nil
}
