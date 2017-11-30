package correiosapi

import (
	"fmt"
	strut "github.com/pintobikez/brazilian-correios-service/api/structures"
	cnf "github.com/pintobikez/brazilian-correios-service/config/structures"
	rever "github.com/pintobikez/brazilian-correios-service/correiosapi/soapreverse"
	track "github.com/pintobikez/brazilian-correios-service/correiosapi/soaptracking"
	repo "github.com/pintobikez/brazilian-correios-service/repository"
	"regexp"
	"strconv"
	"time"
)

var (
	//ServiceTypeMap map of service types string to int codes
	ServiceTypeMap = map[string]string{"PAC": "04677", "SEDEX": "41076", "ESEDEX": "81043"}
	//RequestTypeMap map of service types string to correios codes
	RequestTypeMap = map[string]string{"POSTAGE": "AP", "COLECT": "LR"}
	//ColectTypeMap map of service types string to correios codes
	ColectTypeMap = map[string]string{"PAC": "LR", "SEDEX": "LS", "ESEDEX": "LV"}
	//FollowMap map of service types string to correios codes
	FollowMap = map[string]string{"POSTAGE": "A", "COLECT": "C"}
	//LanguageMap map of service types string to int codes
	LanguageMap = map[string]string{"BR": "101", "EN": "102"}
	//TrackingTypeMap map of service types string to correios codes
	TrackingTypeMap = map[string]string{"ALL": "T", "LAST": "U"}
)

const (
	//FollowPending correios code
	FollowPending = "55"
	//FollowCanceled correios code
	FollowCanceled = "9"
	//FollowExpired correios code
	FollowExpired = "57"
	//FollowOK correios code
	FollowOK = "0"
)

//Handler struct
type Handler struct {
	Repo repo.Definition
	Conf *cnf.CorreiosConfig
}

//TrackObjects Checks in Correios WebServicethe Tracking status of the given objects
func (h *Handler) TrackObjects(o *strut.Tracking) (*strut.TrackingResponse, error) {
	if len(o.Objects) > 0 {
		client := track.NewRastroWS(h.Conf.URLTracking, true)

		response, err := client.BuscaEventosLista(&track.BuscaEventosLista{User: h.Conf.UserTracking, Password: h.Conf.PwTracking, Type: "L", Result: TrackingTypeMap[o.TrackingType], Language: LanguageMap[o.Language], Objects: o.Objects})
		if err != nil {
			return nil, err
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

		return res, nil
	}

	return nil, nil
}

//FollowReverseLogistic Checks in Correios WebService which requests have changed
func (h *Handler) FollowReverseLogistic(requestType string) []*strut.RequestResponse {
	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.URLReverse, true, &oauth)

	// Get the Requests that had updates today in Correios
	currentTime := time.Now().Local()
	response, err := client.AcompanharPedidoPorData(&rever.AcompanharPedidoPorData{CodAdministrativo: h.Conf.CodAdministrativo, TipoSolicitacao: requestType, Data: currentTime.Format("02/01/2006")})
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	if response.AcompanharPedidoPorData.Coderro != "0" {

		length := len(response.AcompanharPedidoPorData.Coleta)
		toRet := make([]*strut.RequestResponse, 0, length)

		for _, col := range response.AcompanharPedidoPorData.Coleta {
			request, err := h.Repo.GetRequestByPostageCode(strconv.Itoa(col.Numeropedido))

			if err == nil && request.RequestID > 0 {
				switch col.Objeto[0].Ultimostatus {
				case FollowCanceled:
					af, err2 := h.Repo.UpdateRequestStatus(request, strut.StatusCanceled, col.Objeto[0].Descricaostatus)
					if err2 != nil && af > 0 {
						toRet = append(toRet, &strut.RequestResponse{request.RequestID, request.PostageCode, request.TrackingCode, strut.StatusCanceled, request.Callback})
					}
					break
				case FollowExpired:
					af, err2 := h.Repo.UpdateRequestStatus(request, strut.StatusExpired, col.Objeto[0].Descricaostatus)
					if err2 != nil && af > 0 {
						toRet = append(toRet, &strut.RequestResponse{request.RequestID, request.PostageCode, request.TrackingCode, strut.StatusExpired, request.Callback})
					}
					break
				case FollowOK:
					err2 := h.Repo.UpdateRequestTracking(request, col.Objeto[0].Numeroetiqueta)
					if err2 != nil {
						toRet = append(toRet, &strut.RequestResponse{request.RequestID, request.PostageCode, request.TrackingCode, strut.StatusUsed, request.Callback})
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

//DoReverseLogistic Performs in Correios WebService a request for a Reverse Postage
func (h *Handler) DoReverseLogistic(o *strut.Request) {

	//Update the status of the items to Processing
	if _, err := h.Repo.UpdateRequestStatus(o, strut.StatusProcessing, ""); err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.URLReverse, true, &oauth)

	dest := buildDestinatario(o)
	// cc, _ := strconv.Atoi(code)
	cole := buildColetaReversa(o)
	coletas := make([]*rever.ColetaReversa, 1)
	coletas[0] = cole

	req := rever.SolicitarPostagemReversa(rever.SolicitarPostagemReversa{CodAdministrativo: h.Conf.CodAdministrativo, Codigoservico: ServiceTypeMap[o.RequestService], Cartao: h.Conf.CartaoPostagem,
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
}

//CancelReverseLogistic Performs in Correios WebService a request for a Reverse Postage
func (h *Handler) CancelReverseLogistic(o *strut.Request) {

	//Update the status of the items to Processing
	if _, err := h.Repo.UpdateRequestStatus(o, strut.StatusProcessing, ""); err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	//Init SOAP Client
	oauth := rever.BasicAuth{Login: h.Conf.UserReverse, Password: h.Conf.PwReverse}
	client := rever.NewLogisticaReversaWS(h.Conf.URLReverse, true, &oauth)

	// Get the PostalRange from Correios
	response, err := client.CancelarPedido(&rever.CancelarPedido{CodAdministrativo: h.Conf.CodAdministrativo, NumeroPedido: o.PostageCode, Tipo: FollowMap[o.RequestService]})
	if err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}

	if response.CancelarPedido.Coderro != "" {
		h.saveErrorMessage(o, response.CancelarPedido.Coderro+" - "+response.CancelarPedido.Msgerro)
		return
	}

	//Update the status of the items to Processing
	if _, err := h.Repo.UpdateRequestStatus(o, strut.StatusCanceled, ""); err != nil {
		h.saveErrorMessage(o, err.Error())
		return
	}
}

//saveErrorMessage Error message
func (h *Handler) saveErrorMessage(o *strut.Request, message string) {
	o.Retries++
	if _, err2 := h.Repo.UpdateRequestStatus(o, strut.StatusError, message); err2 != nil {
		fmt.Println(err2.Error())
	}
	return
}

//buildColetaReversa Builds the ColectaReversa struct
func buildColetaReversa(o *strut.Request) *rever.ColetaReversa {
	r := buildRemetente(o)

	cc := rever.Coleta{Tipo: FollowMap[o.RequestType], Remetente: r}
	var ob []*rever.Objeto

	for _, i := range o.Items {
		obj := new(rever.Objeto)
		obj.ID = o.SlipNumber
		obj.Desc = i.ProductName
		obj.Item = "1"
		ob = append(ob, obj)
	}

	c := rever.ColetaReversa{Coleta: &cc, Objcol: ob}
	if FollowMap[o.RequestType] == "C" {
		c.Ag = o.ColectDate
	}

	return &c
}

//buildDestinatario Builds the Destinatorio struct
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

//buildRemetente Builds the Remetente struct
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
