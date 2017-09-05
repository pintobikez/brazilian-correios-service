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

type AcompanharPedidoPorData struct {
	XMLName xml.Name `xml:"ns1:acompanharPedidoPorData"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	TipoSolicitacao   string `xml:"tipoSolicitacao,omitempty"`
	Data              string `xml:"data,omitempty"`
}

type AcompanharPedidoPorDataResponse struct {
	XMLName xml.Name `xml:"acompanharPedidoPorDataResponse"`

	AcompanharPedidoPorData *RetornoAcompanhamento `xml:"acompanharPedidoPorData,omitempty"`
}

type RetornoAcompanhamento struct {
	Codigoadministrativo string                `xml:"codigo_administrativo,omitempty"`
	Tiposolicitacao      string                `xml:"tipo_solicitacao,omitempty"`
	Coleta               []*ColetasSolicitadas `xml:"coleta,omitempty"`
	Data                 string                `xml:"data,omitempty"`
	Hora                 string                `xml:"hora,omitempty"`
	Coderro              string                `xml:"cod_erro,omitempty"`
	Msgerro              string                `xml:"msg_erro,omitempty"`
}

type ColetasSolicitadas struct {
	Numeropedido    int                `xml:"numero_pedido,omitempty"`
	Controlecliente string             `xml:"controle_cliente,omitempty"`
	Historico       []*HistoricoColeta `xml:"historico,omitempty"`
	Objeto          []*ObjetoPostal    `xml:"objeto,omitempty"`
}

type HistoricoColeta struct {
	Status          int    `xml:"status,omitempty"`
	Descricaostatus string `xml:"descricao_status,omitempty"`
	Dataatualizacao string `xml:"data_atualizacao,omitempty"`
	Horaatualizacao string `xml:"hora_atualizacao,omitempty"`
	Observacao      string `xml:"observacao,omitempty"`
}

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

type WebServiceFaultInfo struct {
	Mensagem      string `xml:"mensagem,omitempty"`
	Excecao       string `xml:"excecao,omitempty"`
	Classificacao string `xml:"classificacao,omitempty"`
	Causa         string `xml:"causa,omitempty"`
	StackTrace    string `xml:"stackTrace,omitempty"`
}

type SolicitarPostagemReversa struct {
	XMLName xml.Name `xml:"ns1:solicitarPostagemReversa"`

	CodAdministrativo  string           `xml:"codAdministrativo,omitempty"`
	Codigoservico      string           `xml:"codigo_servico,omitempty"`
	Cartao             string           `xml:"cartao,omitempty"`
	Destinatario       *Pessoa          `xml:"destinatario,omitempty"`
	Coletassolicitadas []*ColetaReversa `xml:"coletas_solicitadas,omitempty"`
}

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

type ColetaReversa struct {
	*Coleta

	Numero           int       `xml:"numero,omitempty"`
	Ag               string    `xml:"ag,omitempty"`
	Cartao           string    `xml:"cartao,omitempty"`
	Servicoadicional string    `xml:"servico_adicional,omitempty"`
	Ar               int       `xml:"ar,omitempty"`
	Objcol           []*Objeto `xml:"obj_col,omitempty"`
}

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

type Objeto struct {
	Item    string `xml:"item,omitempty"`
	Desc    string `xml:"desc,omitempty"`
	Entrega string `xml:"entrega,omitempty"`
	Num     string `xml:"num,omitempty"`
	ID      string `xml:"id,omitempty"`
}

type Remetente struct {
	*Pessoa

	Identificacao string `xml:"identificacao,omitempty"`
	Dddcelular    string `xml:"ddd_celular,omitempty"`
	Celular       string `xml:"celular,omitempty"`
	Sms           string `xml:"sms,omitempty"`
}

type Produto struct {
	Codigo string `xml:"codigo,omitempty"`
	Tipo   string `xml:"tipo,omitempty"`
	Qtd    string `xml:"qtd,omitempty"`
}

type SolicitarPostagemReversaResponse struct {
	XMLName xml.Name `xml:"solicitarPostagemReversaResponse"`

	SolicitarPostagemReversa *RetornoPostagem `xml:"solicitarPostagemReversa,omitempty"`
}

type RetornoPostagem struct {
	Statusprocessamento  string                  `xml:"status_processamento,omitempty"`
	Dataprocessamento    string                  `xml:"data_processamento,omitempty"`
	Horaprocessamento    string                  `xml:"hora_processamento,omitempty"`
	Coderro              string                  `xml:"cod_erro,omitempty"`
	Msgerro              string                  `xml:"msg_erro,omitempty"`
	Resultadosolicitacao []*ResultadoSolicitacao `xml:"resultado_solicitacao,omitempty"`
}

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

type ValidarPostagemSimultanea struct {
	XMLName xml.Name `xml:"ns1:validarPostagemSimultanea"`

	CodAdministrativo string            `xml:"codAdministrativo,omitempty"`
	Codigoservico     string            `xml:"codigo_servico,omitempty"`
	Cartao            string            `xml:"cartao,omitempty"`
	Cepdestinatario   string            `xml:"cep_destinatario,omitempty"`
	Coleta            *ColetaSimultanea `xml:"coleta,omitempty"`
}

type ColetaSimultanea struct {
	*Coleta

	Obs string `xml:"obs,omitempty"`
	Obj string `xml:"obj,omitempty"`
}

type ValidarPostagemSimultaneaResponse struct {
	XMLName xml.Name `xml:"validarPostagemSimultaneaResponse"`

	ValidarPostagemSimultanea *RetornoValidacao `xml:"validarPostagemSimultanea,omitempty"`
}

type RetornoValidacao struct {
	Coderro int    `xml:"cod_erro,omitempty"`
	Msgerro string `xml:"msg_erro,omitempty"`
}

type AcompanharPedido struct {
	XMLName xml.Name `xml:"ns1:acompanharPedido"`

	CodAdministrativo string   `xml:"codAdministrativo,omitempty"`
	TipoBusca         string   `xml:"tipoBusca,omitempty"`
	TipoSolicitacao   string   `xml:"tipoSolicitacao,omitempty"`
	NumeroPedido      []string `xml:"numeroPedido,omitempty"`
}

type AcompanharPedidoResponse struct {
	XMLName xml.Name `xml:"acompanharPedidoResponse"`

	AcompanharPedido *RetornoAcompanhamento `xml:"acompanharPedido,omitempty"`
}

type RevalidarPrazoAutorizacaoPostagem struct {
	XMLName xml.Name `xml:"ns1:revalidarPrazoAutorizacaoPostagem"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	NumeroPedido      string `xml:"numeroPedido,omitempty"`
	QtdeDias          string `xml:"qtdeDias,omitempty"`
}

