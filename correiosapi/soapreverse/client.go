package soapreverse

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// against "unused imports"
var _ time.Time
var _ xml.Name

//AcompanharPedidoPorData struct
type AcompanharPedidoPorData struct {
	XMLName xml.Name `xml:"ns1:acompanharPedidoPorData"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	TipoSolicitacao   string `xml:"tipoSolicitacao,omitempty"`
	Data              string `xml:"data,omitempty"`
}

//AcompanharPedidoPorDataResponse struct
type AcompanharPedidoPorDataResponse struct {
	XMLName xml.Name `xml:"acompanharPedidoPorDataResponse"`

	AcompanharPedidoPorData *RetornoAcompanhamento `xml:"acompanharPedidoPorData,omitempty"`
}

//RetornoAcompanhamento struct
type RetornoAcompanhamento struct {
	Codigoadministrativo string                `xml:"codigo_administrativo,omitempty"`
	Tiposolicitacao      string                `xml:"tipo_solicitacao,omitempty"`
	Coleta               []*ColetasSolicitadas `xml:"coleta,omitempty"`
	Data                 string                `xml:"data,omitempty"`
	Hora                 string                `xml:"hora,omitempty"`
	Coderro              string                `xml:"cod_erro,omitempty"`
	Msgerro              string                `xml:"msg_erro,omitempty"`
}

//ColetasSolicitadas struct
type ColetasSolicitadas struct {
	Numeropedido    int                `xml:"numero_pedido,omitempty"`
	Controlecliente string             `xml:"controle_cliente,omitempty"`
	Historico       []*HistoricoColeta `xml:"historico,omitempty"`
	Objeto          []*ObjetoPostal    `xml:"objeto,omitempty"`
}

//HistoricoColeta struct
type HistoricoColeta struct {
	Status          int    `xml:"status,omitempty"`
	Descricaostatus string `xml:"descricao_status,omitempty"`
	Dataatualizacao string `xml:"data_atualizacao,omitempty"`
	Horaatualizacao string `xml:"hora_atualizacao,omitempty"`
	Observacao      string `xml:"observacao,omitempty"`
}

//ObjetoPostal struct
type ObjetoPostal struct {
	Numeroetiqueta        string `xml:"numero_etiqueta,omitempty"`
	Controleobjetocliente string `xml:"controle_objeto_cliente,omitempty"`
	Ultimostatus          string `xml:"ultimo_status,omitempty"`
	Descricaostatus       string `xml:"descricao_status,omitempty"`
	Dataultimaatualizacao string `xml:"data_ultima_atualizacao,omitempty"`
	Horaultimaatualizacao string `xml:"hora_ultima_atualizacao,omitempty"`
	Pesocubico            string `xml:"peso_cubico,omitempty"`
	Pesoreal              string `xml:"peso_real,omitempty"`
	Valorpostagem         string `xml:"valor_postagem,omitempty"`
}

//SolicitarPostagemReversa struct
type SolicitarPostagemReversa struct {
	XMLName xml.Name `xml:"ns1:solicitarPostagemReversa"`

	CodAdministrativo  string           `xml:"codAdministrativo,omitempty"`
	Codigoservico      string           `xml:"codigo_servico,omitempty"`
	Cartao             string           `xml:"cartao,omitempty"`
	Destinatario       *Pessoa          `xml:"destinatario,omitempty"`
	Coletassolicitadas []*ColetaReversa `xml:"coletas_solicitadas,omitempty"`
}

//Pessoa struct
type Pessoa struct {
	Nome        string `xml:"nome,omitempty"`
	Logradouro  string `xml:"logradouro,omitempty"`
	Numero      string `xml:"numero,omitempty"`
	Complemento string `xml:"complemento,omitempty"`
	Bairro      string `xml:"bairro,omitempty"`
	Referencia  string `xml:"referencia,omitempty"`
	Cidade      string `xml:"cidade,omitempty"`
	Uf          string `xml:"uf,omitempty"`
	Cep         string `xml:"cep,omitempty"`
	Ddd         string `xml:"ddd,omitempty"`
	Telefone    string `xml:"telefone,omitempty"`
	Email       string `xml:"email,omitempty"`
}

//ColetaReversa struct
type ColetaReversa struct {
	*Coleta

	Numero           int       `xml:"numero,omitempty"`
	Ag               string    `xml:"ag,omitempty"`
	Cartao           string    `xml:"cartao,omitempty"`
	Servicoadicional string    `xml:"servico_adicional,omitempty"`
	Ar               int       `xml:"ar,omitempty"`
	Objcol           []*Objeto `xml:"obj_col,omitempty"`
}

//Coleta struct
type Coleta struct {
	Tipo           string     `xml:"tipo,omitempty"`
	Idcliente      string     `xml:"id_cliente,omitempty"`
	Valordeclarado string     `xml:"valor_declarado,omitempty"`
	Descricao      string     `xml:"descricao,omitempty"`
	Cklist         string     `xml:"cklist,omitempty"`
	Documento      []string   `xml:"documento,omitempty"`
	Remetente      *Remetente `xml:"remetente,omitempty"`
	Produto        []*Produto `xml:"produto,omitempty"`
}

//Objeto struct
type Objeto struct {
	Item    string `xml:"item,omitempty"`
	Desc    string `xml:"desc,omitempty"`
	Entrega string `xml:"entrega,omitempty"`
	Num     string `xml:"num,omitempty"`
	ID      string `xml:"id,omitempty"`
}

//Remetente struct
type Remetente struct {
	*Pessoa
	Identificacao string `xml:"identificacao,omitempty"`
	Dddcelular    string `xml:"ddd_celular,omitempty"`
	Celular       string `xml:"celular,omitempty"`
	Sms           string `xml:"sms,omitempty"`
}

//Produto struct
type Produto struct {
	Codigo string `xml:"codigo,omitempty"`
	Tipo   string `xml:"tipo,omitempty"`
	Qtd    string `xml:"qtd,omitempty"`
}

//SolicitarPostagemReversaResponse struct
type SolicitarPostagemReversaResponse struct {
	XMLName                  xml.Name         `xml:"solicitarPostagemReversaResponse"`
	SolicitarPostagemReversa *RetornoPostagem `xml:"solicitarPostagemReversa,omitempty"`
}

//RetornoPostagem struct
type RetornoPostagem struct {
	Statusprocessamento  string                  `xml:"status_processamento,omitempty"`
	Dataprocessamento    string                  `xml:"data_processamento,omitempty"`
	Horaprocessamento    string                  `xml:"hora_processamento,omitempty"`
	Coderro              string                  `xml:"cod_erro,omitempty"`
	Msgerro              string                  `xml:"msg_erro,omitempty"`
	Resultadosolicitacao []*ResultadoSolicitacao `xml:"resultado_solicitacao,omitempty"`
}

//ResultadoSolicitacao struct
type ResultadoSolicitacao struct {
	Tipo            string `xml:"tipo,omitempty"`
	Idcliente       string `xml:"id_cliente,omitempty"`
	Numerocoleta    string `xml:"numero_coleta,omitempty"`
	Numeroetiqueta  string `xml:"numero_etiqueta,omitempty"`
	Idobj           string `xml:"id_obj,omitempty"`
	Statusobjeto    string `xml:"status_objeto,omitempty"`
	Prazo           string `xml:"prazo,omitempty"`
	Datasolicitacao string `xml:"data_solicitacao,omitempty"`
	Horasolicitacao string `xml:"hora_solicitacao,omitempty"`
	Codigoerro      int    `xml:"codigo_erro,omitempty"`
	Descricaoerro   string `xml:"descricao_erro,omitempty"`
}

//CancelarPedido struct
type CancelarPedido struct {
	XMLName xml.Name `xml:"ns1:cancelarPedido"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	NumeroPedido      string `xml:"numeroPedido,omitempty"`
	Tipo              string `xml:"tipo,omitempty"`
}

