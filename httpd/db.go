package httpd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *Service) initManageHandler() {
	s.router.GET("/api/v1/stats", s.handlerStats)
	s.router.GET("/api/v1/peer", s.handlerPeers)
	s.router.POST("/api/v1/peer", s.handlerJoin)
	s.router.DELETE("/api/v1/peer", s.handlerRemove)
	s.router.GET("/api/v1/db/backup", s.handlerBackup)
	s.router.GET("/api/v1/db/restore", s.handlerRestore)
}

func (s *Service) handlerStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ms := s.cluster.Statistics(nil)
	ReturnJson(w, 200, ms)
}

func (s *Service) handlerPeers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	peers, err := s.cluster.Peers()
	if err != nil {
		ReturnServerError(w, err)
		return
	}
	ReturnJson(w, 200, peers)
}

func (s *Service) handlerJoin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		ReturnBadRequest(w, fmt.Errorf("unmarshal fail"))
		return
	}

	if len(m) != 1 {
		ReturnBadRequest(w, fmt.Errorf("only allow 1 addr to join one time"))
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		ReturnBadRequest(w, fmt.Errorf("ihave no addr to join"))
		return
	}

	if err := s.cluster.Join(remoteAddr); err != nil {
		ReturnServerError(w, err)
		return
	}
}

func (s *Service) handlerRemove(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ReturnBadRequest(w, fmt.Errorf("read body fail"))
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		ReturnBadRequest(w, fmt.Errorf("unmarshal fail"))
		return
	}

	if len(m) != 1 {
		ReturnBadRequest(w, fmt.Errorf("only allow 1 addr to remove one time"))
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		ReturnBadRequest(w, fmt.Errorf("have no addr to join"))
		return
	}

	if err := s.cluster.Remove(remoteAddr); err != nil {
		ReturnServerError(w, err)
		return
	}
}

func (s *Service) handlerBackup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	var data []byte
	if data, err = s.cluster.Backup(); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnByte(w, 200, data)
	}
}

func (s *Service) handlerRestore(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	file := r.FormValue("file")
	var err error
	if err = s.cluster.Restore(file); err != nil {
		ReturnServerError(w, err)
	} else {
		ReturnOK(w, "success")
	}
}
