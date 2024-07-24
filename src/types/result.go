package types

import "fmt"

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResultData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResultHealth struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type ResultError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (r Result) String() string {
	return fmt.Sprintf(`{"code":%d,"message":"%s"}`, r.Code, r.Message)
}

func (rd ResultData) String() string {
	return fmt.Sprintf(`{"code":%d,"message":"%s","data":%s}`, rd.Code, rd.Message, rd.Data)
}

func (rh ResultHealth) String() string {
	return fmt.Sprintf(`{"code":%d,"message":"%s","status":"%s"}`, rh.Code, rh.Message, rh.Status)
}

func (re ResultError) String() string {
	return fmt.Sprintf(`{"code":%d,"message":"%s","error":"%s"}`, re.Code, re.Message, re.Error)
}