type RevalidarPrazoAutorizacaoPostagemResponse struct {
	XMLName xml.Name `xml:"revalidarPrazoAutorizacaoPostagemResponse"`

	RevalidarPrazoAutorizacaoPostagem *RetornoRevalidarPrazo `xml:"revalidarPrazoAutorizacaoPostagem,omitempty"`
}

type RetornoRevalidarPrazo struct {
	Numeropedido string `xml:"numero_pedido,omitempty"`
	Prazo        string `xml:"prazo,omitempty"`
	Coderro      string `xml:"cod_erro,omitempty"`
	Msgerro      string `xml:"msg_erro,omitempty"`
}

type SobreWebService struct {
	XMLName xml.Name `xml:"ns1:sobreWebService"`
}

type SobreWebServiceResponse struct {
	XMLName xml.Name `xml:"sobreWebServiceResponse"`

	SobreWebService *RetornoSobreWebService `xml:"sobreWebService,omitempty"`
}

type RetornoSobreWebService struct {
	Data            string `xml:"data,omitempty"`
	Hora            string `xml:"hora,omitempty"`
	Coderro         string `xml:"cod_erro,omitempty"`
	Msgerro         string `xml:"msg_erro,omitempty"`
	Versao          string `xml:"versao,omitempty"`
	DataHomologacao string `xml:"dataHomologacao,omitempty"`
	DataProducao    string `xml:"dataProducao,omitempty"`
	Fase            string `xml:"fase,omitempty"`
	Resumo          string `xml:"resumo,omitempty"`
}

