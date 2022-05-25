package main

type placeOrderRes struct {
	Code    int    `json:"code"`
	OrderID string `json:"order_id"`
}

type cancelOrderRes struct {
	Code    int      `json:"code"`
	Date    string   `json:"date"`
	Success []string `json:"success"`
	Error   []string `json:"error"`
}

type getBalanceRes struct {
	Code int `json:"code"`
	List []struct {
		Currency string `json:"currency"`
		Free     string `json:"free"`
		Total    string `json:"total"`
	} `json:"list"`
}

type getOpenOrdersRes struct {
	Code int `json:"code"`
	Data []struct {
		Symbol         string `json:"symbol"`
		OrderID        string `json:"order_id"`
		CreatedDate    string `json:"created_date"`
		FinishedDate   string `json:"finished_date"`
		Price          string `json:"price"`
		Amount         string `json:"amount"`
		CashAmount     string `json:"cash_amount"`
		ExecutedAmount string `json:"executed_amount"`
		AvgPrice       string `json:"avg_price"`
		Status         string `json:"status"`
		Type           string `json:"type"`
		Kind           string `json:"kind"`
	} `json:"data"`
}
