package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	s "github.com/pintobikez/brazilian-correios-service/api/structures"
	"strconv"
)

const DefaultOffset = 50

type Repository struct {
	//props
	db *sql.DB
}

// Connects to the mysql database
func (r *Repository) ConnectDB(stringConn string) error {
	var err error
	r.db, err = sql.Open("mysql", stringConn)
	if err != nil {
		return err
	}
	return nil
}

// Disconnects from the mysql database
func (r *Repository) DisconnectDB() {
	r.db.Close()
}

func (r *Repository) InsertRequest(o *s.Request) error {

	stmt, err := r.db.Prepare("INSERT INTO `request` VALUES (null,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'',0,'','',now(),now())")
	if err != nil {
		return fmt.Errorf("Error in insert request prepared statement: %s", err.Error())
	}

	res, err := stmt.Exec(o.RequestType, o.RequestService, o.ColectDate, o.OrderNr, o.SlipNumber, o.OriginNome, o.OriginLogradouro, o.OriginNumero, o.OriginComplemento, o.OriginCep, o.OriginBairro,
		o.OriginCidade, o.OriginUf, o.OriginReferencia, o.OriginEmail, o.OriginDdd, o.OriginTelefone, o.DestinationNome, o.DestinationLogradouro, o.DestinationNumero, o.DestinationComplemento,
		o.DestinationCep, o.DestinationBairro, o.DestinationCidade, o.DestinationUf, o.DestinationReferencia, o.DestinationEmail, o.Callback, s.StatusPending)

	if err != nil {
		return fmt.Errorf("Error in insert request: %d %s", o.OrderNr, err.Error())
	}
	o.RequestID, _ = res.LastInsertId()
	stmt.Close()

	// INSERT THE ITEMS NOW
	for _, i := range o.Items {
		stmt, err := r.db.Prepare("INSERT INTO `request_item` VALUES (null,?,?,?)")
		if err != nil {
			return fmt.Errorf("Error in insert request_item prepared statement: %s", err.Error())
		}

		res, err := stmt.Exec(o.RequestID, i.Item, i.ProductName)
		if err != nil {
			return fmt.Errorf("Error in insert request item: %d %s: %s", o.OrderNr, i.Item, err.Error())
		}
		i.RequestItemID, _ = res.LastInsertId()
	}

	defer stmt.Close()

	return nil
}

// Updates an RequestItems PostageCode
func (r *Repository) UpdateRequest(o *s.Request) error {

	stmt, err := r.db.Prepare("UPDATE `request` SET request_type=?, request_service=?, colect_date=?, origin_nome=?, origin_logradouro=?, origin_numero=?, origin_complemento=?, " +
		"origin_cep=?, origin_bairro=?, origin_cidade=?, origin_uf=?, origin_referencia=?, origin_email=?, origin_ddd=?, origin_telefone=?, destination_nome=?, destination_logradouro=?, destination_numero=?, " +
		"destination_complemento=?, destination_cep=?, destination_bairro=?, destination_cidade=?, destination_uf=?, destination_referencia=?," +
		" destination_email=?, status=?, error_message=? WHERE fk_request_id=?")

	if err != nil {
		return fmt.Errorf("Error in update request prepared statement: %s", err.Error())
	}
	defer stmt.Close()

	_, err = stmt.Exec(o.RequestType, o.RequestService, o.ColectDate, o.OriginNome, o.OriginLogradouro, o.OriginNumero, o.OriginComplemento, o.OriginCep, o.OriginBairro, o.OriginCidade, o.OriginUf, o.OriginReferencia, o.OriginEmail, o.OriginDdd, o.OriginTelefone,
		o.DestinationNome, o.DestinationLogradouro, o.DestinationNumero, o.DestinationComplemento, o.DestinationCep, o.DestinationBairro, o.DestinationCidade, o.DestinationUf, o.DestinationReferencia, o.DestinationEmail,
		o.Status, o.ErrorMessage, o.RequestID)
	if err != nil {
		return fmt.Errorf("Could not update Request %d", o.RequestID)
	}

	return nil
}

