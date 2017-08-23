package soaptracking

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	// "golang.org/x/text/encoding/charmap"
	// "golang.org/x/text/transform"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// against "unused imports"
var _ time.Time
var _ xml.Name

type BuscaEventosLista struct {
	XMLName xml.Name `xml:"res:buscaEventosLista"`

	User     string   `xml:"usuario,omitempty"`
	Password string   `xml:"senha,omitempty"`
	Type     string   `xml:"tipo,omitempty"`
	Result   string   `xml:"resultado,omitempty"`
	Language string   `xml:"lingua,omitempty"`
	Objects  []string `xml:"objetos,omitempty"`
}

type BuscaEventosListaResponse struct {
	XMLName xml.Name `xml:"buscaEventosListaResponse"`

	Result *Return `xml:"return,omitempty"`
}

type Return struct {
	Version  string    `xml:"versao,omitempty"`
	Quantity int       `xml:"qtd,omitempty"`
	Objects  []*Objeto `xml:"objeto,omitempty"`
}

type Objeto struct {
	TrackingCode string    `xml:"numero,omitempty"`
	Error        string    `xml:"erro,omitempty"`
	Initials     string    `xml:"sigla,omitempty"`
	Name         string    `xml:"nome,omitempty"`
	Category     string    `xml:"categoria,omitempty"`
	Events       []*Evento `xml:"evento,omitempty"`
}

type Evento struct {
	Type        string `xml:"tipo,omitempty"`
	StatusCode  string `xml:"status,omitempty"`
	Date        string `xml:"data,omitempty"`
	Hour        string `xml:"hora,omitempty"`
	Description string `xml:"descricao,omitempty"`
	Detail      string `xml:"detalhe,omitempty"`
	Local       string `xml:"local,omitempty"`
	Code        string `xml:"codigo,omitempty"`
	City        string `xml:"cidade,omitempty"`
	FiscalUnit  string `xml:"uf,omitempty"`
}

type RastroWS struct {
	client *SOAPClient
}

func NewRastroWS(url string, tls bool) *RastroWS {
	if url == "" {
		url = ""
	}
	client := NewSOAPClient(url, tls)

	return &RastroWS{
		client: client,
	}
}

// Error can be either of the following types:
//
//   - ComponenteException

func (service *RastroWS) BuscaEventosLista(request *BuscaEventosLista) (*BuscaEventosListaResponse, error) {
	response := new(BuscaEventosListaResponse)
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
	XMLName xml.Name `xml:"soapenv:Envelope"`
	Tag1    string   `xml:"xmlns:soapenv,attr"`
	Tag2    string   `xml:"xmlns:res,attr,omitempty"`

	Body SOAPBody
}

type SOAPHeader struct {
	XMLName xml.Name `xml:"soapenv:Header"`

	Header interface{}
}

type SOAPBody struct {
	XMLName xml.Name `xml:"soapenv:Body"`

	Fault   *SOAPFault  `xml:",omitempty"`
	Content interface{} `xml:",omitempty"`
}

type SOAPFault struct {
	XMLName xml.Name `xml:"soapenv:Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}
type SOAPClient struct {
	url string
	tls bool
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

func NewSOAPClient(url string, tls bool) *SOAPClient {
	return &SOAPClient{
		url: url,
		tls: tls,
	}
}

func (s *SOAPClient) Call(soapAction string, request, response interface{}) error {
	envelope := SOAPEnvelope{
		Tag1: "http://schemas.xmlsoap.org/soap/envelope/",
		Tag2: "http://resource.webservice.correios.com.br/",
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

	req, err := http.NewRequest("POST", s.url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "text/xml; charset=\"utf-8\"")
	if soapAction != "" {
		req.Header.Add("SOAPAction", soapAction)
	}

	req.Header.Set("User-Agent", "gowsdl/0.1")
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

	rawbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if len(rawbody) == 0 {
		return nil
	}

	respEnvelope := new(SOAPEnvelopeResponse)
	respEnvelope.Body = SOAPBodyResponse{Content: response}
	err = xml.Unmarshal(rawbody, respEnvelope)
	if err != nil {
		return err
	}

	/*req.Header.Add("Content-Type", "text/xml; charset=\"ISO-8859-1\"")
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
		return nil
	}

	respEnvelope := new(SOAPEnvelopeResponse)
	respEnvelope.Body = SOAPBodyResponse{Content: response}
	err = xml.Unmarshal(rawbody, respEnvelope)

	if err != nil {
		log.Println(err.Error())
		return err
	}*/

	fault := respEnvelope.Body.Fault
	if fault != nil {
		return fault
	}

	return nil
}
