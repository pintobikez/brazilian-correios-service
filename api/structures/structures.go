package structures

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusGenerated  = "generated"
	StatusCanceled   = "canceled"
	StatusUsed       = "used"
	StatusExpired    = "expired"
	StatusError      = "error"
)

type Search struct {
	Where      []*SearchWhere `json:"where"`
	OrderField string         `json:"order_by_field"`
	OrderType  string         `json:"order_by_type"`
	From       int            `json:"from"`
	Offset     int            `json:"offset"`
}

type SearchWhere struct {
	Field    string `json:"field"`
	Value    string `json:"value"`
	Operator string `json:"operator"`
}

type RequestResponse struct {
	RequestID    int64  `json:"request_id"`
	PostageCode  string `json:"postage_code"`
	TrackingCode string `json:"tracking_code"`
	Status       string `json:"status"`
	Callback     string `json:"-"`
}

type Request struct {
	RequestID              int64          `json:"request_id"`
	OrderNr                int64          `json:"order_nr"`
	RequestType            string         `json:"request_type"`
	RequestService         string         `json:"request_service"`
	ColectDate             string         `json:"colect_date"`
	SlipNumber             string         `json:"slip_number"`
	OriginNome             string         `json:"origin_nome"`
	OriginLogradouro       string         `json:"origin_logradouro"`
	OriginNumero           int64          `json:"origin_numero"`
	OriginComplemento      string         `json:"origin_complemento,omitempty"`
	OriginCep              string         `json:"origin_cep"`
	OriginBairro           string         `json:"origin_bairro"`
	OriginCidade           string         `json:"origin_cidade"`
	OriginUf               string         `json:"origin_uf"`
	OriginReferencia       string         `json:"origin_referencia,omitempty"`
	OriginEmail            string         `json:"origin_email"`
	OriginDdd              string         `json:"origin_ddd"`
	OriginTelefone         string         `json:"origin_telefone"`
	DestinationNome        string         `json:"destination_nome"`
	DestinationLogradouro  string         `json:"destination_logradouro"`
	DestinationNumero      int64          `json:"destination_numero"`
	DestinationComplemento string         `json:"destination_complemento,omitempty"`
	DestinationCep         string         `json:"destination_cep"`
	DestinationBairro      string         `json:"destination_bairro"`
	DestinationCidade      string         `json:"destination_cidade"`
	DestinationUf          string         `json:"destination_uf"`
	DestinationReferencia  string         `json:"destination_referencia,omitempty"`
	DestinationEmail       string         `json:"destination_email"`
	Status                 string         `json:"status,omitempty"`
	ErrorMessage           string         `json:"error_message,omitempty"`
	Retries                int64          `json:"retries,omitempty"`
	PostageCode            string         `json:"postage_code,omitempty"`
	TrackingCode           string         `json:"tracking_code,omitempty"`
	CreatedAt              string         `json:"created_at,omitempty"`
	UpdatedAt              string         `json:"updated_at,omitempty"`
	Callback               string         `json:"callback"`
	Items                  []*RequestItem `json:"items"`
}

type RequestItem struct {
	RequestItemID int64  `json:"request_item_id"`
	FkRequestID   int64  `json:"fk_request_id"`
	Item          string `json:"item"`
	ProductName   string `json:"product_name"`
}

type Tracking struct {
	TrackingType string   `json:"tracking_type"`
	Callback     string   `json:"callback"`
	Language     string   `json:"language"`
	Objects      []string `json:"objects,omitempty"`
}

type TrackingResponse struct {
	Items []*TrackingHeader `json:"items,omitempty"`
}

type TrackingHeader struct {
	Object   string            `json:"object"`
	Error    string            `json:"error,omitempty"`
	Name     string            `json:"name,omitempty"`
	Category string            `json:"category,omitempty"`
	Events   []*TrackingEvents `json:"events,omitempty"`
}

type TrackingEvents struct {
	Type        string `json:"type,omitempty"`
	StatusCode  string `json:"status,omitempty"`
	DateTime    string `json:"date,omitempty"`
	Description string `json:"description,omitempty"`
	Details     string `json:"details,omitempty"`
	CTECorreios string `json:"responsible_unit,omitempty"`
}
