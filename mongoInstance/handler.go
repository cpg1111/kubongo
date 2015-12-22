package mongoInstance

import (
	"net/http"
)

type mongoHandler struct {
	http.Handler
}

func NewHandler() *mongoHandler{
	return &mongoHandler{}
}

func (m *mongoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		m.Get(res, req)
	case "POST":
		m.Post(res, req)
	case "PUT":
		m.Put(res, req)
	case "DELETE":
		m.Delete(res, req)
	}
}

func (m *mongoHandler) Get(res http.ResponseWriter, req *http.Request) {

}

func (m *mongoHandler) Post(res http.ResponseWriter, req *http.Request) {

}

func (m *mongoHandler) Put(res http.ResponseWriter, req *http.Request) {

}

func (m *mongoHandler) Delete(res http.ResponseWriter, req *http.Request) {

}