// Updates an RequestItems status
func (r *Repository) UpdateRequestStatus(o *s.Request, status string, message string) (int64, error) {

	stmt, err := r.db.Prepare("UPDATE `request` SET retries=?, status=?, error_message=? WHERE request_id=?")
	if err != nil {
		return 0, fmt.Errorf("Error in update status prepared statement: %s", err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(o.Retries, status, message, o.RequestID)

	if err != nil {
		return 0, fmt.Errorf("Could not update status for Request %d", o.RequestID)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("Could not update status for Request %d", o.RequestID)
	}

	//update struct
	o.Status = status
	o.ErrorMessage = message

	return affect, nil
}

// Updates an RequestItems PostageCode
func (r *Repository) UpdateRequestPostage(o *s.Request, code string) error {

	stmt, err := r.db.Prepare("UPDATE `request` SET postage_code=?,status=? WHERE request_id=?")
	if err != nil {
		return fmt.Errorf("Error in update postage code prepared statement: %s", err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(code, s.StatusGenerated, o.RequestID)

	if err != nil {
		return fmt.Errorf("Could not update postage code for Request %d", o.RequestID)
	}

	affect, err := res.RowsAffected()
	if err != nil || affect <= 0 {
		return fmt.Errorf("Could not update postage code for Request %d", o.RequestID)
	}

	//update struct
	o.PostageCode = code
	o.Status = s.StatusGenerated

	return nil
}

// Updates an RequestItems PostageCode
func (r *Repository) UpdateRequestTracking(o *s.Request, code string) error {

	stmt, err := r.db.Prepare("UPDATE `request` SET tracking_code=?,status=? WHERE request_id=?")
	if err != nil {
		return fmt.Errorf("Error in update postage code prepared statement: %s", err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(code, s.StatusUsed, o.RequestID)

	if err != nil {
		return fmt.Errorf("Could not update postage code for Request %d", o.RequestID)
	}

	affect, err := res.RowsAffected()
	if err != nil || affect <= 0 {
		return fmt.Errorf("Could not update postage code for Request %d", o.RequestID)
	}

	//update struct
	o.PostageCode = code
	o.Status = s.StatusUsed

	return nil
}

// Finds Request by RequestID
func (r *Repository) FindRequestByID(requestID int64) (bool, error) {
	exists := false

	rows, err := r.db.Query("SELECT EXISTS(SELECT * as counter FROM `request` request_id=?) as exist", requestID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	err = rows.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Gets Request info by RequestID
func (r *Repository) GetRequestByID(requestID int) (*s.Request, error) {
	var resp = new(s.Request)

	rows, err := r.db.Query("SELECT * FROM `request` as o INNER JOIN `request_item` as items on items.fk_request_id=o.request_id WHERE o.request_id=?", requestID)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	err = r.processRows(resp, rows)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Gets Request info by PostageCode
func (r *Repository) GetRequestByPostageCode(code string) (*s.Request, error) {
	var resp = new(s.Request)

	rows, err := r.db.Query("SELECT * FROM `request` as o INNER JOIN `request_item` as oi on oi.fk_request_id=o.request_id WHERE oi.postage_code=?", code)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	err = r.processRows(resp, rows)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Gets Request info by ?
func (r *Repository) GetRequestBy(req *s.Search) ([]*s.Request, error) {
	var resp = []*s.Request{}

	query := "SELECT * FROM `request` as o INNER JOIN `request_item` as items on items.fk_request_id=o.request_id %s ORDER BY %s %s LIMIT %d,%d"
	q := ""

	if len(req.Where) > 0 {
		for _, e := range req.Where {
			if e.Operator == "LIKE" {
				e.Value = "%" + e.Value + "%"
			}
			if e.Operator == "IN" || e.Operator == "NOT IN" {
				e.Value = "(" + e.Value + ")"
			}

			// build the where bits
			if e.Field != "" && e.Value != "" && e.Operator != "" {
				// check if search value is not int
				if _, err := strconv.Atoi(e.Value); err != nil {
					e.Value = "'" + e.Value + "'"
				}

				if q == "" {
					q = "WHERE " + e.Field + " " + e.Operator + " " + e.Value
				} else {
					q += " AND " + e.Field + " " + e.Operator + " " + e.Value
				}
			}
		}
	}

	// set the default order field
	if req.OrderField == "" {
		req.OrderField = "request_id"
	}

	// set the default order value
	if req.OrderType == "" {
		req.OrderType = "DESC"
	}

	// set the default offset
	if req.Offset == 0 {
		req.Offset = DefaultOffset
	}

	query = fmt.Sprintf(query, q, req.OrderField, req.OrderType, req.From, req.Offset)
	rows, err := r.db.Query(query)

	if err != nil {
		return resp, err
	}
	defer rows.Close()

	resp, err = r.processMultipleRows(resp, rows)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Processes a Row result into a Request struct
func (r *Repository) processRows(resp *s.Request, rows *sql.Rows) error {
	var arr []*s.RequestItem

	for rows.Next() {
		aux := new(s.RequestItem)

		err := rows.Scan(&resp.RequestID, &resp.RequestType, &resp.RequestService, &resp.ColectDate, &resp.OrderNr, &resp.SlipNumber, &resp.OriginNome, &resp.OriginLogradouro, &resp.OriginNumero, &resp.OriginComplemento, &resp.OriginCep, &resp.OriginBairro, &resp.OriginCidade,
			&resp.OriginUf, &resp.OriginReferencia, &resp.OriginEmail, &resp.OriginDdd, &resp.OriginTelefone, &resp.DestinationNome, &resp.DestinationLogradouro, &resp.DestinationNumero, &resp.DestinationComplemento,
			&resp.DestinationCep, &resp.DestinationBairro, &resp.DestinationCidade, &resp.DestinationUf, &resp.DestinationReferencia, &resp.DestinationEmail, &resp.Callback, &resp.Status, &resp.ErrorMessage,
			&resp.Retries, &resp.PostageCode, &resp.TrackingCode, &resp.CreatedAt, &resp.UpdatedAt, &aux.RequestItemID, &aux.FkRequestID, &aux.Item, &aux.ProductName)

		if err != nil {
			return fmt.Errorf("Error reading rows: %s", err.Error())
		}

		arr = append(arr, aux)
		resp.Items = arr
	}

	return nil
}

// Processes a Row result into a Request struct
func (r *Repository) processMultipleRows(resp []*s.Request, rows *sql.Rows) ([]*s.Request, error) {
	var (
		arr    []*s.RequestItem
		prevID int64
	)

	for rows.Next() {
		req := new(s.Request)
		aux := new(s.RequestItem)

		err := rows.Scan(&req.RequestID, &req.RequestType, &req.RequestService, &req.ColectDate, &req.OrderNr, &req.SlipNumber, &req.OriginNome, &req.OriginLogradouro, &req.OriginNumero, &req.OriginComplemento, &req.OriginCep, &req.OriginBairro, &req.OriginCidade,
			&req.OriginUf, &req.OriginReferencia, &req.OriginEmail, &req.OriginDdd, &req.OriginTelefone, &req.DestinationNome, &req.DestinationLogradouro, &req.DestinationNumero, &req.DestinationComplemento,
			&req.DestinationCep, &req.DestinationBairro, &req.DestinationCidade, &req.DestinationUf, &req.DestinationReferencia, &req.DestinationEmail, &req.Callback, &req.Status, &req.ErrorMessage,
			&req.Retries, &req.PostageCode, &req.TrackingCode, &req.CreatedAt, &req.UpdatedAt, &aux.RequestItemID, &aux.FkRequestID, &aux.Item, &aux.ProductName)

		if err != nil {
			return resp, fmt.Errorf("Error reading rows: %s", err.Error())
		}

		// new element
		if prevID != req.RequestID {
			//cleanup the items array
			var arr = []*s.RequestItem{}
			arr = append(arr, aux)
			req.Items = arr
			//append the item
			resp = append(resp, req)
		}
		if prevID == req.RequestID {
			//append the items to the previous item
			l := len(resp)
			arr = resp[l-1].Items
			resp[l-1].Items = append(arr, aux)
		}

		prevID = req.RequestID
	}

	return resp, nil
}
