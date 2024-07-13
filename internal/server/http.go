package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Address struct {
	port     string
	hostname string
}

func (a Address) String() string {
	return a.hostname + ":" + a.port
}

func NewAddress(hostname, port string) Address {
	return Address{hostname: hostname, port: port}
}

type API struct {
	address Address
	router  *http.ServeMux
	log     *Log
}

func NewAPIServer(address Address) *API {
	api := &API{
		address: address, router: http.NewServeMux(),
		log: NewLog(),
	}
	api.InitializeRoutes()
	return api
}

func (a *API) InitializeRoutes() {
	a.router.HandleFunc("POST /", a.handleProduce)
	a.router.HandleFunc("GET /{offset}", a.handleConsume)
}

func (a *API) Start() error {
	return http.ListenAndServe(a.address.String(), a.router)
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record Record `json:"record"`
}

func (a *API) handleConsume(w http.ResponseWriter, r *http.Request) {
	offset := r.PathValue("offset")
	if offset == "" {
		BadRequestError(w, errors.New("missing offset"))
		return
	}
	offsetVal, err := strconv.ParseUint(offset, 10, 64)
	if err != nil {
		BadRequestError(w, err)
		return
	}
	record, err := a.log.Read(offsetVal)
	if err != nil {
		InternalServerError(w, err)
		return

	}
	response := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		InternalServerError(w, err)
		return
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

func (a *API) handleProduce(w http.ResponseWriter, r *http.Request) {
	var req ProduceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		BadRequestError(w, err)
		return
	}
	offset, err := a.log.Append(req.Record)
	if err != nil {
		InternalServerError(w, err)
		return
	}
	response := ProduceResponse{Offset: offset}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		InternalServerError(w, err)
		return
	}
}