type CancelarPedido struct {
	XMLName xml.Name `xml:"ns1:cancelarPedido"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	NumeroPedido      string `xml:"numeroPedido,omitempty"`
	Tipo              string `xml:"tipo,omitempty"`
}

type CancelarPedidoResponse struct {
	XMLName xml.Name `xml:"cancelarPedidoResponse"`

	CancelarPedido *RetornoCancelamento `xml:"cancelarPedido,omitempty"`
}

type RetornoCancelamento struct {
	Codigoadministrativo string              `xml:"codigo_administrativo,omitempty"`
	Objetopostal         *ObjetoSimplificado `xml:"objeto_postal,omitempty"`
	Data                 string              `xml:"data,omitempty"`
	Hora                 string              `xml:"hora,omitempty"`
	Coderro              string              `xml:"cod_erro,omitempty"`
	Msgerro              string              `xml:"msg_erro,omitempty"`
}

type ObjetoSimplificado struct {
	Numeropedido         int    `xml:"numero_pedido,omitempty"`
	Statuspedido         string `xml:"status_pedido,omitempty"`
	Datahoracancelamento string `xml:"datahora_cancelamento,omitempty"`
}

type SolicitarPostagemSimultanea struct {
	XMLName xml.Name `xml:"ns1:solicitarPostagemSimultanea"`

	CodAdministrativo  string              `xml:"codAdministrativo,omitempty"`
	Codigoservico      string              `xml:"codigo_servico,omitempty"`
	Cartao             string              `xml:"cartao,omitempty"`
	Destinatario       *Pessoa             `xml:"destinatario,omitempty"`
	Coletassolicitadas []*ColetaSimultanea `xml:"coletas_solicitadas,omitempty"`
}

type SolicitarPostagemSimultaneaResponse struct {
	XMLName xml.Name `xml:"solicitarPostagemSimultaneaResponse"`

	SolicitarPostagemSimultanea *RetornoPostagem `xml:"solicitarPostagemSimultanea,omitempty"`
}

type ValidarPostagemReversa struct {
	XMLName xml.Name `xml:"ns1:validarPostagemReversa"`

	CodAdministrativo string         `xml:"codAdministrativo,omitempty"`
	Codigoservico     string         `xml:"codigo_servico,omitempty"`
	Cartao            string         `xml:"cartao,omitempty"`
	Cepdestinatario   string         `xml:"cep_destinatario,omitempty"`
	Coleta            *ColetaReversa `xml:"coleta,omitempty"`
}

type ValidarPostagemReversaResponse struct {
	XMLName xml.Name `xml:"validarPostagemReversaResponse"`

	ValidarPostagemReversa *RetornoValidacao `xml:"validarPostagemReversa,omitempty"`
}

type CalcularDigitoVerificador struct {
	XMLName xml.Name `xml:"ns1:calcularDigitoVerificador"`

	Numero string `xml:"numero,omitempty"`
}

type CalcularDigitoVerificadorResponse struct {
	XMLName xml.Name `xml:"calcularDigitoVerificadorResponse"`

	CalcularDigitoVerificador *RetornoDigitoVerificador `xml:"calcularDigitoVerificador,omitempty"`
}

type RetornoDigitoVerificador struct {
	Data    string `xml:"data,omitempty"`
	Hora    string `xml:"hora,omitempty"`
	Coderro string `xml:"cod_erro,omitempty"`
	Msgerro string `xml:"msg_erro,omitempty"`
	Digito  int    `xml:"digito,omitempty"`
	Numero  int    `xml:"numero,omitempty"`
}

type SolicitarRange struct {
	XMLName xml.Name `xml:"ns1:solicitarRange"`

	CodAdministrativo string `xml:"codAdministrativo,omitempty"`
	Tipo              string `xml:"tipo,omitempty"`
	Servico           string `xml:"servico,omitempty"`
	Quantidade        string `xml:"quantidade,omitempty"`
}

type SolicitarRangeResponse struct {
	XMLName xml.Name `xml:"solicitarRangeResponse"`

	SolicitarRange *RetornoFaixaNumerica `xml:"solicitarRange,omitempty"`
}

type RetornoFaixaNumerica struct {
	Data         string `xml:"data,omitempty"`
	Hora         string `xml:"hora,omitempty"`
	Coderro      string `xml:"cod_erro,omitempty"`
	Msgerro      string `xml:"msg_erro,omitempty"`
	Faixainicial int    `xml:"faixa_inicial,omitempty"`
	Faixafinal   int    `xml:"faixa_final,omitempty"`
}

type LogisticaReversaWS struct {
	client *SOAPClient
}

func NewLogisticaReversaWS(url string, tls bool, auth *BasicAuth) *LogisticaReversaWS {
	if url == "" {
		url = ""
	}
	client := NewSOAPClient(url, tls, auth)

	return &LogisticaReversaWS{
		client: client,
	}
}

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

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) ValidarPostagemSimultanea(request *ValidarPostagemSimultanea) (*ValidarPostagemSimultaneaResponse, error) {
	response := new(ValidarPostagemSimultaneaResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) AcompanharPedido(request *AcompanharPedido) (*AcompanharPedidoResponse, error) {
	response := new(AcompanharPedidoResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) RevalidarPrazoAutorizacaoPostagem(request *RevalidarPrazoAutorizacaoPostagem) (*RevalidarPrazoAutorizacaoPostagemResponse, error) {
	response := new(RevalidarPrazoAutorizacaoPostagemResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) SobreWebService(request *SobreWebService) (*SobreWebServiceResponse, error) {
	response := new(SobreWebServiceResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

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

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) SolicitarPostagemSimultanea(request *SolicitarPostagemSimultanea) (*SolicitarPostagemSimultaneaResponse, error) {
	response := new(SolicitarPostagemSimultaneaResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) ValidarPostagemReversa(request *ValidarPostagemReversa) (*ValidarPostagemReversaResponse, error) {
	response := new(ValidarPostagemReversaResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) CalcularDigitoVerificador(request *CalcularDigitoVerificador) (*CalcularDigitoVerificadorResponse, error) {
	response := new(CalcularDigitoVerificadorResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *LogisticaReversaWS) SolicitarRange(request *SolicitarRange) (*SolicitarRangeResponse, error) {
	response := new(SolicitarRangeResponse)
	err := service.client.Call("", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

var timeout = time.Duration(30 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

type SOAPEnvelopeResponse struct {
	XMLName xml.Name `xml:"Envelope"`

	Body SOAPBodyResponse
}

type SOAPBodyResponse struct {
	XMLName xml.Name `xml:"Body"`

	Fault   *SOAPFaultResponse `xml:",omitempty"`
	Content interface{}        `xml:",omitempty"`
}

type SOAPFaultResponse struct {
	XMLName xml.Name `xml:"Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

type SOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	Tag1    string   `xml:"xmlns:soap,attr"`
	Tag2    string   `xml:"xmlns:ns1,attr,omitempty"`

	Body SOAPBody
}

type SOAPHeader struct {
	XMLName xml.Name `xml:"soap:Header"`

	Header interface{}
}

type SOAPBody struct {
	XMLName xml.Name `xml:"soap:Body"`

	Fault   *SOAPFault  `xml:",omitempty"`
	Content interface{} `xml:",omitempty"`
}

type SOAPFault struct {
	XMLName xml.Name `xml:"soap:Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

type BasicAuth struct {
	Login    string
	Password string
}

type SOAPClient struct {
	url  string
	tls  bool
	auth *BasicAuth
}

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

func (f *SOAPFault) Error() string {
	return f.String
}
func (f *SOAPFaultResponse) Error() string {
	return f.String
}

func NewSOAPClient(url string, tls bool, auth *BasicAuth) *SOAPClient {
	return &SOAPClient{
		url:  url,
		tls:  tls,
		auth: auth,
	}
}

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