//CancelarPedidoResponse struct
type CancelarPedidoResponse struct {
	XMLName        xml.Name             `xml:"cancelarPedidoResponse"`
	CancelarPedido *RetornoCancelamento `xml:"cancelarPedido,omitempty"`
}

//RetornoCancelamento struct
type RetornoCancelamento struct {
	Codigoadministrativo string              `xml:"codigo_administrativo,omitempty"`
	Objetopostal         *ObjetoSimplificado `xml:"objeto_postal,omitempty"`
	Data                 string              `xml:"data,omitempty"`
	Hora                 string              `xml:"hora,omitempty"`
	Coderro              string              `xml:"cod_erro,omitempty"`
	Msgerro              string              `xml:"msg_erro,omitempty"`
}

//ObjetoSimplificado struct
type ObjetoSimplificado struct {
	Numeropedido         int    `xml:"numero_pedido,omitempty"`
	Statuspedido         string `xml:"status_pedido,omitempty"`
	Datahoracancelamento string `xml:"datahora_cancelamento,omitempty"`
}

//LogisticaReversaWS struct
type LogisticaReversaWS struct {
	client *SOAPClient
}

//NewLogisticaReversaWS creates a new struct with the given information
func NewLogisticaReversaWS(url string, tls bool, auth *BasicAuth) *LogisticaReversaWS {
	if url == "" {
		url = ""
	}
	client := NewSOAPClient(url, tls, auth)

	return &LogisticaReversaWS{
		client: client,
	}
}

