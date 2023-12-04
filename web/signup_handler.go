package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

type accountCreator interface {
	CreateAccount(context.Context, string, string, []byte) error
}

type SignupReq struct {
	Login    string
	Password string
	Name     string
	PartyID  string
}

type SignupResponse struct {
	Message string
}

func SignUpHandler(logger *slog.Logger, db accountCreator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		req := SignupReq{}

		err = json.Unmarshal(body, &req)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		err = db.CreateAccount(ctx, req.Name, req.Login, hashedPassword)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		resp := SignupResponse{Message: fmt.Sprintf("Successfully signed up user %s with login %s", req.Name, req.Login)}

		respBody, err := json.Marshal(resp)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		_, err = w.Write(respBody)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}
	})
}