//AcompanharPedidoPorData checks the status of the reverese postage requests in Correios
// Error can be either of the following types:
//
//   - ComponenteException
func (service *LogisticaReversaWS) AcompanharPedidoPorData(request *AcompanharPedidoPorData) (*AcompanharPedidoPorDataResponse, error) {
	response := new(AcompanharPedidoPorDataResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

//SolicitarPostagemReversa Request a reverse postage to correios
// Error can be either of the following types:
//
//   - ComponenteException
func (service *LogisticaReversaWS) SolicitarPostagemReversa(request *SolicitarPostagemReversa) (*SolicitarPostagemReversaResponse, error) {
	response := new(SolicitarPostagemReversaResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

//CancelarPedido cancels a request in Correios
// Error can be either of the following types:
//
//   - ComponenteException
func (service *LogisticaReversaWS) CancelarPedido(request *CancelarPedido) (*CancelarPedidoResponse, error) {
	response := new(CancelarPedidoResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

var timeout = time.Duration(30 * time.Second)

//dialTimeout Sets the timeout of the connection
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

//SOAPEnvelopeResponse struct
type SOAPEnvelopeResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    SOAPBodyResponse
}

//SOAPBodyResponse struct
type SOAPBodyResponse struct {
	XMLName xml.Name           `xml:"Body"`
	Fault   *SOAPFaultResponse `xml:",omitempty"`
	Content interface{}        `xml:",omitempty"`
}

//SOAPFaultResponse struct
type SOAPFaultResponse struct {
	XMLName xml.Name `xml:"Fault"`
	Code    string   `xml:"faultcode,omitempty"`
	String  string   `xml:"faultstring,omitempty"`
	Actor   string   `xml:"faultactor,omitempty"`
	Detail  string   `xml:"detail,omitempty"`
}

//SOAPEnvelope struct
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	Tag1    string   `xml:"xmlns:soap,attr"`
	Tag2    string   `xml:"xmlns:ns1,attr,omitempty"`
	Body    SOAPBody
}

//SOAPHeader struct
type SOAPHeader struct {
	XMLName xml.Name `xml:"soap:Header"`
	Header  interface{}
}

//SOAPBody struct
type SOAPBody struct {
	XMLName xml.Name    `xml:"soap:Body"`
	Fault   *SOAPFault  `xml:",omitempty"`
	Content interface{} `xml:",omitempty"`
}

//SOAPFault struct
type SOAPFault struct {
	XMLName xml.Name `xml:"soap:Fault"`
	Code    string   `xml:"faultcode,omitempty"`
	String  string   `xml:"faultstring,omitempty"`
	Actor   string   `xml:"faultactor,omitempty"`
	Detail  string   `xml:"detail,omitempty"`
}

//BasicAuth struct
type BasicAuth struct {
	Login    string
	Password string
}

//SOAPClient struct
type SOAPClient struct {
	url  string
	tls  bool
	auth *BasicAuth
}

//UnmarshalXML unmarshal the XML struct
func (b *SOAPBodyResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if b.Content == nil {
		return xml.UnmarshalError("Content must be a pointer to a struct")
	}

	var (
		token    xml.Token
		err      error
		consumed bool
	)

Loop:
	for {
		if token, err = d.Token(); err != nil {
			return err
		}

		if token == nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			if consumed {
				return xml.UnmarshalError("Found multiple elements inside SOAP body; not wrapped-document/literal WS-I compliant")
			} else if se.Name.Space == "http://schemas.xmlsoap.org/soap/envelope/" && se.Name.Local == "Fault" {
				b.Fault = &SOAPFaultResponse{}
				b.Content = nil

				err = d.DecodeElement(b.Fault, &se)
				if err != nil {
					return err
				}

				consumed = true
			} else {
				if err = d.DecodeElement(b.Content, &se); err != nil {
					return err
				}

				consumed = true
			}
		case xml.EndElement:
			break Loop
		}
	}

	return nil
}

//Error returns the SOAPFault error
func (f *SOAPFault) Error() string {
	return f.String
}

//Error returns the SOAPFaultResponse error
func (f *SOAPFaultResponse) Error() string {
	return f.String
}

//NewSOAPClient creates a new SOAP client
func NewSOAPClient(url string, tls bool, auth *BasicAuth) *SOAPClient {
	return &SOAPClient{
		url:  url,
		tls:  tls,
		auth: auth,
	}
}

//Call call SOAP method
func (s *SOAPClient) Call(soapAction string, request, response interface{}) error {
	envelope := SOAPEnvelope{
		Tag1: "http://schemas.xmlsoap.org/soap/envelope/",
		Tag2: "http://service.logisticareversa.correios.com.br/",
	}

	envelope.Body.Content = request
	buffer := new(bytes.Buffer)

	encoder := xml.NewEncoder(buffer)

	if err := encoder.Encode(envelope); err != nil {
		return err
	}

	if err := encoder.Flush(); err != nil {
		return err
	}

	log.Println(buffer.String())

	req, err := http.NewRequest("POST", s.url, buffer)
	if err != nil {
		return err
	}
	if s.auth != nil {
		req.SetBasicAuth(s.auth.Login, s.auth.Password)
	}

	req.Header.Add("Content-Type", "text/xml; charset=\"ISO-8859-1\"")
	if soapAction != "" {
		req.Header.Add("SOAPAction", soapAction)
	}

	req.Close = true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: s.tls,
		},
		Dial: dialTimeout,
	}

	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//Encode to UTF8
	rInUTF8 := transform.NewReader(res.Body, charmap.ISO8859_1.NewDecoder())
	rawbody, err := ioutil.ReadAll(rInUTF8)
	if err != nil {
		return err
	}
	if len(rawbody) == 0 {
		log.Println("empty response")
		return nil
	}

	log.Println(string(rawbody))
	respEnvelope := new(SOAPEnvelopeResponse)
	respEnvelope.Body = SOAPBodyResponse{Content: response}
	err = xml.Unmarshal(rawbody, respEnvelope)

	if err != nil {
		return err
	}

	fault := respEnvelope.Body.Fault
	if fault != nil {
		return fault
	}

	return nil
}
